package main

import (
	"flag"
	"log"
	"github.com/tteige/uit-go/autoscalingService"
	"github.com/tteige/uit-go/models"
	"github.com/tteige/uit-go/simulator"
	"github.com/tteige/uit-go/algorithm"
	"os"
	"github.com/tteige/uit-go/config"
	"github.com/tteige/uit-go/estimator"
	"github.com/tteige/uit-go/autoscale"
	"encoding/json"
	"github.com/tteige/uit-go/clouds"
)

func main() {

	service := flag.Bool("production", false, "run the auto scaling service")

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

	auth := autoscale.Oath2{
		User:     conf.OAuthConf.Username,
		Password: conf.OAuthConf.ClientSecret,
	}

	_, err = auth.GetSetAccessToken()

	if err != nil {
		log.Fatal(err)
		return
	}
	// TODO: Reinit database periodically, spawn a job that does this every day
	if false {
		err = models.InitDatabase(db, auth)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	est := estimator.LinearRegression{
		DB: db,
	}

	if *service {
		log.Printf("Starting the auto scaling service at: %s ", serviceHostname)
		s := autoscalingService.Service{
			DB:       db,
			Hostname: serviceHostname,
		}
		s.Run()
	} else {
		alg := algorithm.NaiveAlgorithm{
			ScaleUpThreshold:   10,
			ScaleDownThreshold: 3,
		}
		simClusterMap := make(autoscale.ClusterCollection)

		configLocation := os.Getenv("SIM_CLUSTER_CONFIG")
		reader, err := os.Open(configLocation)
		if err != nil {
			log.Fatal(err)
			return
		}

		dec := json.NewDecoder(reader)
		err = dec.Decode(&simClusterMap)
		if err != nil {
			log.Fatal(err)
		}

		simCloudMap := make(autoscale.CloudCollection)

		simCloudMap[autoscale.CPouta] = &clouds.SimCloud{
			Cluster: simClusterMap[autoscale.CPouta],
			Db:      db,
		}
		simCloudMap[autoscale.AWS] = &clouds.SimCloud{
			Cluster: simClusterMap[autoscale.AWS],
			Db:      db,
		}
		simCloudMap[autoscale.Stallo] = &clouds.SimCloud{
			Cluster: simClusterMap[autoscale.Stallo],
			Db:      db,
		}

		sim := simulator.Simulator{
			DB:        db,
			Hostname:  serviceHostname,
			Algorithm: alg,
			Log:       log.New(os.Stdout, "SIMULATOR LOGGER: ", log.Lshortfile|log.LstdFlags),
			Estimator: &est,
		}
		sim.Run()
	}
	return
}
