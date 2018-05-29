package simulator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/tteige/uit-go/autoscale"
	"time"
	"log"
	"html/template"
	"github.com/tteige/uit-go/models"
	"github.com/segmentio/ksuid"
	"encoding/json"
	"net/url"
	"io"
	"github.com/tteige/uit-go/metapipe"
	"sort"
)

type simulationOutput map[int]map[string][]autoscale.AlgorithmJob
type fullSimulationOutput struct {
	Name        string                   `json:"name"`
	Jobs        []autoscale.AlgorithmJob `json:"jobs"`
	SimEvents   []models.SimulatorEvent  `json:"sim_events"`
	CloudEvents []models.CloudEvent      `json:"cloud_events"`
}

type Simulator struct {
	DB          *sql.DB
	Hostname    string
	SimClusters autoscale.ClusterCollection
	Algorithm   autoscale.Algorithm
	Log         *log.Logger
	templates   *template.Template
	tmplLoc     string
	Estimator   autoscale.Estimator
}

type metapipeReturn struct {
	id         string
	input      autoscale.AlgorithmInput
	jobs       []autoscale.AlgorithmJob
	timestamp  time.Time
	err        error
	iterations int
	timestep   int
}

func (sim *Simulator) Run() {
	sim.Log.Printf("Starting the auto scaling simulator at: %s ", sim.Hostname)
	sim.tmplLoc = "simulator/templates/"
	sim.templates = template.Must(template.ParseFiles(sim.tmplLoc+"footer.html", sim.tmplLoc+"header.html",
		sim.tmplLoc+"index.html", sim.tmplLoc+"navbar.html"))
	err := sim.Estimator.Init()
	if err != nil {
		sim.Log.Fatal(err)
		return
	}
	sim.serveSim()
}

func (sim *Simulator) serveSim() {
	r := mux.NewRouter()
	r.HandleFunc("/", sim.indexHandle).Methods("GET")
	r.HandleFunc("/metapipe/simulate/", sim.metapipeSimulationHandle).Methods("POST")
	r.HandleFunc("/metapipe/simulation/", sim.getPreviousScalingHandle).Methods("GET")
	r.HandleFunc("/metapipe/simulation/all", sim.getAllSimulations).Methods("GET")
	http.ListenAndServe(sim.Hostname, r)
}

func (sim *Simulator) getAllSimulations(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("GetAllSimulationsRequest: /simulation/all")
	runs, err := models.GetAllAutoscalingRunStats(sim.DB)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(runs)
	w.Write(b)
}

