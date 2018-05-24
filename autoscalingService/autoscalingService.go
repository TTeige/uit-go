package autoscalingService

import (
	"github.com/gorilla/mux"
	"net/http"
	"database/sql"
)

type Service struct {
	DB       *sql.DB
	Hostname string
}

func (s *Service) indexHandle(w http.ResponseWriter, r *http.Request) {
	// Create serving for the overview of the algorithms that have been input to the scaling server
}

func (s *Service) runScalingHandle(w http.ResponseWriter, r *http.Request) {
	// Run the autoscaling algorithm
	// Request contains the current queue
	// Do not create 201 response immediately and return expected output URI because of RFC 7231, section 6.3.2
	// Use 202 Accepted and return expected output location autoscale/{id}
}

func (s *Service) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {
	// Request for resource which has not yet completed, no entry in database, return (303 See Other) code, with a
	// Location field in the header for /autoscale/{id}
	// If error for the autoscaling, return appropriate error code
}

func (s *Service) Run() error {
	s.serve()
	return nil
}

func (s *Service) serve() {

	r := mux.NewRouter()
	r.HandleFunc("/", s.indexHandle).Methods("GET")
	r.HandleFunc("/autoscale/", s.runScalingHandle).Methods("POST")
	r.HandleFunc("/autoscale/{id}/", s.getPreviousScalingHandle).Methods("GET")

	http.ListenAndServe(s.Hostname, r)
}
