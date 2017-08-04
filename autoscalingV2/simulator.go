package autoscalingV2

import (
	"time"
	"github.com/tteige/uit-go/mapReduce"
	"log"
	"runtime"
	"encoding/json"
	"strings"
	"io"
)

type NodeConfig struct {
	InstanceType string
	VCpu         int
	Memory       float64
	ClockSpeed   float64
	Storage
}

type Storage struct {
	Type  string
	Count int
	Size  int
}

type ClusterConfig struct {
	Nodes []NodeConfig
}

type Job struct {
	Id          string
	Duration    time.Duration
	DataSetSize float64
	Tags        []string
}

func RunSimulation(historicalData []Job, currentQueue []Job, config ClusterConfig, seed int64) (int, error) {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	type Tmp struct {
		Request struct{
			Key string `json:"key"`
		}
	}

	mapReducer := mapReduce.NewMapReducer(func(filename interface{}, output chan interface{}) {
		results := map[interface{}]int{}

		for line := range mapReduce.EnumerateJSON(filename.(string)) {
			dec := json.NewDecoder(strings.NewReader(line))
			var iFace Tmp
			if err := dec.Decode(&iFace); err == io.EOF {
				continue
			} else if err != nil {
				continue
			}
			previousCount, exists := results[iFace]
			if !exists {
				results[iFace] = 1
			} else {
				results[iFace] = previousCount + 1
			}
		}

		output <- results
	}, func(input chan interface{}, output chan interface{}) {
		results := map[Tmp]int{}
		for matches := range input {
			for key, value := range matches.(map[Tmp]int) {
				_, exists := results[key]
				if !exists {
					results[key] = value
				} else {
					results[key] = results[key] + value
				}
			}
		}
		output <- results
	}, runtime.GOMAXPROCS(runtime.NumCPU()))
	fileChan := make(chan interface{})
	go func() {
		fileChan <- "autoscalingV2/stuff.json"
		close(fileChan)
	}()
	result := mapReducer.MapReduce(fileChan)
	type tmp struct {
		Key string `json:"key"`
	}
	for k, v := range result.(map[Tmp]int) {
		log.Printf("%s: %d\n", k, v)
	}
	return 0, nil
}
