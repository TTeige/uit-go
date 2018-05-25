package autoscale

import (
	"time"
)

const (
	AWS    = "aws"
	Stallo = "metapipe"
	CPouta = "csc"
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
	CostLimit       float64                 `json:"cost_limit"`
	MoneyUsed       float64                 `json:"money_used"`
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
	Id            string
	Tag           string
	Parameters    JobParameters
	State         string
	Priority      int
	ExecutionTime []int64
	Deadline      time.Time
	Created       time.Time
	Started       time.Time
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
	GetCostLimit() float64
	GetCurrentAvailableFunds() float64
	GetExpectedJobCost(instanceType string, execTime int64) float64
	AddInstance(instance *Instance, currentTime time.Time) (string, error)
	DeleteInstance(id string, currentTime time.Time) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() (map[string]InstanceType, error)
	GetInstanceLimit() int
}
