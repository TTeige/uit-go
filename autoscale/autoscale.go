package autoscale

import (
	"time"
	"net/url"
	"github.com/tteige/uit-go/models"
)

type Instance struct {
	Id      string
	Type    string
	Cluster string
}

type InstanceType struct {
	Id string
}

type Cluster struct {
	Name       string
	Limit      float32
	AcceptTags []string
}

type AlgorithmInput struct {
	JobQueue []AlgorithmJob
}

type AlgorithmOutput struct {
	Instances []Instance
}

type AlgorithmJob struct {
	Job      models.Job
	Deadline time.Time
	Priority int
}

type MetapipeJob struct {
	Id         string
	Input      url.URL
	Tags       []string
	Parameters []string
	State      string
	Intervals  []time.Time
	Priority   int
}

type Cloud interface {
	AddInstance(instance *Instance) (string, error)
	DeleteInstance(id string) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() ([]InstanceType, error)
}

type AlgorithmInterface interface {
	Step(input AlgorithmInput, stepTime time.Time) (*AlgorithmOutput, error)
}
