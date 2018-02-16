package models

import (
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"
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

func insertJobAndParam(db *sql.DB, job Job, par Parameters) error {
	err := InsertJob(db, job)
	if err != nil {
		return err
	}

	err = InsertParameter(db, par)
	if err != nil {
		return err
	}
	return nil
}

func InitDatabase(db *sql.DB) error {

	jobs, err := GetAllJobs(db)
	if err != nil {
		return err
	}

	if len(jobs) > 0 {
		return nil
	}

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
	/*
	TODO: Add batch insertion instead of individual inserts, not a problem yet, but if the database grows it will become slow
	*/
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
			par := job.Parameters
			par.JobId = job.Id

			err = insertJobAndParam(db, dbJob, par)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func OpenDatabase(DB_USER string, DB_NAME string, DB_PASSWORD string) (*sql.DB, error) {
	dbStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
