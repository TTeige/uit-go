package simulator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/tteige/uit-go/autoscale"
	"time"
	"log"
)

type Simulator struct {
	DB        *sql.DB
	Hostname  string
	SimClouds []autoscale.Cloud
	Algorithm autoscale.Algorithm
}

func (sim *Simulator) Run() {

	sim.serveSim()
}

func (sim *Simulator) serveSim() {
	r := mux.NewRouter()
	r.HandleFunc("/", sim.indexHandle).Methods("GET")
	r.HandleFunc("/simulate/", sim.simulationHandle).Methods("POST")
	r.HandleFunc("/simulate/{id}/", sim.getPreviousScalingHandle).Methods("GET")
	http.ListenAndServe(sim.Hostname, r)
}

func (sim *Simulator) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

}

func (sim *Simulator) simulationHandle(w http.ResponseWriter, r *http.Request) {
	// An empty body request to this endpoint will initiate the simulation based on the inputted algorithm
	// A body with jobs: {...} uses these jobs, else, standard jobs are used
	// A body with cloudState: {...} uses this cloud state, else, standard state is used

	jobs := []autoscale.AlgorithmJob{
		{
			Id:                   "123abc",
			Tags:                 []string{"aws", "csc"},
			Parameters:           []string{"blast"},
			State:                "RUNNING",
			StateTransitionTimes: nil,
			Deadline:             time.Now().Add(time.Hour),
			Priority:             10,
		},
	}

	clusters := []autoscale.Cluster{
		{
			Name:            "aws",
			Limit:           100,
			AcceptTags:      []string{"aws", "aws-meta"},
			Types:           nil,
			ActiveInstances: nil,
		},
		{
			Name:            "cpouta",
			Limit:           1000,
			AcceptTags:      []string{"cpouta", "csc"},
			Types:           nil,
			ActiveInstances: nil,
		},
	}

	algInput := autoscale.AlgorithmInput{
		JobQueue: jobs,
		Clusters: clusters,
	}

	out, err := sim.Algorithm.Step(algInput, time.Now().Add(time.Hour*2))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("%+v", out)

}

func (sim *Simulator) indexHandle(w http.ResponseWriter, r *http.Request) {

}
