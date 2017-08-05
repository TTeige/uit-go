package autoscalingV2

import (
	"log"
	"encoding/json"
	"os"
)

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

func mapJobs(filename interface{}, output chan interface{}) {
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
