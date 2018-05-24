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
	"sort"
	"github.com/tteige/uit-go/metapipe"
)

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
	r.HandleFunc("/simulate/", sim.simulationHandle).Methods("POST")
	r.HandleFunc("/simulate/instancetype/", sim.simulationHandle).Methods("POST")
	r.HandleFunc("/simulation/", sim.getPreviousScalingHandle).Methods("GET")
	r.HandleFunc("/simulation/all", sim.getAllSimulations).Methods("GET")
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
		sim.Log.Printf("GetAllSimulationsRequest: /simulation/?id=%s", q["id"])
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

func (sim *Simulator) simulationHandle(w http.ResponseWriter, r *http.Request) {
	var out autoscale.AlgorithmOutput
	sim.Log.Print("SimulationRequest: /simulate/")

	simId, algInput, completeQueue, algTimestamp, err := sim.handleMetapipe(r)
	jsonSimQueue := make(map[int]map[string][]autoscale.AlgorithmJob)

	sim.Log.Printf("Starting simulation: %s", simId)

	for i := 0; i < 96; i++ {
		if i > 0 {
			//Simulates a 30 minute interval between each scaling attempt
			algTimestamp = algTimestamp.Add(time.Minute * time.Duration(30))
		}

		deleted := 0
		for k := range completeQueue {
			j := k - deleted
			if completeQueue[j].Created.Before(algTimestamp) {
				algInput.JobQueue = append(algInput.JobQueue, completeQueue[j])
				completeQueue = completeQueue[:j+copy(completeQueue[j:], completeQueue[j+1:])]
				deleted++
			}
		}
		out, err = sim.Algorithm.Run(algInput, algTimestamp)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		queueMap := make(map[string][]autoscale.AlgorithmJob)
		for _, j := range out.JobQueue {
			queueMap[j.Tag] = append(queueMap[j.Tag], j)
		}
		resp := make(map[string][]autoscale.AlgorithmJob)

		newInputQueue := make([]autoscale.AlgorithmJob, 0)

		for key, queue := range queueMap {
			instances, err := algInput.Clouds[key].GetInstances()
			if err != nil {
				sim.Log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
			sort.Slice(queue, func(i, j int) bool {
				return queue[i].Priority > queue[j].Priority
			})

			deleted := 0
			for k := range queue {
				j := k - deleted
				if instancesActive > runningJobs {
					if queue[j].State != "RUNNING" {
						queue[j].State = "RUNNING"
						deadline := queue[j].Deadline.Sub(queue[j].Created)
						queue[j].Created = algTimestamp
						queue[j].Deadline.Add(deadline)
					}
				}

				if deleted == len(instances) {
					break
				}
				t := queue[j].Created.Add(time.Duration(time.Millisecond * time.Duration(queue[j].ExecutionTime[0])))
				if t.Before(algTimestamp) && queue[j].State == "RUNNING" {
					queue[j].State = "FINISHED"
					sim.Log.Printf("Job: %+v finished on instance %+v %s", queue[j], instances[deleted], key)
					queue = queue[:j+copy(queue[j:], queue[j+1:])]
					instances[deleted].State = "INACTIVE"
					deleted++
				}
			}

			for _, jobAfterDelete := range queue {
				newInputQueue = append(newInputQueue, jobAfterDelete)
			}

			resp[key] = queue
		}
		jsonSimQueue[i] = resp
		algInput.JobQueue = newInputQueue
		sim.Log.Printf("%+v", out.Instances)
	}
	err = sim.endRun(simId)
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

func (sim *Simulator) createClouds(inClusterStates autoscale.ClusterCollection) (autoscale.CloudCollection, error) {
	simCloudMap := make(autoscale.CloudCollection)

	simCloudMap[autoscale.CPouta] = &SimCloud{
		Cluster: inClusterStates[autoscale.CPouta],
		Db:      sim.DB,
	}
	simCloudMap[autoscale.AWS] = &SimCloud{
		Cluster: inClusterStates[autoscale.AWS],
		Db:      sim.DB,
	}
	simCloudMap[autoscale.Stallo] = &SimCloud{
		Cluster: inClusterStates[autoscale.Stallo],
		Db:      sim.DB,
	}
	return simCloudMap, nil
}

func (sim *Simulator) handleMetapipe(r *http.Request) (string, autoscale.AlgorithmInput, []autoscale.AlgorithmJob, time.Time, error) {
	utcNorway := int((time.Hour).Seconds())
	nor := time.FixedZone("Norway", utcNorway)
	defaultTime := time.Date(2018, 5, 23, 20, 40, 23, 0, nor)
	jobs := getMetapipeJobs(defaultTime.Add(time.Duration(time.Minute * -5)))
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
		simC, err = sim.createClouds(reqInput.Clusters)
		if err != nil {
			return simId, algInput, completeQueue, algTimestamp, err
		}
	} else {
		simC, err = sim.createClouds(sim.SimClusters)
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

func getMetapipeJobs(defaultTime time.Time) []autoscale.AlgorithmJob {
	return []autoscale.AlgorithmJob{
		{
			Id:  "1",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "RUNNING",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour),
			Created:       defaultTime.Add(-time.Hour),
			ExecutionTime: []int64{91847471},
		},
		{
			Id:  "2",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(-time.Hour),
			ExecutionTime: []int64{130184712},
		},
		{
			Id:  "3",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(time.Hour),
			ExecutionTime: []int64{14712732},
		},
		{
			Id:  "a",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 2),
			Created:       defaultTime.Add(time.Hour),
			ExecutionTime: []int64{24712734},
		},
		{
			Id:  "b",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			ExecutionTime: []int64{2347127},
		},
		{
			Id:  "c",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			ExecutionTime: []int64{6647127},
		},
		{
			Id:  "d",
			Tag: "aws",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			ExecutionTime: []int64{6647127},
		},
		{
			Id:  "e",
			Tag: "csc",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			ExecutionTime: []int64{6647127},
		},
		{
			Id:  "f",
			Tag: "csc",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			ExecutionTime: []int64{6647127},
		},
		{
			Id:  "g",
			Tag: "csc",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			ExecutionTime: []int64{3547127},
		},
		{
			Id:  "h",
			Tag: "metapipe",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			ExecutionTime: []int64{3547127},
		},
		{
			Id:  "i",
			Tag: "metapipe",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			ExecutionTime: []int64{21347127},
		},
		{
			Id:  "j",
			Tag: "metapipe",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(time.Hour * 2),
			ExecutionTime: []int64{45347127},
		},
		{
			Id:  "k",
			Tag: "metapipe",
			Parameters: metapipe.ConvertFromMetapipeParameters(metapipe.Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			Deadline:      defaultTime.Add(time.Hour * 2),
			Created:       defaultTime.Add(time.Hour * 1),
			ExecutionTime: []int64{34712713},
		},
	}
}
