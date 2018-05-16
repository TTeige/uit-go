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
)

type Simulator struct {
	DB        *sql.DB
	Hostname  string
	SimClouds map[string]autoscale.Cloud
	Algorithm autoscale.Algorithm
	Log       *log.Logger
	templates *template.Template
	tmplLoc   string
	Estimator autoscale.Estimator
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
			ExecutionTime: 1301847471273,
		},
		{
			Id:            "1213455abc",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 2),
			ExecutionTime: 13018471273,
		},
		{
			Id:            "123aaaaaaabc",
			Tag:           "aws",
			Parameters:    []string{"removeNonCompleteGenes", "useBlastUniref50"},
			State:         "QUEUED",
			Priority:      2000,
			Deadline:      time.Now().Add(time.Hour * 3),
			ExecutionTime: 1471273,
		},
	}

	//clusters := []autoscale.Cluster{
	//	{
	//		Name:       "aws",
	//		Limit:      100,
	//		AcceptTags: []string{"aws", "aws-meta"},
	//		Types:      nil,
	//	},
	//	{
	//		Name:       "cpouta",
	//		Limit:      1000,
	//		AcceptTags: []string{"cpouta", "csc"},
	//		Types:      nil,
	//	},
	//}

	err := r.ParseForm()
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var algInput autoscale.AlgorithmInput
	algInput.JobQueue = jobs
	algInput.Clouds = sim.SimClouds

	friendlyName := r.PostForm.Get("name")

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}

	simId, err := models.CreateSimulation(sim.DB, friendlyName, time.Now())
	setScalingIds(sim.SimClouds, simId)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	algStartTime := time.Now()
	for i := 0; i < 10; i++ {
		out, err := sim.Algorithm.Step(algInput, algStartTime.Add(time.Minute*time.Duration(30*i)))
		if err != nil {
			sim.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sim.Log.Printf("%+v", out)

		//for _, e := range out.Instances {
		//	for key, cluster := range sim.SimClouds {
		//		if strings.Contains(strings.ToLower(e.ClusterTag), strings.ToLower(key)) {
		//			err = cluster.ProcessEvent(e, simId.Name)
		//			if err != nil {
		//				sim.Log.Print(err)
		//				http.Error(w, err.Error(), http.StatusInternalServerError)
		//				return
		//			}
		//		}
		//	}
		//}
		algInput.JobQueue = out.JobQueue
	}
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
