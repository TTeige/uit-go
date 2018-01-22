package autoscalingServer

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/tteige/uit-go/models"
	"database/sql"
	"log"
	"fmt"
)

type Env struct {
	db *sql.DB
}

func indexHandle(env *Env) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		jobs, err := models.AllJobs(env.db)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
			return
		}
		for _, job := range jobs {
			w.Write([]byte(fmt.Sprintf("%s, %d, %#v, %#v", job.Id, job.Runtime, job.Parameters, job.Tags)))
		}

	})
	// Create serving for the overview of the algorithms that have been input to the scaling server
}

func submitJobHistoryHandle(env *Env) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {

	})
	// Submit metapipe history to the server, store the history and convert it to the appropriate format
}

func runScalingHandle(env *Env) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {

	})
	// Run the autoscaling algorithm
	// route the request to a list of whitelisted addresses. This should enable algorithms implemented in any language
	// wait for response and write to database
	// Request contains the current queue
	// Do not create 201 response immediately and return expected output URI because of RFC 7231, section 6.3.2
	// Use 202 Accepted and return expected output location /history/autoscaling/{id}

	// If the request is to the same scaling algorithm, maybe create a queue to hold them instead.
}

func getPreviousScalingHandle(env *Env) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {

	})
	// Return previous values for a simulated run

	// Request for resource which has not yet completed, no entry in database, return (303 See Other) code, with a
	// Location field in the header for /history/autoscaling/{id}
	// If error for the autoscaling, return appropriate error code
}

func Serve(hostUrl string) {

	db, err := models.NewDatabase("user=tim dbname=autoscaling password=something")
	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	env := &Env{db:db}

	r := mux.NewRouter()
	r.Path("/").Methods("GET").Handler(indexHandle(env))
	r.Path("/autoscale/").Methods("POST").Handler(runScalingHandle(env))
	hist := r.PathPrefix("/history").Subrouter()
	hist.Path("/jobs/").Methods("POST").Handler(submitJobHistoryHandle(env))
	hist.Path("/autoscaling/{id}/").Methods("GET").Handler(getPreviousScalingHandle(env))

	http.ListenAndServe(hostUrl, r)
}
