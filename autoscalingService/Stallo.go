package autoscalingService

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"database/sql"
)

type Stallo struct {
	Cluster autoscale.Cluster
	DB *sql.DB
}

func (*Stallo) Authenticate() error {
	panic("implement me")
}

func (*Stallo) SetScalingId(id string) error {
	panic("implement me")
}

func (*Stallo) GetExpectedJobCost(job autoscale.AlgorithmJob, instanceType string, currentTime time.Time) float64 {
	panic("implement me")
}

func (*Stallo) AddInstance(instance *autoscale.Instance, currentTime time.Time) (string, error) {
	panic("implement me")
}

func (*Stallo) DeleteInstance(id string, currentTime time.Time) error {
	panic("implement me")
}

func (*Stallo) GetInstances() ([]autoscale.Instance, error) {
	panic("implement me")
}

func (*Stallo) GetInstanceTypes() (map[string]autoscale.InstanceType, error) {
	panic("implement me")
}

func (*Stallo) GetInstanceLimit() int {
	panic("implement me")
}

func (*Stallo) GetTotalDuration(queue []autoscale.AlgorithmJob, currentTime time.Time) (int64, error) {
	panic("implement me")
}

func (*Stallo) GetTotalCost(queue []autoscale.AlgorithmJob, currentTime time.Time) float64 {
	panic("implement me")
}
