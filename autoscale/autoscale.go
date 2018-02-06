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

type BaseJob struct {
	Id         string
	Tags       []string
	Parameters []string
	State      string
	Priority   int
}

type AlgorithmJob struct {
	BaseJob
	StateTransitionTimes []time.Time
	Deadline             time.Time
}

type attempt struct {
	State       string
	Tag         string
	TimeCreated time.Time
	TimeStarted time.Time
	TimeEnded   time.Time
	Runtime     time.Duration
}

type MetapipeJob struct {
	BaseJob
	Input       url.URL
	StartTime   time.Time
	EndTime     time.Time
	TimeCreated time.Time
	Attempts    []attempt
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