func (sim *Simulator) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

	raw, err := url.Parse(r.RequestURI)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var out fullSimulationOutput
	q := raw.Query()
	if val, ok := q["id"]; ok {
		sim.Log.Printf("GetAllSimulationsRequest: /metapipe/simulation/?id=%s", q["id"])
		events, err := models.GetAutoscalingRunEvents(sim.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.CloudEvents = events

		simEvents, err := models.GetSimulatorEvents(sim.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.SimEvents = simEvents

		jobs, err := models.GetAllAlgorithmJobs(sim.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.Jobs = jobs
		out.Name = val[0]
	}

	b, err := json.Marshal(out)
	w.Write(b)
}

func (sim *Simulator) simulate(simId string, completeQueue []autoscale.AlgorithmJob, algInput autoscale.AlgorithmInput, algTimestamp time.Time, timestep int, iterations int) (simulationOutput, error) {
	jsonSimQueue := make(simulationOutput)
	sim.Log.Printf("Starting simulation: %s", simId)

	for i := 0; i < iterations; i++ {
		if i > 0 {
			//Simulates a 30 minute interval between each scaling attempt
			algTimestamp = algTimestamp.Add(time.Minute * time.Duration(timestep))
		}

		//Selects jobs that are created before the timestamp
		removedFromCompleted := 0
		for k := range completeQueue {
			j := k - removedFromCompleted
			if completeQueue[j].Created.Before(algTimestamp) {
				algInput.JobQueue = append(algInput.JobQueue, completeQueue[j])
				completeQueue = completeQueue[:j+copy(completeQueue[j:], completeQueue[j+1:])]
				removedFromCompleted++
			}
		}

		queueMapBefore := make(map[string][]autoscale.AlgorithmJob)
		for _, j := range algInput.JobQueue {
			queueMapBefore[j.Tag] = append(queueMapBefore[j.Tag], j)
		}
		totalCostBeforeMap := make(map[string]float64)
		for key, queueBefore := range queueMapBefore {
			if key == "" {
				continue
			}
			totalCostBeforeMap[key] = algInput.Clouds[key].GetTotalCost(queueBefore, algTimestamp)
		}

		//Run the algorithm
		out, err := sim.Algorithm.Run(algInput, algTimestamp)
		if err != nil {
			return nil, err
		}

		//Split the output to queues defined by tag
		queueMap := make(map[string][]autoscale.AlgorithmJob)
		for _, j := range out.JobQueue {
			queueMap[j.Tag] = append(queueMap[j.Tag], j)
		}
		resp := make(map[string][]autoscale.AlgorithmJob)

		newInputQueue := make([]autoscale.AlgorithmJob, 0)

		//Iterate the queues in the map
		for key, queue := range queueMap {
			if _, ok := algInput.Clouds[key]; !ok {
				continue
			}
			instances, err := algInput.Clouds[key].GetInstances()
			if err != nil {
				return nil, err
			}

			instancesInactive := 0
			instancesActive := 0
			for _, i := range instances {
				if i.State == autoscale.ACTIVE {
					instancesActive++
				} else {
					instancesInactive++
				}
			}
			runningJobs := 0
			notRunningJobs := 0
			for _, i := range queue {
				if i.State == autoscale.RUNNING {
					runningJobs++
				} else {
					notRunningJobs++
				}
			}

			sort.Slice(queue, func(i, j int) bool {
				if queue[i].State == autoscale.RUNNING {
					return true
				}
				return queue[i].Priority > queue[j].Priority
			})

			//Iterate the queue
			deleted := 0
			for k := range queue {
				j := k - deleted
				//Simulate the job manager launching the job on the correct cluster
				if instancesActive > runningJobs {
					if queue[j].State != autoscale.RUNNING {
						queue[j].State = autoscale.RUNNING
						runningJobs++
						notRunningJobs--
						queue[j].Started = algTimestamp
					}
				} else if instancesInactive != 0 {
					if queue[j].State != autoscale.RUNNING {
						queue[j].State = autoscale.RUNNING
						queue[j].Started = algTimestamp
						runningJobs++
						notRunningJobs--
						for index := range instances {
							if instances[index].State == autoscale.INACTIVE {
								instances[index].State = autoscale.ACTIVE
								instancesInactive--
								instancesActive++
								break
							}
						}
					}
				}
				//Simulates a job finishing
				t := queue[j].Started.Add(time.Duration(time.Millisecond * time.Duration(queue[j].ExecutionTime[queue[j].Tag])))
				if t.Before(algTimestamp) && queue[j].State == autoscale.RUNNING {
					queue[j].State = autoscale.FINISHED
					instanceIndex := 0
					for _, instance := range instances {
						if instance.State == autoscale.ACTIVE {
							if queue[j].InstanceFlavour != "" {
								if queue[j].InstanceFlavour == instances[instanceIndex].Type {
									break
								}
							} else {
								break
							}
						}
						instanceIndex++
					}
					if instanceIndex < len(instances) && instances[instanceIndex].State == autoscale.ACTIVE {
						instances[instanceIndex].State = autoscale.INACTIVE
						queue = queue[:j+copy(queue[j:], queue[j+1:])]
						deleted++
					} else {
						sim.Log.Println("Something went wrong with the simulation of job state transition. No active instances found")
					}
				}
			}

			dur, err := algInput.Clouds[key].GetTotalDuration(queue, algTimestamp)
			if err != nil {
				return nil, err
			}
			queueCost := algInput.Clouds[key].GetTotalCost(queue, algTimestamp)
			err = models.InsertSimulatorEvent(sim.DB, models.SimulatorEvent{
				RunName:            simId,
				QueueDuration:      dur,
				AlgorithmTimestamp: algTimestamp,
				Tag:                key,
				CostBefore:         totalCostBeforeMap[key],
				CostAfter:          queueCost,
			})
			if err != nil {
				return nil, err
			}

			for _, jobAfterDelete := range queue {
				newInputQueue = append(newInputQueue, jobAfterDelete)
				//err = models.InsertAlgorithmJob(sim.DB, jobAfterDelete, simId)
				if err != nil {
					return nil, err
				}
			}

			resp[key] = queue
		}
		jsonSimQueue[i] = resp
		algInput.JobQueue = newInputQueue
	}
	err := sim.endRun(simId)
	if err != nil {
		return nil, err
	}
	return jsonSimQueue, nil
}

func (sim *Simulator) metapipeSimulationHandle(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("SimulationRequest: /metapipe/simulate/")

	metaOutput := sim.handleMetapipe(r)
	if metaOutput.err != nil {
		sim.Log.Print(metaOutput.err)
		http.Error(w, metaOutput.err.Error(), http.StatusInternalServerError)
		return
	}
	jsonSimQueue, err := sim.simulate(metaOutput.id, metaOutput.jobs, metaOutput.input, metaOutput.timestamp, metaOutput.timestep, metaOutput.iterations)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sim.Log.Println("FINISHED SIMULATION")

	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(&jsonSimQueue)
}

func (sim *Simulator) endRun(id string) error {
	err := models.UpdateAutoscalingRun(sim.DB, id, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (sim *Simulator) indexHandle(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("IndexRequest: /")
	err := sim.renderTemplate(w, "index", []string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (sim *Simulator) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) error {
	err := sim.templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		return err
	}
	return nil
}

func setScalingIds(clouds autoscale.CloudCollection, id string) {
	for _, c := range clouds {
		c.SetScalingId(id)
	}
}

func (sim *Simulator) createMetapipeClouds(inClusterStates autoscale.ClusterCollection) (autoscale.CloudCollection, error) {
	simCloudMap := make(autoscale.CloudCollection)

	simCloudMap[metapipe.CPouta] = &SimCloud{
		Cluster: inClusterStates[metapipe.CPouta],
		Db:      sim.DB,
	}
	simCloudMap[metapipe.AWS] = &SimCloud{
		Cluster: inClusterStates[metapipe.AWS],
		Db:      sim.DB,
	}
	simCloudMap[metapipe.Stallo] = &SimCloud{
		Cluster: inClusterStates[metapipe.Stallo],
		Db:      sim.DB,
	}
	return simCloudMap, nil
}

func (sim *Simulator) handleMetapipe(r *http.Request) (metapipeReturn) {
	utcNorway := int((time.Hour).Seconds())
	nor := time.FixedZone("Norway", utcNorway)
	defaultTime := time.Date(2018, 5, 23, 20, 40, 23, 0, nor)
	jobs := metapipe.GetMetapipeJobs(defaultTime.Add(time.Duration(time.Minute * -5)))
	var algInput autoscale.AlgorithmInput
	var simC autoscale.CloudCollection
	var reqInput metapipe.ScalingRequestInput
	var retVal metapipeReturn

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqInput)
	if err != nil && err != io.EOF {
		return retVal
	}

	if reqInput.Clusters != nil {
		simC, err = sim.createMetapipeClouds(reqInput.Clusters)
		if err != nil {
			return retVal
		}
	} else {
		simC, err = sim.createMetapipeClouds(sim.SimClusters)
		if err != nil {
			return retVal
		}
	}

	algInput.Clouds = simC

	friendlyName := reqInput.Name
	if reqInput.Jobs != nil {
		algjobs, err := metapipe.ConvertMetapipeQueueToAlgInputJobs(reqInput.Jobs)
		retVal.err = err
		if err != nil {
			return retVal
		}
		retVal.jobs, retVal.err = sim.Estimator.ProcessQueue(algjobs)
		if retVal.err != nil {
			return retVal
		}
	} else {
		retVal.jobs = jobs[:]
	}

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}
	retVal.id, err = models.CreateAutoscalingRun(sim.DB, friendlyName, time.Now())
	if retVal.err != nil {
		return retVal
	}
	setScalingIds(simC, retVal.id)
	if reqInput.StartTime != "" {
		retVal.timestamp, retVal.err = metapipe.ParseMetapipeTimestamp(reqInput.StartTime)
	} else {
		retVal.timestamp = defaultTime
	}

	retVal.input = algInput
	retVal.err = nil
	retVal.timestep = reqInput.Timestep
	retVal.iterations = reqInput.Iterations
	if retVal.timestep == 0 {
		retVal.timestep = 30
	}
	if retVal.iterations == 0 {
		retVal.iterations = 96
	}
	return retVal
}
