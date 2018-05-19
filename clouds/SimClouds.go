package clouds

import (
	"github.com/tteige/uit-go/autoscale"
	"math/rand"
	"time"
	"database/sql"
	"github.com/tteige/uit-go/models"
	"github.com/segmentio/ksuid"
)

type SimCloud struct {
	Cluster autoscale.Cluster
	Db            *sql.DB
	runId         string
	lastIteration time.Time
	beginTime     time.Time
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
	reusedIndex := -1
	for i, inst := range c.Cluster.ActiveInstances {
		if inst.State == "INACTIVE" && inst.Type == instance.Type {
			instance.Id = inst.Id
			eventType = "REUSED"
			reusedIndex = i
			break
		}
	}

	if instance.Id == "" {
		instance.Id = ksuid.New().String()
	}
	instance.State = "Active"
	err := models.WriteSimEvent(c.Db, models.SimEvent{
		SimId:        c.runId,
		Created:      time.Now(),
		Instance:     *instance,
		InstanceType: c.Cluster.Types[instance.Type],
		Type:         eventType,
	})
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
				Type:         "delete",
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
