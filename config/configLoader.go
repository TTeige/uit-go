package config

import (
	"os"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config interface {
	LoadConfig() error
}

type FullConfig struct {
	DBConfig      DatabaseConfig
	ServiceConfig ServiceConfig
}

type DatabaseConfig struct {
	Password string `yaml:"pw"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
}

type ServiceConfig struct {
	Hostname string `yaml:"host"`
	Port     string `yaml:"port"`
}

func (conf *DatabaseConfig) LoadConfig() error {
	configLocation := os.Getenv("DATABASE_CONFIG_AUTOSCALE")
	configData, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configData, conf)
	if err != nil {
		return err
	}
	return nil
}

func (conf *ServiceConfig) LoadConfig() error {
	configLocation := os.Getenv("SERVICE_CONFIG_AUTOSCALE")
	configData, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configData, conf)
	if err != nil {
		return err
	}
	return nil
}

func (conf *FullConfig) LoadConfig() error {
	err := conf.DBConfig.LoadConfig()
	if err != nil {
		return err
	}
	err = conf.ServiceConfig.LoadConfig()
	if err != nil {
		return err
	}
	return nil
}
