package models

import (
	"database/sql"
	"github.com/tteige/uit-go/autoscale"
)

type Job struct {
	Id string
	Runtime int
	Tags []string
	Parameters []string
	InputDataSize int
}

func AllJobs(db *sql.DB) ([]*Job, error) {
	rows, err := db.Query("SELECT * FROM jobs")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	jobs := make([]*Job, 0)

	for rows.Next() {
		job := new(Job)
		err := rows.Scan(&job.Id, &job.Runtime, &job.Tags, &job.Parameters)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func InsertJob(db *sql.DB, job Job) error {

	sqlStmt :=
	`INSERT INTO jobs (id, runtime, tags, parameters, datasetsize)
	VALUES ($1, $2, $3, $4, $5)`



	_, err := db.Exec(sqlStmt, job.Id, job.Runtime, job.Tags, job.Parameters, job.InputDataSize)
	if err != nil {
		return err
	}
	return nil
}