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
			cloud.DeleteInstance(instances[0].Id, startTime)
		}
		var iType autoscale.InstanceType
		types, err := cloud.GetInstanceTypes()
		if err != nil {
			return out, err
		}
		if key == metapipe.AWS {
			iType = types["c5-4xl"]
		} else if key == metapipe.Stallo {
			iType = types["default"]
		} else if key == metapipe.CPouta {
			iType = types["default"]
		}
		if len(instances) == 0 {
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
