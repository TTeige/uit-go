package algorithm

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"math"
)

type BadAlgorithm struct {
}

func (BadAlgorithm) Run(input autoscale.AlgorithmInput, startTime time.Time) (autoscale.AlgorithmOutput, error) {
	var out autoscale.AlgorithmOutput
	queueMap := make(map[string][]autoscale.AlgorithmJob)
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
				var duration int64
				duration = 0
				flav := "default"
				if job.InstanceFlavour.Name != "" {
					flav = job.InstanceFlavour.Name
				}
				cost := cloud.GetExpectedJobCost(job, flav, startTime)
				if cost < lowestCost {
					lowestCost = cost
					lowestCostcloud = key
				}
				if duration < shortest {
					shortest = duration
					shortestCloud = key
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
	var outQueue []autoscale.AlgorithmJob
	for _, m := range queueMap {
		for _, j := range m {
			outQueue = append(outQueue, j)
		}
	}

	for _, cloud := range input.Clouds {
		instances, err := cloud.GetInstances()
		if err != nil {
			return out, err
		}
		for len(instances) > 1 {
			index := 0
			for _, instance := range instances {
				if instance.State == autoscale.INACTIVE {
					break
				}
				index++
			}
			if index < len(instances) && instances[index].State == autoscale.INACTIVE {
				cloud.DeleteInstance(instances[index].Id, startTime)
				instances, err = cloud.GetInstances()
				if err != nil {
					return out, err
				}
			} else if index == len(instances) {
				break
			}
		}
		instances, err = cloud.GetInstances()
		if len(instances) == 1 && instances[0].State == autoscale.INACTIVE {
			cloud.AddInstance(&instances[0], startTime)
		}
		if len(instances) == 0 {

			types, err := cloud.GetInstanceTypes()
			if err != nil {
				return out, err
			}
			iType := types["default"]
			i := autoscale.Instance{
				Id:    "",
				Type:  iType.Name,
				State: "",
			}
			cloud.AddInstance(&i, startTime)
			out.Instances = append(out.Instances, i)
		}
	}
	out.JobQueue = outQueue
	return out, nil
}
