package algorithm

import (
	"time"
	"github.com/tteige/uit-go/autoscale"
	"sort"
	"math"
)

type NaiveAlgorithm struct {
}

func (n NaiveAlgorithm) Run(input autoscale.AlgorithmInput, startTime time.Time) (autoscale.AlgorithmOutput, error) {
	queueMap := make(map[string][]autoscale.AlgorithmJob)
	var outInstances []autoscale.Instance
	out := autoscale.AlgorithmOutput{}
	emptyTagJobs := make([]autoscale.AlgorithmJob, 0)
	for _, j := range input.JobQueue {
		if j.Tag == "" {
			emptyTagJobs = append(emptyTagJobs, j)
			continue
		}
		queueMap[j.Tag] = append(queueMap[j.Tag], j)
	}

	for _, job := range emptyTagJobs {
		lowestCostcloud := ""
		lowestCost := math.MaxFloat64
		shortestCloud := ""
		var shortest int64
		shortest = math.MaxInt64
		for key, cloud := range input.Clouds {
			if queue, ok := queueMap[key]; ok {
				cost := cloud.GetTotalCost(queue, startTime)
				duration, err := cloud.GetTotalDuration(queue, startTime)
				if err != nil {
					return out, err
				}
				if duration < shortest {
					shortest = duration
					shortestCloud = key
				}
				if cost < lowestCost {
					lowestCost = cost
					lowestCostcloud = key
				}
			} else {
				shortestCloud = key
				flav := "default"
				if job.InstanceFlavour != "" {
					flav = job.InstanceFlavour
				}
				cost := cloud.GetExpectedJobCost(job, flav, startTime)
				if cost < lowestCost {
					lowestCost = cost
					lowestCostcloud = key
				}
				if dur, ok := job.ExecutionTime[key]; ok {
					duration := dur
					if duration < shortest {
						shortest = duration
						shortestCloud = key
					}
				}
			}
		}
		if shortestCloud == lowestCostcloud {
			job.Tag = lowestCostcloud
			queueMap[lowestCostcloud] = append(queueMap[lowestCostcloud], job)
		} else {
			job.Tag = shortestCloud
			queueMap[shortestCloud] = append(queueMap[shortestCloud], job)
		}
	}

	for key, queue := range queueMap {
		curClust := input.Clouds[key]
		instances, err := curClust.GetInstances()
		if err != nil {
			return out, err
		}

		//Set running jobs as the first elements
		//These does not require to use the priority since they are already running, should not pause jobs
		//First sort on priority
		//If priority is equal, sort by deadline
		sort.Slice(queue, func(i, j int) bool {
			if queue[i].State == autoscale.RUNNING {
				return true
			}
			if queue[i].Priority < queue[j].Priority {
				return true
			} else if queue[i].Priority == queue[j].Priority {
				if queue[i].Deadline.Before(queue[j].Deadline) {
					//If the job has a closer deadline, deacrease the priority value (increase priority)
					queue[i].Priority--
					return true
				} else {
					return false
				}
			}
			return false
		})

		runningJobs := 0
		activeInstances := 0
		for _, i := range instances {
			if i.State == autoscale.ACTIVE {
				activeInstances++
			}
		}

		sort.Slice(instances, func(i, j int) bool {
			return instances[i].State == "ACTIVE"
		})

		for _, job := range queue {
			if job.State == autoscale.RUNNING {
				runningJobs++
				continue
			}
			instances, err = curClust.GetInstances()
			if err != nil {
				return autoscale.AlgorithmOutput{}, err
			}
			//Check if there are room to add more cluster
			if curClust.GetInstanceLimit() > len(instances) {
				types, err := curClust.GetInstanceTypes()
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				iType := types["default"]
				instance := autoscale.Instance{
					Id:    "",
					Type:  iType.Name,
					State: "",
				}

				for k := activeInstances; k < len(instances); k++ {
					if instances[k].State == "INACTIVE" {
						instance = instances[k]
						activeInstances++
						break
					}
				}

				_, err = curClust.AddInstance(&instance, startTime)
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				outInstances = append(outInstances, instance)
			} else {
				//If all instances have been used, check for inactive instances and use them instead
				for k := activeInstances-1; k < len(instances); k++ {
					if instances[k].State == "INACTIVE" {
						_, err = curClust.AddInstance(&instances[k], startTime)
						activeInstances++
						break
					}
				}
			}
		}
		//DELETE instances based on the queue size, if the instance size is larger than queue,
		//there should be at least one INACTIVE instance
		//if len(instances) > (len(queue)) {
		//	for _, i := range instances {
		//		if i.State == autoscale.INACTIVE {
		//			err = curClust.DeleteInstance(i.Id, startTime)
		//			if err != nil {
		//				return autoscale.AlgorithmOutput{}, err
		//			}
		//			outInstances = append(outInstances, i)
		//		}
		//	}
		//}

		for key, _ := range input.Clouds {
			if _, ok := queueMap[key]; !ok {
				instances, err := input.Clouds[key].GetInstances()
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				if len(instances) > (len(queue)) {
					for _, inst := range instances {
						if inst.State == autoscale.INACTIVE {
							err := input.Clouds[key].DeleteInstance(inst.Id, startTime)
							if err != nil {
								return autoscale.AlgorithmOutput{}, err
							}
						}
					}
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
