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
