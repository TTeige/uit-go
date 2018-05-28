package autoscalingService

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"database/sql"
)

type CPouta struct {
	Cluster autoscale.Cluster
	DB *sql.DB
}

func (*CPouta) Authenticate() error {
	panic("implement me")
}

func (*CPouta) SetScalingId(id string) error {
	panic("implement me")
}

func (*CPouta) GetExpectedJobCost(job autoscale.AlgorithmJob, instanceType string, currentTime time.Time) float64 {
	panic("implement me")
}

func (*CPouta) AddInstance(instance *autoscale.Instance, currentTime time.Time) (string, error) {
	panic("implement me")
}

func (*CPouta) DeleteInstance(id string, currentTime time.Time) error {
	panic("implement me")
}

func (*CPouta) GetInstances() ([]autoscale.Instance, error) {
	panic("implement me")
}

func (*CPouta) GetInstanceTypes() (map[string]autoscale.InstanceType, error) {
	panic("implement me")
}

func (*CPouta) GetInstanceLimit() int {
	panic("implement me")
}

func (*CPouta) GetTotalDuration(queue []autoscale.AlgorithmJob, currentTime time.Time) (int64, error) {
	panic("implement me")
}

func (*CPouta) GetTotalCost(queue []autoscale.AlgorithmJob, currentTime time.Time) float64 {
	panic("implement me")
}
