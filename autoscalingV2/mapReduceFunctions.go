package autoscalingV2

import (
	"log"
	"encoding/json"
	"os"
)
//Create a single valued input channel
func createClosedInputChannel(obj interface{}) chan interface{} {
	newChan := make(chan interface{})
	go func() {
		newChan <- obj
		close(newChan)
	}()
	return newChan
}

func parseJsonObject(filename string) chan Job {
	output := make(chan Job)
	go func() {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer file.Close()
		dec := json.NewDecoder(file)
		_, err = dec.Token()
		if err != nil {
			log.Fatal(err)
		}
		for dec.More() {
			var job Job
			err = dec.Decode(&job)
			if err != nil {
				log.Fatal(err)
			}
			output <- job
		}
		_, err = dec.Token()
		if err != nil {
			log.Fatal(err)
		}
		close(output)
	}()
	return output
}

//BEGIN COUNT TAGS
func mapTags(filename interface{}, output chan interface{}) {
	results := map[string]int{}
	for job := range parseJsonObject(filename.(string)) {
		for i := 0; i < len(job.Tags); i++ {
			key := job.Tags[i]
			previousCount, exists := results[key]
			if !exists {
				results[key] = 1
			} else {
				results[key] = previousCount + 1
			}
		}
	}
	output <- results
}

func countTags(input chan interface{}, output chan interface{}) {
	results := map[string]int{}
	for matches := range input {
		for key, value := range matches.(map[string]int) {
			_, exists := results[key]
			if !exists {
				results[key] = value
			} else {
				results[key] = results[key] + value
			}
		}
	}
	output <- results
}
//END COUNT TAGS

//BEGIN COUNT JOBS
func mapJobsById(filename interface{}, output chan interface{}) {
	results := make(map[string]Job)

	for job := range parseJsonObject(filename.(string)) {
		key := job.Id
		results[key] = job
	}
	output <- results
}

func sumJobs(input chan interface{}, output chan interface{}) {
	results := 0
	for i := range input {
		for _, _ = range i.(map[string]Job) {
			results++
		}
	}

	output <- results
}
//END COUNT JOBS
//BEGIN REDUCE BY DURATION GROUP BY DATA SIZE
func mapByDataSize(filename interface{}, output chan interface{}) {
	results := make(map[float64][]Job)

	for job := range parseJsonObject(filename.(string)) {
		key := job.DataSetSize
		results[key] = append(results[key], job)
	}
	output <- results
}

func sumDuration(input chan interface{}, output chan interface{}) {
	results := map[float64]int64{}
	for matches := range input {
		for key, value := range matches.(map[float64][]Job) {
			var duration int64
			for i := range value {
				duration = duration + value[i].Duration
			}
			_, exists := results[key]
			if !exists {
				results[key] = duration
			} else {
				results[key] = results[key] + duration
			}
		}
	}
	output <- results
}

func simple(input chan interface{}, output chan interface{}) {
	var val int
	for matches := range input {
		val = matches.(int) + 10
	}
	output <- val
}
//END REDUCE BY DURATION GROUP BY DATA SIZE

