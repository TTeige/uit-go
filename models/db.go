package models

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
	"strconv"
	"fmt"
	"time"
	"log"
	"github.com/tteige/uit-go/metapipe"
)

func getTotalDuration(job metapipe.Job) (int64, error) {
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

func InitDatabase(db *sql.DB, auth metapipe.Oath2, fetchNewJobs bool) error {

	if !fetchNewJobs {
		return nil
	}

	jobs, err := GetAllJobs(db)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	client := metapipe.RetryClient{
		Auth:        auth,
		MaxAttempts: 3,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}
	for _, j := range jobs {
		if j.InputDataSize == 0 {

			size, err := client.GetMetapipeJobSize(j.JobId)
			if err != nil {
				return err
			}
			if size > 0 {
				err = UpdateJob(db, *j)
				if err != nil {
					return err
				}
			}
		}
	}

	all, err := client.GetAllMetapipeJobs()
	if err != nil {
		return err
	}
	/*
	TODO: Add batch insertion instead of individual inserts, not a problem yet, but if the database grows it will become slow
	*/
	log.Printf("Begin insertions")
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
			par := Parameters{
				MP: job.Parameters,
				JobId:      job.Id,
			}

			s, err := client.GetMetapipeJobSize(job.Id)
			dbJob.InputDataSize = s
			if err != nil {
				return err
			}

			err = insertJobAndParam(db, dbJob, par)
			if err != nil {
				return err
			}
		}
	}
	log.Printf("Insertions complete")

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
