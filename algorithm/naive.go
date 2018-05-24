package algorithm

import (
	"time"
	"github.com/tteige/uit-go/autoscale"
	"sort"
)

type NaiveAlgorithm struct {
}

func (n NaiveAlgorithm) Run(input autoscale.AlgorithmInput, startTime time.Time) (autoscale.AlgorithmOutput, error) {
	queueMap := make(map[string][]autoscale.AlgorithmJob)
	var outInstances []autoscale.Instance
	var out autoscale.AlgorithmOutput
	for _, j := range input.JobQueue {
		queueMap[j.Tag] = append(queueMap[j.Tag], j)
	}

	for key, queue := range queueMap {
		curClust := input.Clouds[key]
		instances, err := curClust.GetInstances()
		if err != nil {
			return out, err
		}
		sort.Slice(queue, func(i, j int) bool {
			if queue[i].Priority > queue[j].Priority {
				return true
			} else if queue[i].Priority == queue[j].Priority {
				return queue[i].Deadline.Before(queue[j].Deadline)
			}
			return false
		})

		runningJobs := 0
		for _, job := range queue {
			if job.State == "RUNNING" {
				runningJobs++
				continue
			}
			instances, err := curClust.GetInstances()
			if err != nil {
				return out, err
			}
			if curClust.GetInstanceLimit() > len(instances) {
				if len(instances) == runningJobs {
					break
				}
				types, err := curClust.GetInstanceTypes()
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				var iType autoscale.InstanceType
				if key == autoscale.AWS {
					iType = types["c5-4xl"]
				} else if key == autoscale.Stallo {
					iType = types["default"]
				} else if key == autoscale.CPouta {
					iType = types["default"]
				}
				instance := autoscale.Instance{
					Id:    "",
					Type:  iType.Name,
					State: "",
				}
				_, err = curClust.AddInstance(&instance, startTime)
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				outInstances = append(outInstances, instance)
			}
		}
		//DELETE instances based on the queue size, if the instance size is larger than queue,
		//there should be at least one INACTIVE instance
		if len(instances) > (len(queue)) {
			for _, i := range instances {
				if i.State == "INACTIVE" {
					err = curClust.DeleteInstance(i.Id, startTime)
					if err != nil {
						return autoscale.AlgorithmOutput{}, err
					}
					outInstances = append(outInstances, i)
				}
			}
		}

		queueMap[key] = queue
	}

	var outQueue []autoscale.AlgorithmJob

	for _, m := range queueMap {
		for _, j := range m {
			outQueue = append(outQueue, j)
		}
	}

	out = autoscale.AlgorithmOutput{
		Instances: outInstances,
		JobQueue:  outQueue,
	}

	return out, nil
}
