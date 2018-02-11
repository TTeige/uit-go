package models

import (
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
	"net/http"
	"strconv"
	"log"
)

type inputFas struct {
	Url string `json:"url"`
}

type dataUrl struct {
	InputFas inputFas `json:"input.fas"`
}

type attempt struct {
	ExecutorId          string  `json:"executorId"`
	State               string  `json:"state"`
	AttemptId           string  `json:"attemptId"`
	Tag                 string  `json:"tag"`
	TimeCreated         string  `json:"timeCreated"`
	TimeStarted         string  `json:"timeStarted"`
	TimeEnded           string  `json:"timeEnded"`
	LastHeartbeat       string  `json:"lastHeartbeat"`
	RuntimeMillis       int     `json:"runtimeMillis"`
	QueueDurationMillis int     `json:"queueDurationMillis"`
	Outputs             dataUrl `json:"outputs"`
	Priority            int     `json:"priority"`
}

type metaJob struct {
	Id                       string     `json:"jobId"`
	TimeSubmitted            string     `json:"timeSubmitted"`
	State                    string     `json:"state"`
	UserId                   string     `json:"userId"`
	Tag                      string     `json:"tag"`
	Priority                 int        `json:"priority"`
	Hold                     bool       `json:"hold"`
	Parameters               Parameters `json:"parameters"`
	Inputs                   dataUrl    `json:"inputs"`
	Outputs                  dataUrl    `json:"outputs"`
	TotalRuntimeMillis       int64      `json:"totalRuntimeMillis"`
	TotalQueueDurationMillis int64      `json:"totalQueueDurationMillis"`
	Attempts                 []attempt  `json:"attempts"`
}

func getTotalDuration(job metaJob) (int64, error) {
	var totalDuration int64

	if job.TotalRuntimeMillis == 0 {
		var lastHeartbeat string
		for _, attempt := range job.Attempts {
			if attempt.State == "FINISHED" {
				lastHeartbeat = attempt.LastHeartbeat
			}
		}
		if lastHeartbeat != "" && job.TimeSubmitted != "" {
			lhb, err := strconv.ParseInt(lastHeartbeat, 10, 64)
			if err != nil {
				return 0, err
			}
			timeStart, err := strconv.ParseInt(job.TimeSubmitted, 10, 64)
			if err != nil {
				return 0, err
			}

			totalDuration = lhb - timeStart
		} else {
			totalDuration = 0
		}
	} else {
		totalDuration = job.TotalRuntimeMillis
	}
	return totalDuration, nil
}

func InitDatabase(db *sql.DB) error {

	resp, err := http.Get("https://jobs.metapipe.uit.no/jobs")
	if err != nil {
		return err
	}
	defer resp.Body.Close()


	var all []metaJob
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&all)
	if err != nil {
		return err
	}

	for _, job := range all {
		if job.State == "FINISHED" {
			totalDuration, err := getTotalDuration(job)

			if err != nil {
				return err
			}

			dbJob := Job{
				JobId:         job.Id,
				Tag:           job.Tag,
				Runtime:       totalDuration,
				InputDataSize: 0,
				QueueDuration: job.TotalQueueDurationMillis,
			}
			err = InsertJob(db, dbJob)
			if err != nil {
				return err
			}

			par := job.Parameters
			par.JobId = job.Id

			err = InsertParameter(db, par)
			if err != nil {
				log.Print(job)
				return err
			}
		}
	}

	return nil
}

func OpenDatabase(source string) (*sql.DB, error) {
	db, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
