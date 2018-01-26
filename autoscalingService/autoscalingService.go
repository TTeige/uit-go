package autoscalingService

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs, err := models.AllJobs(env.db)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for _, job := range jobs {
			w.Write([]byte(fmt.Sprintf("%s, %d, %#v, %#v", job.Id, job.Runtime, job.Parameters, job.Tags)))
		}

	})
	// Create serving for the overview of the algorithms that have been input to the scaling server
}

func runScalingHandle(env *Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
	// Run the autoscaling algorithm
	// Request contains the current queue
	// Do not create 201 response immediately and return expected output URI because of RFC 7231, section 6.3.2
	// Use 202 Accepted and return expected output location autoscale/{id}
}

func getPreviousScalingHandle(env *Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
	// Request for resource which has not yet completed, no entry in database, return (303 See Other) code, with a
	// Location field in the header for /autoscale/{id}
	// If error for the autoscaling, return appropriate error code
}

func Run(hostUrl string, db *sql.DB) error {
	serve(hostUrl, db)
	return nil
}

func serve(hostUrl string, db *sql.DB) {

	env := &Env{db: db}

	r := mux.NewRouter()
	r.Path("/").Methods("GET").Handler(indexHandle(env))
	r.Path("/autoscale/").Methods("POST").Handler(runScalingHandle(env))
	r.Path("/autoscale/{id}/").Methods("GET").Handler(getPreviousScalingHandle(env))

	http.ListenAndServe(hostUrl, r)
}
