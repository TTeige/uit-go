package clouds

import (
	"github.com/tteige/uit-go/autoscale"
	"math/rand"
	"time"
	"database/sql"
	"github.com/tteige/uit-go/models"
)

// DOCS for openstack golang http://gophercloud.io/docs/compute/

type SimCpouta struct {
	autoscale.Cluster
	// Needs a database handle to interact with the "external compute system". The database emulates the network
	// connection
	Db *sql.DB
	Simulation models.Simulation
}

func (c *SimCpouta) AddInstance(instance *autoscale.Instance) (string, error) {
	available, err := c.GetInstances()
	newInstanceName := ""
	if err != nil {
		return newInstanceName, err
	}
	//create new server if none are up and available
	//Assume 1 server is up and available, 2 servers are up and unavailable
	//if a server is up and available, resize it

	//return the new id of the instance
	// Gives a range of 100-500 milisec delay for fetching available servers

	for _, i := range available {
		// Check for an available VM / server first
		if i.Type.Name == instance.Type.Name {
			// The correct type is found and is available for usage.
			if i.State == "Shutoff" {
				// The server / VM is in the correct state and has the correct flavor
				models.WriteSimEvent(c.Db, models.SimEvent{
					SimId:      c.Simulation.Name,
					Created:    time.Now(),//Needs to be changed to support the time abstraction
					InstanceId: i.Id,
					Type:       "reboot",
				})
				newInstanceName = i.Id
				i.State = "Active"
				return newInstanceName, nil
			}
		}
	}

	stall_for_feedback(100, 500)
	return newInstanceName, nil
}

func (c *SimCpouta) DeleteInstance(id string) error {
	// Simulates fetching the instances, needs to be done since it can have changed since the last time
	instances, err := models.GetInstances(c.Db, c.Name)
	if err != nil {
		return nil
	}
	for i, e := range instances {
		if e.Id == id {
			c.ActiveInstances = append(instances[:i], instances[i+1:]...)
			models.WriteSimEvent(c.Db, models.SimEvent{
				SimId:      c.Simulation.Name,
				Created:    time.Now(),//Needs to be changed to support the time abstraction
				InstanceId: e.Id,
				Type:       "delete",
			})
			break
		}
	}
	return nil
}

func (c *SimCpouta) GetInstances() ([]autoscale.Instance, error) {
	// simulates the call to get all instances from cPouta, but just fetches the active instances at runtime
	/*
	// We have the option of filtering the server list. If we want the full
	// collection, leave it as an empty struct
	opts := servers.ListOpts{Name: "server_1"}

	// Retrieve a pager (i.e. a paginated collection)
	pager := servers.List(client, opts)

	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)

		for _, s := range serverList {
			server, id
		}
	})
	*/
	if c.ActiveInstances == nil {
		// Only load static data from the database once
		i, err := models.GetInstances(c.Db, "cpouta")
		if err != nil {
			return nil, err
		}
		c.ActiveInstances = i
	}
	stall_for_feedback(100, 500)
	return c.ActiveInstances, nil
}

func (c *SimCpouta) GetInstanceTypes() (map[string]autoscale.InstanceType, error) {
	// Request to fetch all flavors, not sure if any API support this
	stall_for_feedback(100, 500)
	return c.Types, nil
}

func stall_for_feedback(min int32, max int32) {
	//This should add to the global time counter which is curated manually
	waitMilli := rand.Int31n(max-min) + min
	time.Sleep(time.Duration(waitMilli) * time.Millisecond)
}
