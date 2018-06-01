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
	"database/sql"
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

	log.Printf("Run service %+v\nUpdateDB %+v", *service, *updateDb)

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

	alg := algorithm.NaiveAlgorithm{}
	//alg := algorithm.BadAlgorithm{}
	//alg := algorithm.NilAlg{}

	if *service {
		clusters, err := loadClusters("CLUSTER_CONFIG")
		clouds, err := createMetapipeClouds(db, clusters)
		if err != nil {
			return
		}
		log.Printf("Starting the auto scaling service at: %s ", serviceHostname)
		s := autoscalingService.Service{
			DB:        db,
			Hostname:  serviceHostname,
			Clouds:    clouds,
			Algorithm: alg,
			Log:       log.New(os.Stdout, "AUTOSCALE LOGGER: ", log.Lshortfile|log.LstdFlags),
			Estimator: &est,
		}
		s.Run()
	} else {
		simClusterMap, err := loadClusters("SIM_CLUSTER_CONFIG")
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

func createMetapipeClouds(db *sql.DB, clusters autoscale.ClusterCollection) (autoscale.CloudCollection, error) {
	simCloudMap := make(autoscale.CloudCollection)

	simCloudMap[metapipe.CPouta] = &autoscalingService.CPouta{
		Cluster: clusters[metapipe.CPouta],
		DB:      db,
	}
	simCloudMap[metapipe.AWS] = &autoscalingService.Aws{
		Cluster: clusters[metapipe.AWS],
		DB:      db,
	}
	simCloudMap[metapipe.Stallo] = &autoscalingService.Stallo{
		Cluster: clusters[metapipe.Stallo],
		DB:      db,
	}
	return simCloudMap, nil
}

func loadClusters(configEnvName string) (autoscale.ClusterCollection, error) {
	simClusterMap := make(autoscale.ClusterCollection)

	configLocation := os.Getenv(configEnvName)
	reader, err := os.Open(configLocation)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(reader)
	err = dec.Decode(&simClusterMap)
	if err != nil {
		return nil, err
	}

	return simClusterMap, nil
}
