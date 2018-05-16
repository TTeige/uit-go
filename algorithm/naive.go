package algorithm

import (
	"time"
	"github.com/tteige/uit-go/autoscale"
)

type NaiveAlgorithm struct {
	ScaleUpThreshold   int
	ScaleDownThreshold int
}

func (n NaiveAlgorithm) Step(input autoscale.AlgorithmInput, stepTime time.Time) (autoscale.AlgorithmOutput, error) {
	queueMap := make(map[string][]autoscale.AlgorithmJob)
	var outInstances []autoscale.Instance
	var out autoscale.AlgorithmOutput
	//Split the input, if no tag is defined, default value is ""
	for _, j := range input.JobQueue {
		queueMap[j.Tag] = append(queueMap[j.Tag], j)
	}

	for key, queue := range queueMap {
		for _, job := range queue {
			if job.State == "RUNNING" {
				continue
			}
			curClust := input.Clouds[key]
			instances, err := curClust.GetInstances()
			if err != nil {
				return out, err
			}
			if curClust.GetInstanceLimit() < len(instances) {
				types, err := curClust.GetInstanceTypes()
				if err != nil {
					return autoscale.AlgorithmOutput{}, err
				}
				var iType autoscale.InstanceType
				if key == autoscale.AWS {
					iType = types["c5_4xl"]
				} else if key == autoscale.Stallo {
					iType = types["default_stallo"]
				} else if key == autoscale.CPouta {
					iType = types["default_csc"]
				}
				instance := autoscale.Instance{
					Id:    "",
					Type:  iType.Name,
					State: "",
				}
				curClust.AddInstance(&instance)
				outInstances = append(outInstances, instance)
			}
		}
	}

	out = autoscale.AlgorithmOutput{
		Instances: outInstances,
		JobQueue:  input.JobQueue,
	}

	return out, nil
}
