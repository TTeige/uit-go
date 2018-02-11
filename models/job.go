package models

import (
	"database/sql"
	"log"
)

type Job struct {
	Runtime       int64
	Tag           string
	InputDataSize int
	JobId         string
	QueueDuration int64
}

func CheckExists(db *sql.DB, jobId string) (bool, error) {
	existStmt :=
		`SELECT EXISTS(SELECT 1 FROM jobs WHERE jobid = $1)`

	_, err := db.Query(existStmt, jobId)

	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("Error when initializing the database: %s", err)
	}
	if err == sql.ErrNoRows {
		return false, err
	}
	return true, err
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
		err := rows.Scan(&job.JobId, &job.Runtime, &job.Tag)
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
		`INSERT INTO jobs (jobid, runtime, tag, datasetsize, queueduration)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (jobid)
		DO NOTHING`

	_, err := db.Exec(sqlStmt, job.JobId, job.Runtime, job.Tag, job.InputDataSize, job.QueueDuration)
	if err != nil {
		return err
	}
	return nil
}
