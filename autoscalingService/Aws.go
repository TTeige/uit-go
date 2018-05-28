package autoscalingService

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"database/sql"
)

type Aws struct {
	Cluster autoscale.Cluster
	DB *sql.DB
}

func (*Aws) Authenticate() error {
	panic("implement me")
}

func (*Aws) SetScalingId(id string) error {
	panic("implement me")
}

func (*Aws) GetExpectedJobCost(job autoscale.AlgorithmJob, instanceType string, currentTime time.Time) float64 {
	panic("implement me")
}

func (*Aws) AddInstance(instance *autoscale.Instance, currentTime time.Time) (string, error) {
	panic("implement me")
}

func (*Aws) DeleteInstance(id string, currentTime time.Time) error {
	panic("implement me")
}

func (*Aws) GetInstances() ([]autoscale.Instance, error) {
	panic("implement me")
}

func (*Aws) GetInstanceTypes() (map[string]autoscale.InstanceType, error) {
	panic("implement me")
}

func (*Aws) GetInstanceLimit() int {
	panic("implement me")
}

func (*Aws) GetTotalDuration(queue []autoscale.AlgorithmJob, currentTime time.Time) (int64, error) {
	panic("implement me")
}

func (*Aws) GetTotalCost(queue []autoscale.AlgorithmJob, currentTime time.Time) float64 {
	panic("implement me")
}
