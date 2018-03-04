package algorithm

import (
	"time"
	"github.com/tteige/uit-go/autoscale"
)

type NaiveAlgorithm struct {
	ScaleUpThreshold   int
	ScaleDownThreshold int
}

func (n NaiveAlgorithm) Step(input autoscale.AlgorithmInput, stepTime time.Time) (*autoscale.AlgorithmOutput, error) {
	return nil, nil
}
