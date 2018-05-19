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
			Id:            "123abc",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "RUNNING",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour),
			Created:       time.Now().Add(-time.Hour),
			ExecutionTime: 1301847471273,
		},
		{
			Id:            "1213455abc",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 3),
			Created:       time.Now().Add(-time.Hour),
			ExecutionTime: 13018471273,
		},
		{
			Id:            "123aaaaaaabc",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 3),
			Created:       time.Now().Add(time.Hour),
			ExecutionTime: 1471273,
		},
		{
			Id:            "a",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 2),
			Created:       time.Now().Add(time.Hour),
			ExecutionTime: 1471273,
		},
		{
			Id:            "b",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now().Add(time.Minute * 30),
			ExecutionTime: 1471273,
		},
		{
			Id:            "c",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now().Add(time.Minute * 30),
			ExecutionTime: 1471273,
		},
		{
			Id:            "d",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 1),
			Created:       time.Now(),
			ExecutionTime: 1471273,
		},
	}

	err := r.ParseForm()
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var algInput autoscale.AlgorithmInput
	simC, err := sim.createClouds(sim.SimClusters)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	algInput.Clouds = simC

	friendlyName := r.PostForm.Get("name")

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}

	simId, err := models.CreateSimulation(sim.DB, friendlyName, time.Now())
	setScalingIds(simC, simId)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	algStartTime := time.Now()
	for i := 0; i < 10; i++ {
		if i > 0 {
			//Simulates a 30 minute interval between each scaling attempt
			algStartTime = algStartTime.Add(time.Minute*time.Duration(5))
		}
		var jInput []autoscale.AlgorithmJob
		for _, j := range jobs {
			if j.Created.Before(algStartTime) {
				jInput = append(jInput, j)
			}
		}
		algInput.JobQueue = jInput
		out, err := sim.Algorithm.Run(algInput, algStartTime)
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		deleted := 0
		for k := range algInput.JobQueue {
			j := k - deleted
			if algInput.JobQueue[j].Created.Add(time.Duration(time.Millisecond * time.Duration(algInput.JobQueue[j].ExecutionTime))).Before(algStartTime) {
				instances, err := simC[algInput.JobQueue[j].Tag].GetInstances()
				if err != nil {
					sim.Log.Print(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				//Simulates the completion of a job by removing it from the job queue, removing an active resources for a job
				//This is reflected in the real world by getting the instance information and see that there are no running jobs
				instances[0].State = "INACTIVE"
				log.Printf("Job: %+v finished on instance %+v", algInput.JobQueue[j], instances[0])
				algInput.JobQueue = algInput.JobQueue[:j+copy(algInput.JobQueue[j:], algInput.JobQueue[j+1:])]
				deleted++
			}
		}
		sim.Log.Printf("%+v", out.Instances)
		algInput.JobQueue = out.JobQueue
	}
	err = sim.endRun(simId)
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
