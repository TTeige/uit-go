package autoscale

import (
	"time"
)

type ClusterCollection map[string]Cluster
type CloudCollection map[string]Cloud
type JobParameters map[string]string

type Instance struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	State string `json:"state"`
}

type ScalingEvent struct {
	Instance   Instance
	Type       string
	ClusterTag string
}

type InstanceType struct {
	Name           string  `json:"name"`
	PriceIncrement float64 `json:"price"`
}

type Cluster struct {
	Name            string                  `json:"name"`
	Limit           int                     `json:"limit"`
	AcceptTag       string                  `json:"tag"`
	Types           map[string]InstanceType `json:"types"`
	ActiveInstances []Instance              `json:"instances"`
}

type AlgorithmInput struct {
	JobQueue []AlgorithmJob
	Clouds   CloudCollection
}

type AlgorithmOutput struct {
	Instances []Instance
	JobQueue  []AlgorithmJob
}

type AlgorithmJob struct {
	Id              string
	Tag             string
	Parameters      JobParameters
	State           string
	Priority        int
	ExecutionTime   map[string]int64
	Deadline        time.Time
	Created         time.Time
	Started         time.Time
	InstanceFlavour InstanceType
}

type Algorithm interface {
	Run(input AlgorithmInput, startTime time.Time) (AlgorithmOutput, error)
}

type Estimator interface {
	Init() error
	ProcessQueue(jobs []AlgorithmJob) ([]AlgorithmJob, error)
}

type Cloud interface {
	Authenticate() error
	SetScalingId(id string) error
	GetExpectedJobCost(job AlgorithmJob, instanceType string, currentTime time.Time) float64
	AddInstance(instance *Instance, currentTime time.Time) (string, error)
	DeleteInstance(id string, currentTime time.Time) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() (map[string]InstanceType, error)
	GetInstanceLimit() int
	GetTotalDuration(queue []AlgorithmJob, currentTime time.Time) (int64, error)
	GetTotalCost(queue []AlgorithmJob, currentTime time.Time) float64
}
