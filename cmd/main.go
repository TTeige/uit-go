package main

import (
	"encoding/json"
	"flag"
	"github.com/tteige/uit-go/algorithm"
	"github.com/tteige/uit-go/autoscale"
	"github.com/tteige/uit-go/autoscalingService"
	"github.com/tteige/uit-go/config"
	"github.com/tteige/uit-go/estimator"
	"github.com/tteige/uit-go/metapipe"
	"github.com/tteige/uit-go/models"
	"github.com/tteige/uit-go/simulator"
	"log"
	"os"
)

func main() {

	service := flag.Bool("production", false, "run the auto scaling service")
	updateDb := flag.Bool("updateDB", false, "update the database on launch")

	conf := config.FullConfig{}
	err := conf.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	serviceHostname := conf.ServiceConfig.Hostname + ":" + conf.ServiceConfig.Port

	db, err := models.OpenDatabase(conf.DBConfig.User, conf.DBConfig.Host, conf.DBConfig.Password)
	if err != nil {
		log.Fatal(err)
		return
	}

	auth := metapipe.Oath2{
		User:     conf.OAuthConf.Username,
		Password: conf.OAuthConf.ClientSecret,
	}

	_, err = auth.GetSetAccessToken()

	if err != nil {
		log.Fatal(err)
		return
	}
	// TODO: Reinit database periodically, spawn a job that does this every day
	err = models.InitDatabase(db, auth, *updateDb)
	if err != nil {
		log.Fatal(err)
		return
	}

	est := estimator.LinearRegression{
		Auth: auth,
		DB:   db,
	}

	if *service {
		log.Printf("Starting the auto scaling service at: %s ", serviceHostname)
		s := autoscalingService.Service{
			DB:       db,
			Hostname: serviceHostname,
		}
		s.Run()
	} else {
		alg := algorithm.NaiveAlgorithm{}

		simClusterMap, err := loadClouds()
		if err != nil {
			log.Fatal(err)
			return
		}

		sim := simulator.Simulator{
			DB:          db,
			Hostname:    serviceHostname,
			Algorithm:   alg,
			Log:         log.New(os.Stdout, "SIMULATOR LOGGER: ", log.Lshortfile|log.LstdFlags),
			Estimator:   &est,
			SimClusters: simClusterMap,
		}
		sim.Run()
	}
	return
}

func loadClouds() (autoscale.ClusterCollection, error) {
	simClusterMap := make(autoscale.ClusterCollection)

	configLocation := os.Getenv("SIM_CLUSTER_CONFIG")
	reader, err := os.Open(configLocation)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(reader)
	err = dec.Decode(&simClusterMap)
	if err != nil {
		return nil, err
	}

	configLocation = os.Getenv("SIM_INSTANCE_FLAVOURS_CONFIG")
	if configLocation == "" {
		return simClusterMap, nil
	}
	reader, err = os.Open(configLocation)
	if err != nil {
		return nil, err
	}

	flavours := make(map[string]map[string]autoscale.InstanceType)

	dec = json.NewDecoder(reader)
	err = dec.Decode(&flavours)
	if err != nil {
		return nil, err
	}

	for k, v := range flavours {
		c := simClusterMap[k]
		c.Types = v
		simClusterMap[k] = c
	}

	return simClusterMap, nil
}
