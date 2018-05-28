package autoscalingService

import (
	"github.com/gorilla/mux"
	"net/http"
	"database/sql"
	"github.com/tteige/uit-go/metapipe"
	"github.com/segmentio/ksuid"
	"github.com/tteige/uit-go/models"
	"time"
	"encoding/json"
	"io"
	"github.com/tteige/uit-go/autoscale"
	"log"
)

type scalingInput struct {
	Name      string         `json:"name"`
	Queue     []metapipe.Job `json:"queue"`
	StartTime string         `json:"start_time"`
}

type Service struct {
	DB        *sql.DB
	Hostname  string
	Clouds    autoscale.CloudCollection
	Algorithm autoscale.Algorithm
	Log       *log.Logger
	Estimator autoscale.Estimator
}

func (s *Service) indexHandle(w http.ResponseWriter, r *http.Request) {
	// Create serving for the overview of the algorithms that have been input to the scaling server
}

func (s *Service) runScalingHandle(w http.ResponseWriter, r *http.Request) {
	// Run the autoscaling algorithm
	// Request contains the current queue
	// Do not create 201 response immediately and return expected output URI because of RFC 7231, section 6.3.2
	// Use 202 Accepted and return expected output location autoscale/{id}
	var reqInput scalingInput
	var algInput autoscale.AlgorithmInput
	var algTimestamp time.Time
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqInput)
	if err != nil && err != io.EOF {
		return
	}

	friendlyName := reqInput.Name
	if friendlyName == "" {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		friendlyName = ksuid.New().String()
	}
	runId, err := models.CreateAutoscalingRun(s.DB, friendlyName, time.Now())
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setScalingIds(s.Clouds, runId)
	if reqInput.StartTime != "" {
		algTimestamp, err = metapipe.ParseMetapipeTimestamp(reqInput.StartTime)
	} else {
		algTimestamp = time.Now()
	}

	//Run the algorithm
	out, err := s.Algorithm.Run(algInput, algTimestamp)
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, job := range out.JobQueue {
		err = models.InsertAlgorithmJob(s.DB, job, runId)
		if err != nil {
			s.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = s.endRun(runId)
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(out)
	w.Write(b)
}

func (s *Service) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

}

func (s *Service) Run() error {
	s.serve()
	return nil
}

func (s *Service) serve() {

	r := mux.NewRouter()
	r.HandleFunc("/", s.indexHandle).Methods("GET")
	r.HandleFunc("/metapipe/autoscale/", s.runScalingHandle).Methods("POST")
	r.HandleFunc("/metapipe/autoscale/{id}/", s.getPreviousScalingHandle).Methods("GET")

	http.ListenAndServe(s.Hostname, r)
}

func setScalingIds(clouds autoscale.CloudCollection, id string) {
	for _, c := range clouds {
		c.SetScalingId(id)
	}
}

func (s *Service) endRun(id string) error {
	err := models.UpdateAutoscalingRun(s.DB, id, time.Now())
	if err != nil {
		return err
	}
	return nil
}
