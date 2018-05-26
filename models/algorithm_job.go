package models

import (
	"database/sql"
	"github.com/tteige/uit-go/autoscale"
)

func InsertAlgorithmJob(db *sql.DB, job autoscale.AlgorithmJob, runName string) error {
	_, err := db.Exec("INSERT INTO algorithm_job (run_name, jobid, created, started, executiontime, tag, deadline, priority, state) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		runName, job.Id, job.Created, job.Started, job.ExecutionTime, job.Tag, job.Deadline, job.Priority, job.State)
	if err != nil {
		return err
	}
	return nil
}

func GetAllAlgorithmJobs(db *sql.DB, runName string) ([]autoscale.AlgorithmJob, error) {
	rows, err := db.Query("SELECT (jobid, created, started, executiontime, tag, deadline, priority, state) FROM algorithm_job WHERE run_name = $1", runName)
	if err != nil {
		return nil, err
	}
	var jobs []autoscale.AlgorithmJob
	for rows.Next() {
		var job autoscale.AlgorithmJob
		err = rows.Scan(&job.Id, &job.Created, &job.Started, &job.ExecutionTime, &job.Tag, &job.Deadline, &job.Priority, &job.State)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
