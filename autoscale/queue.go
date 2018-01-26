package autoscale

import "github.com/tteige/uit-go/models"

type QueueHandle struct {

}

func (q *QueueHandle) PredictJobs(jobs MetapipeJob) ([]models.Job, error) {
	return nil, nil
}

func (q *QueueHandle) GetJobs() []models.Job {
	return nil
}