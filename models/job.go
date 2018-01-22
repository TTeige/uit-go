package models

import "database/sql"

type Job struct {
	Runtime int
	Id string
	Parameters []string
	Tags []string
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