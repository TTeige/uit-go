package simulator

import (
	"database/sql"
	"github.com/gorilla/mux"
	"net/http"
)

func Run(hostUrl string, db *sql.DB) {
	serveSim(hostUrl, db)
}

func serveSim(hostUrl string, db *sql.DB) {
	r := mux.NewRouter()
	r.Path("/").Methods("GET").Handler(indexHandle(db))
	r.Path("/simulate/").Methods("POST").Handler(simulationHandle(db))
	r.Path("/simulate/{id}/").Methods("GET").Handler(getPreviousScalingHandle(db))

	http.ListenAndServe(hostUrl, r)
}
func getPreviousScalingHandle(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func simulationHandle(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func indexHandle(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}