package algorithm

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
	"github.com/tteige/uit-go/metapipe"
)

type BadAlgorithm struct {
}

func (BadAlgorithm) Run(input autoscale.AlgorithmInput, startTime time.Time) (autoscale.AlgorithmOutput, error) {
	var out autoscale.AlgorithmOutput
	for key, cloud := range input.Clouds {
		instances, err := cloud.GetInstances()
		if err != nil {
			return out, err
		}
		for len(instances) > 1 {
			index := 0
			for _, instance := range instances {
				if instance.State == "INACTIVE" {
					break
				}
				index++
			}
			if index < len(instances) && instances[index].State == "INACTIVE" {
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
		if len(instances) == 1 && instances[0].State == "INACTIVE" {
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
	out.JobQueue = input.JobQueue
	return out, nil
}
