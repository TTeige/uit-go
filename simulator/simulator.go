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
	sims, err := models.GetAllSimulationStats(sim.DB)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(sims)
	w.Write(b)
}

func (sim *Simulator) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

	raw, err := url.Parse(r.RequestURI)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var simList [][]models.SimEvent
	q := raw.Query()
	if val, ok := q["id"]; ok {
		sim.Log.Printf("GetAllSimulationsRequest: /metapipe/simulation/?id=%s", q["id"])
		for _, i := range val {
			events, err := models.GetSimEvents(sim.DB, i)
			if err != nil && err != sql.ErrNoRows {
				sim.Log.Print(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			simList = append(simList, events)
		}
	}

	b, err := json.Marshal(simList)
	w.Write(b)
}

func (sim *Simulator) simulate(simId string, completeQueue []autoscale.AlgorithmJob, algInput autoscale.AlgorithmInput, algTimestamp time.Time) (simulationOutput, error) {
	jsonSimQueue := make(simulationOutput)
	sim.Log.Printf("Starting simulation: %s", simId)

	for i := 0; i < 96; i++ {
		if i > 0 {
			//Simulates a 30 minute interval between each scaling attempt
			algTimestamp = algTimestamp.Add(time.Minute * time.Duration(30))
		}

		//Selects jobs that are created before the timestamp
		removedFromCompleted := 0
		for k := range completeQueue {
			j := k - removedFromCompleted
			if completeQueue[j].Created.Before(algTimestamp) {
				algInput.JobQueue = append(algInput.JobQueue, completeQueue[j])
				//sim.Log.Printf("Added job %v to the queue", completeQueue[j])
				completeQueue = completeQueue[:j+copy(completeQueue[j:], completeQueue[j+1:])]
				removedFromCompleted++
			}
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
			instances, err := algInput.Clouds[key].GetInstances()
			if err != nil {
				return nil, err
			}

			instancesActive := 0
			for _, i := range instances {
				if i.State == "ACTIVE" {
					instancesActive++
				}
			}
			runningJobs := 0
			for _, i := range queue {
				if i.State == "RUNNING" {
					runningJobs++
				}
			}

			//Iterate the queue
			deleted := 0
			for k := range queue {
				j := k - deleted
				//Simulate the job manager launching the job on the correct cluster
				if instancesActive > runningJobs {
					if queue[j].State != "RUNNING" {
						queue[j].State = "RUNNING"
						runningJobs++
						queue[j].Started = algTimestamp
					}
				}
				//Simulates a job finishing
				t := queue[j].Started.Add(time.Duration(time.Millisecond * time.Duration(queue[j].ExecutionTime[queue[j].Tag])))
				if t.Before(algTimestamp) && queue[j].State == "RUNNING" {
					queue[j].State = "FINISHED"
					instanceIndex := sort.Search(len(instances), func(i int) bool {
						return instances[i].State == "ACTIVE"
					})
					if instanceIndex < len(instances) && instances[instanceIndex].State == "ACTIVE" {
						instances[instanceIndex].State = "INACTIVE"
						//sim.Log.Printf("Job: %+v finished on instance %+v %s", queue[j], instances[instanceIndex], key)
						queue = queue[:j+copy(queue[j:], queue[j+1:])]
						deleted++
					} else {
						sim.Log.Println("Something went wrong with the simulation of job state transition. No active instances found")
					}
				}
			}

			for _, jobAfterDelete := range queue {
				newInputQueue = append(newInputQueue, jobAfterDelete)
			}
			dur, err := getTotalDuration(queue, instances, algTimestamp)
			sim.Log.Printf("%s has a total expected queue duration of %v", key, time.Duration(dur)*time.Millisecond)

			resp[key] = queue
		}
		jsonSimQueue[i] = resp
		algInput.JobQueue = newInputQueue
		sim.Log.Println("-------------------------------------------------------------")
		//sim.Log.Printf("%+v", out.Instances)
	}
	err := sim.endRun(simId)
	if err != nil {
		return nil, err
	}
	return jsonSimQueue, nil
}

func (sim *Simulator) metapipeSimulationHandle(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("SimulationRequest: /metapipe/simulate/")

	simId, algInput, completeQueue, algTimestamp, err := sim.handleMetapipe(r)
	jsonSimQueue, err := sim.simulate(simId, completeQueue, algInput, algTimestamp)

	sim.Log.Println("FINISHED SIMULATION")
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(&jsonSimQueue)
}

func (sim *Simulator) endRun(id string) error {
	err := models.UpdateSim(sim.DB, id, time.Now())
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

func (sim *Simulator) handleMetapipe(r *http.Request) (string, autoscale.AlgorithmInput, []autoscale.AlgorithmJob, time.Time, error) {
	utcNorway := int((time.Hour).Seconds())
	nor := time.FixedZone("Norway", utcNorway)
	defaultTime := time.Date(2018, 5, 23, 20, 40, 23, 0, nor)
	jobs := metapipe.GetMetapipeJobs(defaultTime.Add(time.Duration(time.Minute * -5)))
	var algInput autoscale.AlgorithmInput
	var simC autoscale.CloudCollection
	var reqInput metapipe.ScalingRequestInput
	var algTimestamp time.Time
	var simId string
	var completeQueue []autoscale.AlgorithmJob

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqInput)
	if err != nil && err != io.EOF {
		return simId, algInput, completeQueue, algTimestamp, err
	}

	if reqInput.Clusters != nil {
		simC, err = sim.createMetapipeClouds(reqInput.Clusters)
		if err != nil {
			return simId, algInput, completeQueue, algTimestamp, err
		}
	} else {
		simC, err = sim.createMetapipeClouds(sim.SimClusters)
		if err != nil {
			return simId, algInput, completeQueue, algTimestamp, err
		}
	}

	algInput.Clouds = simC

	friendlyName := reqInput.Name
	if reqInput.Jobs != nil {
		algjobs, err := metapipe.ConvertMetapipeQueueToAlgInputJobs(reqInput.Jobs)
		completeQueue, err = sim.Estimator.ProcessQueue(algjobs)
		if err != nil {
			return simId, algInput, completeQueue, algTimestamp, err
		}
	} else {
		completeQueue = jobs[:]
	}

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}
	simId, err = models.CreateSimulation(sim.DB, friendlyName, time.Now())
	if err != nil {
		return simId, algInput, completeQueue, algTimestamp, err
	}
	setScalingIds(simC, simId)
	if reqInput.StartTime != "" {
		algTimestamp, err = metapipe.ParseMetapipeTimestamp(reqInput.StartTime)
	} else {
		algTimestamp = defaultTime
	}
	return simId, algInput, completeQueue, algTimestamp, nil
}

func getTotalDuration(queue []autoscale.AlgorithmJob, instances []autoscale.Instance, currentTime time.Time) (int64, error) {
	activeInstances := 0
	for _, i := range instances {
		if i.State == "ACTIVE" {
			activeInstances++
		}
	}
	if activeInstances == 0 {
		return 0, nil
	}
	longestInstanceUpTime := make([]int64, activeInstances)
	for i := 0; i < len(queue); i = i + activeInstances {
		for j := 0; j < activeInstances; j++ {
			if i+j > len(queue)-1 {
				break
			}
			timeLeftOfJob := time.Duration(queue[i+j].ExecutionTime[queue[i+j].Tag]) * time.Millisecond
			//The running jobs
			if queue[i+j].State == "RUNNING" {
				sinceStart := currentTime.Sub(queue[i+j].Started)
				timeLeftOfJob = timeLeftOfJob - sinceStart
			}
			longestInstanceUpTime[j] += int64(timeLeftOfJob / time.Millisecond)
		}
	}

	var longest int64
	longest = 0
	for _, l := range longestInstanceUpTime {
		if longest < l {
			longest = l
		}
	}
	return longest, nil
}
