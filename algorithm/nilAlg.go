package algorithm

import (
	"github.com/tteige/uit-go/autoscale"
	"time"
)

type NilAlg struct {
}

func (NilAlg) Run(input autoscale.AlgorithmInput, stepTime time.Time) (autoscale.AlgorithmOutput, error) {
	o := autoscale.AlgorithmOutput{
		Instances: nil,
		JobQueue:  input.JobQueue,
	}
	return o, nil
}



