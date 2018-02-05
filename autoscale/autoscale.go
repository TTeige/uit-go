package autoscale

import (
	"time"
	"net/url"
)

type Instance struct {
	Id   string
	Type InstanceType
}

type InstanceType struct {
	Id             string
	PriceIncrement int
}

type Cluster struct {
	Name            string
	Limit           int
	AcceptTags      []string
	Types           []InstanceType
	ActiveInstances []Instance
}

type AlgorithmInput struct {
	JobQueue []AlgorithmJob
	Clusters []Cluster
}

type AlgorithmOutput struct {
	Instances []Instance
}

type AlgorithmJob struct {
	Id                   string
	Tags                 []string
	Parameters           []string
	State                string
	StateTransitionTimes []time.Time
	Deadline             time.Time
	Priority             int
}

type MetapipeJob struct {
	Job   AlgorithmJob
	Input url.URL
}

type Cloud interface {
	AddInstance(instance *Instance) (string, error)
	DeleteInstance(id string) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() ([]InstanceType, error)
}

type Algorithm interface {
	Step(input AlgorithmInput, stepTime time.Time) (*AlgorithmOutput, error)
}
