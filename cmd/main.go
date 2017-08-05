package main

import (
	"github.com/tteige/uit-go/autoscalingV2"
	"log"
)

func main() {

	//history :=
	//	[]autoscalingV2.Job{
	//		{
	//			Id:          "historicalJob1",
	//			Duration:    1000,
	//			DataSetSize: 1.4,
	//			Tags:        []string{"metapipe"},
	//		},
	//	}
	//
	//queue :=
	//	[]autoscalingV2.Job{
	//		{
	//			Id:          "currentJob1",
	//			Duration:    10,
	//			DataSetSize: 1.4,
	//			Tags:        []string{"metapipe"},
	//		},
	//	}

	clusterConfig :=
		autoscalingV2.ClusterConfig{
			Nodes: []autoscalingV2.NodeConfig{
				{
					InstanceType: "m4.xlarge",
					VCpu:         4,
					Memory:       16,
					ClockSpeed:   2.4,
					Storage: autoscalingV2.Storage{
						Type:  "EBS-Only",
						Count: 0,
						Size:  0,
					},
				},
				{
					InstanceType: "m4.xlarge",
					VCpu:         4,
					Memory:       16,
					ClockSpeed:   2.4,
					Storage: autoscalingV2.Storage{
						Type:  "EBS-Only",
						Count: 0,
						Size:  0,
					},
				},
			},
		}

	nodes, err := autoscalingV2.RunSimulation("autoscalingV2/test.json", "", clusterConfig, 0)

	if err != nil {
		log.Fatalf("%s\n", err)
	}

	log.Printf("Possible best number of nodes = %d\n", nodes)

	return
}
