package autoscale

import (
	"github.com/tteige/uit-go/models"
	"database/sql"
)

type QueueHandle struct {
	DB *sql.DB
}

func (q *QueueHandle) PredictQueue(jobs MetapipeJob) ([]AlgorithmJob, error) {
	// Predict the jobs based on the database, conserve the MetapipeJob priority and etc
	return nil, nil
}

func (q *QueueHandle) GetJobs() []models.Job {
	// Function not needed?
	return nil
}