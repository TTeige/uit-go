package autoscalingV2

import (
	"github.com/tteige/uit-go/mapReduce"
	"log"
	"runtime"
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
	Id          string `json:"id"`
	Duration    int64 `json:"duration"`
	DataSetSize float64 `json:"dataSetSize"`
	Tags        []string `json:"tags"`
}

func parseHistoricalData(filename string) interface{} {
	maxWorkers := runtime.GOMAXPROCS(runtime.NumCPU())
	tagToCount := mapReduce.NewMapReducer(mapTags, countTags, maxWorkers)

	fileChan := make(chan interface{})
	go func() {
		fileChan <- filename
		close(fileChan)
	}()

	result := tagToCount.MapReduce(fileChan)
	for k, v := range result.(map[string]int) {
		log.Printf("%s: %d\n", k, v)
	}

	fileChan = make(chan interface{})
	go func() {
		fileChan <- filename
		close(fileChan)
	}()

	jobCount := mapReduce.NewMapReducer(mapJobsById, sumJobs, maxWorkers)
	result = jobCount.MapReduce(fileChan)
	log.Printf("Job count: %d\n", result)

	fileChan = make(chan interface{})
	go func() {
		fileChan <- filename
		close(fileChan)
	}()

	averageDuration := mapReduce.NewMapReducer(mapByDataSize, sumDuration, maxWorkers)
	result = averageDuration.MapReduce(fileChan)
	for k, v := range result.(map[float64]int64) {
		log.Printf("DataSize: %f Duration: %d\n", k, v)
	}
	return result
}

func RunSimulation(historicalDataFile string, currentQueueFile string, config ClusterConfig, seed int64) (int, error) {
	_ = parseHistoricalData(historicalDataFile)
	return 0, nil
}
