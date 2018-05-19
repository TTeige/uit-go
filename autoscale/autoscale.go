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
	Id            string
	Tag           string
	Parameters    []string
	State         string
	Priority      int
	ExecutionTime int64
	Deadline      time.Time
	Created       time.Time
}

type InputFas struct {
	Url string `json:"url"`
}

type dataUrl struct {
	InputFas InputFas `json:"input.fas"`
}

type Attempt struct {
	ExecutorId          string  `json:"executorId"`
	State               string  `json:"state"`
	AttemptId           string  `json:"attemptId"`
	Tag                 string  `json:"tag"`
	TimeCreated         string  `json:"timeCreated"`
	TimeStarted         string  `json:"timeStarted"`
	TimeEnded           string  `json:"timeEnded"`
	LastHeartbeat       string  `json:"lastHeartbeat"`
	RuntimeMillis       int     `json:"runtimeMillis"`
	QueueDurationMillis int     `json:"queueDurationMillis"`
	Outputs             dataUrl `json:"outputs"`
	Priority            int     `json:"priority"`
}

type MetapipeParameter struct {
	InputContigsCutoff     int  `json:"inputContigsCutoff"`
	UseBlastUniref50       bool `json:"useBlastUniref50"`
	UseInterproScan5       bool `json:"useInterproScan5"`
	UsePriam               bool `json:"usePriam"`
	RemoveNonCompleteGenes bool `json:"removeNonCompleteGenes"`
	ExportMergedGenbank    bool `json:"exportMergedGenbank"`
	UseBlastMarRef         bool `json:"useBlastMarRef"`
}

type MetapipeJob struct {
	Id                       string            `json:"jobId"`
	TimeSubmitted            string            `json:"timeSubmitted"`
	State                    string            `json:"state"`
	UserId                   string            `json:"userId"`
	Tag                      string            `json:"tag"`
	Priority                 int               `json:"priority"`
	Hold                     bool              `json:"hold"`
	Parameters               MetapipeParameter `json:"parameters"`
	Inputs                   dataUrl           `json:"inputs"`
	Outputs                  dataUrl           `json:"outputs"`
	TotalRuntimeMillis       int64             `json:"totalRuntimeMillis"`
	TotalQueueDurationMillis int64             `json:"totalQueueDurationMillis"`
	Attempts                 []Attempt         `json:"attempts"`
}

type Algorithm interface {
	Run(input AlgorithmInput, stepTime time.Time) (AlgorithmOutput, error)
}

type Estimator interface {
	Init() error
	ProcessQueue(jobs []MetapipeJob) ([]AlgorithmJob, error)
}

type Cloud interface {
	Authenticate() error
	SetScalingId(id string) error
	AddInstance(instance *Instance) (string, error)
	DeleteInstance(id string) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() (map[string]InstanceType, error)
	GetInstanceLimit() int
}
