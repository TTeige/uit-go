package algorithm

import (
	"time"
	"github.com/tteige/uit-go/autoscale"
)

type NaiveAlgorithm struct {
	ScaleUpThreshold   int
	ScaleDownThreshold int
}

func (n NaiveAlgorithm) Run(input autoscale.AlgorithmInput, stepTime time.Time) (autoscale.AlgorithmOutput, error) {
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

		for _, job := range queue {
			if job.State == "RUNNING" {
				continue
			}
			if curClust.GetInstanceLimit() > len(instances) {
				if len(instances) == len(queue) {
					continue
				}
				if job.State != "QUEUED" {
					continue
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
				_, err = curClust.AddInstance(&instance)
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
					err = curClust.DeleteInstance(i.Id)
					if err != nil {
						return autoscale.AlgorithmOutput{}, err
					}
					outInstances = append(outInstances, i)
				}
			}
		}
	}

	out = autoscale.AlgorithmOutput{
		Instances: outInstances,
		JobQueue:  input.JobQueue,
	}

	return out, nil
}
