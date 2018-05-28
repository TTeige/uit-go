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
	"net/url"
)

type scalingInput struct {
	Name      string         `json:"name"`
	Queue     []metapipe.Job `json:"queue"`
	StartTime string         `json:"start_time"`
}

type prevRun struct {
	Queue []autoscale.AlgorithmJob `json:"queue"`
	CloudEvents []models.CloudEvent `json:"cloud_events"`
	Start time.Time `json:"start"`
	Finish time.Time `json:"finish"`
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
		friendlyName = ksuid.New().String()
		s.Log.Printf("No name was given, created name %s", friendlyName)
	}

	if reqInput.Queue == nil {
		s.Log.Print("Empty queue")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	algJobs, err := metapipe.ConvertMetapipeQueueToAlgInputJobs(reqInput.Queue)
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	algInput.JobQueue, err = s.Estimator.ProcessQueue(algJobs)
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		s.Log.Println("No algorithm time provided in request, setting time to now")
		algTimestamp = time.Now()
	}

	algInput.Clouds = s.Clouds

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

	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	err = enc.Encode(&out.JobQueue)
}

func (s *Service) getPreviousScalingHandle(w http.ResponseWriter, r *http.Request) {

	raw, err := url.Parse(r.RequestURI)
	if err != nil {
		s.Log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var out prevRun

	q := raw.Query()
	if val, ok := q["id"]; ok {

		run, err := models.GetAutoscalingRun(s.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			s.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events, err := models.GetAutoscalingRunEvents(s.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			s.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jobs, err := models.GetAllAlgorithmJobs(s.DB, val[0])
		if err != nil && err != sql.ErrNoRows {
			s.Log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out.CloudEvents = events
		out.Queue = jobs
		out.Start = run.Started
		out.Finish = run.Finished.Time
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(&out)
}

func (s *Service) Run() error {
	s.serve()
	return nil
}

func (s *Service) serve() {

	r := mux.NewRouter()
	r.HandleFunc("/", s.indexHandle).Methods("GET")
	r.HandleFunc("/metapipe/autoscale/", s.runScalingHandle).Methods("POST")
	r.HandleFunc("/metapipe/autoscale/", s.getPreviousScalingHandle).Methods("GET")

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
