package simulator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/tteige/uit-go/autoscale"
)

type Simulator struct {
	DB        *sql.DB
	Hostname  string
	SimClouds []autoscale.Cloud
	Algorithm autoscale.AlgorithmInterface
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
	//An empty body request to this endpoint will initiate the simulation based on the inputted algorithm
	//qh := autoscale.QueueHandle{
	//	DB: sim.DB,
	//}

}

func (sim *Simulator) indexHandle(w http.ResponseWriter, r *http.Request) {

}
