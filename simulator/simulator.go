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
	"strconv"
	"github.com/segmentio/ksuid"
	"encoding/json"
	"net/url"
	"github.com/tteige/uit-go/clouds"
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
	sim.Log.Print("GetAllSimulationsRequest: /simulation/?id=")
	raw, err := url.Parse(r.RequestURI)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var simList [][]models.SimEvent
	q := raw.Query()
	if val, ok := q["id"]; ok {
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

func (sim *Simulator) addInstanceType(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("AddInstanceTypeRequest: /simulate/instancetype/")

	err := r.ParseForm()
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	increment, err := strconv.ParseFloat(r.PostForm.Get("price_increment"), 64)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	iType := autoscale.InstanceType{Name: r.PostForm.Get("name"), PriceIncrement: increment}

	clusterName := r.PostForm.Get("cluster_name")

	err = models.InsertInstanceType(sim.DB, iType, clusterName)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	return
}

func (sim *Simulator) simulationHandle(w http.ResponseWriter, r *http.Request) {
	sim.Log.Print("SimulationRequest: /simulate/")
	// An empty body request to this endpoint will initiate the simulation based on the inputted algorithm
	// A body with jobs: {...} uses these jobs, else, standard jobs are used
	// A body with cloudState: {...} uses this cloud state, else, standard state is used

	jobs := []autoscale.AlgorithmJob{
		{
			Id:  "123abc",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "RUNNING",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour),
			Created:       time.Now().Add(-time.Hour),
			ExecutionTime: 1301847471273,
		},
		{
			Id:  "1213455abc",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 3),
			Created:       time.Now().Add(-time.Hour),
			ExecutionTime: 13018471273,
		},
		{
			Id:  "123aaaaaaabc",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 3),
			Created:       time.Now().Add(time.Hour),
			ExecutionTime: 1471273,
		},
		{
			Id:  "a",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 2),
			Created:       time.Now().Add(time.Hour),
			ExecutionTime: 1471273,
		},
		{
			Id:  "b",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now().Add(time.Minute * 30),
			ExecutionTime: 1471273,
		},
		{
			Id:  "c",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now().Add(time.Minute * 30),
			ExecutionTime: 1471273,
		},
		{
			Id:  "d",
			Tag: "aws",
			Parameters: autoscale.MetapipeParameter{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now(),
			ExecutionTime: 1471273,
		},
	}

	dec := json.NewDecoder(r.Body)
	var reqInput autoscale.ScalingRequestInput
	err := dec.Decode(&reqInput)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var algInput autoscale.AlgorithmInput
	var simC autoscale.CloudCollection
	if reqInput.Clusters != nil {
		simC, err = sim.createClouds(reqInput.Clusters)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		simC, err = sim.createClouds(sim.SimClusters)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	algInput.Clouds = simC

	friendlyName := reqInput.Name
	if reqInput.Jobs != nil {
		algInput.JobQueue, err = sim.Estimator.ProcessQueue(reqInput.Jobs)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		algInput.JobQueue = jobs
	}

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}

	simId, err := models.CreateSimulation(sim.DB, friendlyName, time.Now())
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setScalingIds(simC, simId)
	var algTimestamp time.Time
	if reqInput.StartTime != "" {
		algTimestamp, err = autoscale.ParseMetapipeTimestamp(reqInput.StartTime)
	} else {
		algTimestamp = time.Now()
	}
	for i := 0; i < 10; i++ {
		if i > 0 {
			//Simulates a 30 minute interval between each scaling attempt
			algTimestamp = algTimestamp.Add(time.Minute * time.Duration(30))
		}
		var jInput []autoscale.AlgorithmJob
		for _, j := range algInput.JobQueue {
			if j.Created.Before(algTimestamp) {
				jInput = append(jInput, j)
			}
		}
		algInput.JobQueue = jInput
		out, err := sim.Algorithm.Run(algInput, algTimestamp)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		queueMap := make(map[string][]autoscale.AlgorithmJob)
		for _, j := range out.JobQueue {
			queueMap[j.Tag] = append(queueMap[j.Tag], j)
		}

		for key, queue := range queueMap {
			instances, err := simC[key].GetInstances()
			if err != nil {
				sim.Log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			deleted := 0
			for k := range queue {
				j := k - deleted
				if deleted == len(instances) {
					break
				}
				if out.JobQueue[j].Created.Add(time.Duration(time.Millisecond * time.Duration(out.JobQueue[j].ExecutionTime))).Before(algTimestamp) {
					if instances[deleted].State == "INACTIVE" {
						continue
					}
					instances[deleted].State = "INACTIVE"
					log.Printf("Job: %+v finished on instance %+v", algInput.JobQueue[j], instances[deleted])
					algInput.JobQueue = out.JobQueue[:j+copy(out.JobQueue[j:], out.JobQueue[j+1:])]
					deleted++
				}
			}
		}
		algInput.JobQueue = out.JobQueue

		sim.Log.Printf("%+v", out.Instances)
	}
	err = sim.endRun(simId)
	sim.Log.Println("FINISHED SIMULATION")
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	simCloudMap[autoscale.CPouta] = &clouds.SimCloud{
		Cluster: inClusterStates[autoscale.CPouta],
		Db:      sim.DB,
	}
	simCloudMap[autoscale.AWS] = &clouds.SimCloud{
		Cluster: inClusterStates[autoscale.AWS],
		Db:      sim.DB,
	}
	simCloudMap[autoscale.Stallo] = &clouds.SimCloud{
		Cluster: inClusterStates[autoscale.Stallo],
		Db:      sim.DB,
	}
	return simCloudMap, nil
}
