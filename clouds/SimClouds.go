package clouds

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"database/sql"
	"github.com/tteige/uit-go/models"
	"github.com/segmentio/ksuid"
)

type SimCloud struct {
	Cluster       autoscale.Cluster
	Db            *sql.DB
	runId         string
	lastIteration time.Time
	beginTime     time.Time
}

func (c *SimCloud) GetCurrentAvailableFunds() float64 {
	return c.Cluster.CostLimit - c.Cluster.MoneyUsed
}

func (c *SimCloud) GetCostLimit() float64 {
	return c.Cluster.CostLimit
}

func (c *SimCloud) GetExpectedJobCost(instanceType string, execTime int64) float64 {
	cost := c.Cluster.Types[instanceType].PriceIncrement * float64(execTime/int64(time.Hour))
	return cost
}

func (c *SimCloud) SetScalingId(id string) error {
	c.runId = id
	sim, err := models.GetSimulation(c.Db, id)
	if err != nil {
		return err
	}
	c.beginTime = sim.Started
	return nil
}

func (c *SimCloud) Authenticate() error {
	return nil
}

func (c *SimCloud) GetInstanceLimit() int {
	return c.Cluster.Limit
}

func (c *SimCloud) AddInstance(instance *autoscale.Instance) (string, error) {

	eventType := "CREATED"
	instance.State = "STARTING"
	reusedIndex := -1
	for i, inst := range c.Cluster.ActiveInstances {
		if inst.State == "INACTIVE" && inst.Type == instance.Type {
			instance.Id = inst.Id
			instance.State = inst.State
			eventType = "REUSED"
			reusedIndex = i
			break
		}
	}

	if instance.Id == "" {
		instance.Id = ksuid.New().String()
	}
	err := models.WriteSimEvent(c.Db, models.SimEvent{
		SimId:        c.runId,
		Created:      time.Now(),
		Instance:     *instance,
		InstanceType: c.Cluster.Types[instance.Type],
		Type:         eventType,
	})
	instance.State = "ACTIVE"
	if err != nil {
		return "", err
	}

	if eventType == "CREATED" {
		c.Cluster.ActiveInstances = append(c.Cluster.ActiveInstances, *instance)
	} else if eventType == "REUSED" {
		c.Cluster.ActiveInstances[reusedIndex] = *instance
	}

	return instance.Id, nil
}

func (c *SimCloud) DeleteInstance(id string) error {
	instances, err := c.GetInstances()
	if err != nil {
		return nil
	}
	for i, e := range instances {
		if e.Id == id {
			c.Cluster.ActiveInstances = append(instances[:i], instances[i+1:]...)
			models.WriteSimEvent(c.Db, models.SimEvent{
				SimId:        c.runId,
				Created:      time.Now(),
				Instance:     e,
				InstanceType: c.Cluster.Types[e.Type],
				Type:         "DELETED",
			})
			break
		}
	}
	return nil
}

func (c *SimCloud) GetInstances() ([]autoscale.Instance, error) {
	return c.Cluster.ActiveInstances, nil
}

func (c *SimCloud) GetInstanceTypes() (map[string]autoscale.InstanceType, error) {
	return c.Cluster.Types, nil
}
