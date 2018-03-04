package clouds

import (
	"github.com/tteige/uit-go/autoscale"
	"math/rand"
	"time"
)

// DOCS for openstack golang http://gophercloud.io/docs/compute/

type SimCpouta struct {
	autoscale.Cluster
}

func (c *SimCpouta) AddInstance(instance *autoscale.Instance) (string, error) {
	active, err := c.GetInstances()
	//create new server if none are up and available
	//if a server is up and available, resize it
	//return the new id of the instance
	// Gives a range of 100-500 milisec delay for fetching available servers

	waitMilli := rand.Int31n(400) + 100
	time.Sleep(time.Duration(waitMilli)* time.Millisecond)
}

func (c *SimCpouta) DeleteInstance(id string) error {
	panic("implement me")
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
			// "s" will be a servers.Server
		}
	})
	*/
	waitMilli := rand.Int31n(400) + 100
	time.Sleep(time.Duration(waitMilli)* time.Millisecond)
	return c.ActiveInstances, nil
}

func (c *SimCpouta) GetInstanceTypes() ([]autoscale.InstanceType, error) {
	panic("implement me")
}
