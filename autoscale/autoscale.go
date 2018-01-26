package autoscale

import (
	"time"
	"net/url"
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
	JobQueue QueueHandle
}

type AlgorithmOutput struct {
	Instances []Instance
}

type MetapipeJob struct {
	Id         string
	Input      url.URL
	Tags       []string
	Parameters []string
	State      string
	Intervals  []time.Time
}

type Cloud interface {
	AddInstance(instance *Instance) (string, error)
	DeleteInstance(id string) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() ([]InstanceType, error)
}

type AlgorithmInterface interface {
	Step(input AlgorithmInput, stepTime time.Time) (error, AlgorithmOutput)
}
