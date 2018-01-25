package simulator

import "database/sql"

func Run(hostUrl string, db *sql.DB) {

}

//
//import (
//	"github.com/tteige/uit-go/mapReduce"
//	"runtime"
//	"log"
//)
//
//type NodeConfig struct {
//	InstanceType string
//	VCpu         int
//	Memory       float64
//	ClockSpeed   float64
//	Storage
//}
//
//type Storage struct {
//	Type  string
//	Count int
//	Size  int
//}
//
//type ClusterConfig struct {
//	Nodes []NodeConfig
//}
//
//type Job struct {
//	Id          string `json:"id"`
//	Duration    int64 `json:"duration"`
//	DataSetSize float64 `json:"dataSetSize"`
//	Tags        []string `json:"tags"`
//}
//
//func parseHistoricalData(filename string) interface{} {
//	fileChan := createClosedInputChannel(filename)
//	maxWorkers := runtime.GOMAXPROCS(runtime.NumCPU())
//	mapreducer := mapReduce.New(maxWorkers)
//	mapreducer.Map(mapTags)
//	mapreducer.Reduce(countTags)
//	result := mapreducer.Run(fileChan)
//
//	//log.Printf("Job count: %d\n", result)
//	for k, v := range result.(map[string]int) {
//		log.Printf("Tag: %s. count: %d\n", k, v)
//	}
//	//for k, v := range result.(map[string]Job) {
//	//	log.Printf("JobId: %s = Job %+v\n", k, v)
//	//}
//	return result
//}
//
//func RunSimulation(historicalDataFile string, currentQueueFile string, config ClusterConfig, seed int64) (int, error) {
//	_ = parseHistoricalData(historicalDataFile)
//	return 0, nil
//}
