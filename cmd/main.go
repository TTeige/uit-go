package main

import (
	"flag"
	"log"
	"github.com/tteige/uit-go/autoscalingService"
	"github.com/tteige/uit-go/models"
	"github.com/tteige/uit-go/simulator"
)

func main() {

	service := flag.Bool("production", false, "run the auto scaling service")
	dbHost := flag.String("databaseUrl", "user=tim dbname=autoscaling password=something", "database source")
	hostname := flag.String("hostname", "localhost:8080", "hostname for either service or simulator")

	db, err := models.OpenDatabase(*dbHost)
	if err != nil {
		log.Fatalf("%s", err)
		return
	}

	if *service {
		log.Printf("Starting the auto scaling service at: %s ", *hostname)
		autoscalingService.Run(*hostname, db)
	} else {
		log.Printf("Starting the auto scaling simulator at: %s ", *hostname)
		sim := simulator.Simulator{
			DB:db,
			Hostname:*hostname,
		}
		sim.Run()
	}
	return
}
