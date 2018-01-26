package cloud

type Instance struct {
	Id string
	Type string
	Cluster string
}

type InstanceType struct {
	Id string
}

type Cluster struct {
	Name string
	Limit float32
	AcceptTags []string
}

type Cloud interface {
	AddInstance() (string, error)
	DeleteInstance(string) error
	GetInstances() ([]Instance, error)
	GetInstanceTypes() ([]InstanceType, error)
}