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
	"strings"
)

type Simulator struct {
	DB        *sql.DB
	Hostname  string
	SimClouds map[string]autoscale.Cloud
	Algorithm autoscale.Algorithm
	Log       *log.Logger
	templates *template.Template
	tmplLoc   string
}

func (sim *Simulator) Run() {
	sim.Log.Printf("Starting the auto scaling simulator at: %s ", sim.Hostname)
	sim.tmplLoc = "simulator/templates/"
	sim.templates = template.Must(template.ParseFiles(sim.tmplLoc+"footer.html", sim.tmplLoc+"header.html",
		sim.tmplLoc+"index.html", sim.tmplLoc+"navbar.html"))
	sim.serveSim()
}

func (sim *Simulator) serveSim() {
	r := mux.NewRouter()
	r.HandleFunc("/", sim.indexHandle).Methods("GET")
	r.HandleFunc("/simulate/", sim.simulationHandle).Methods("POST")
	r.HandleFunc("/simulate/instancetype/", sim.simulationHandle).Methods("POST")
	r.HandleFunc("/simulate/{id}/", sim.getPreviousScalingHandle).Methods("GET")
	http.ListenAndServe(sim.Hostname, r)
}

func (sim *Simulator) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

}

func (sim *Simulator) addInstanceType(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	increment, err := strconv.Atoi(r.PostForm.Get("price_increment"))
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
	// An empty body request to this endpoint will initiate the simulation based on the inputted algorithm
	// A body with jobs: {...} uses these jobs, else, standard jobs are used
	// A body with cloudState: {...} uses this cloud state, else, standard state is used

	jobs := []autoscale.AlgorithmJob{
		{
			BaseJob: autoscale.BaseJob{
				Id:         "123abc",
				Tags:       []string{"aws", "cpouta"},
				Parameters: []string{"removeNonCompleteGenes", "useBlastUniref50"},
				State:      "RUNNING",
				Priority:   2000,
			},
			StateTransitionTimes: nil,
			Deadline:             time.Now().Add(time.Hour),
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


	clust, err := models.GetDefaultClusters(sim.DB)
	if err != nil {
		sim.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var algInput autoscale.AlgorithmInput
	algInput.JobQueue = jobs
	algInput.Clusters = clust

	SimStartTime := time.Now()
	friendlyName := r.PostForm.Get("name")

	if friendlyName == "" {
		friendlyName = ksuid.New().String()
	}


	simHandle, err := models.CreateSimulation(sim.DB, friendlyName, SimStartTime)
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

		for _, e := range out.Instances {
			for key, cluster := range sim.SimClouds {
				if strings.Contains(strings.ToLower(e.ClusterTag), strings.ToLower(key)) {
					err = cluster.ProcessEvent(e, simHandle.Name)
					if err != nil {
						sim.Log.Print(err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
		}
		algInput.JobQueue = out.JobQueue
	}

}

func (sim *Simulator) indexHandle(w http.ResponseWriter, r *http.Request) {
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
