package estimator

import (
	"github.com/tteige/uit-go/autoscale"
	"github.com/sajari/regression"
	"time"
	"net/http"
	"database/sql"
	"github.com/tteige/uit-go/models"
	"github.com/tteige/uit-go/metapipe"
)

type LinearRegression struct {
	models map[string]*regression.Regression
	Auth   metapipe.Oath2
	DB     *sql.DB
}

type RegressionJob struct {
	Params   models.Parameters
	DataSize float64
	ExecTime float64
	Tag      string
}

func (lr *LinearRegression) Init() error {
	var dataPoints []RegressionJob
	jobs, err := models.GetAllJobs(lr.DB)
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if j.QueueDuration <= 0 {
			continue
		}
		param, err := models.GetParameters(lr.DB, j.JobId)
		if err != nil {
			return err
		}
		var regJ RegressionJob
		regJ.Params = param
		regJ.ExecTime = float64(j.Runtime)
		regJ.DataSize = float64(j.InputDataSize)
		regJ.Tag = j.Tag
		dataPoints = append(dataPoints, regJ)
	}
	lr.InitModel(dataPoints)
	return nil
}

func generateParamBin(params metapipe.Parameters) float64 {
	x := 1
	if params.UseBlastMarRef {
		x = x | 1<<1
	}
	if params.ExportMergedGenbank {
		x = x | 1<<2
	}
	if params.RemoveNonCompleteGenes {
		x = x | 1<<3
	}
	if params.UsePriam {
		x = x | 1<<4
	}
	if params.UseInterproScan5 {
		x = x | 1<<5
	}
	if params.UseBlastUniref50 {
		x = x | 1<<6
	}
	return float64(x)
}

func (lr *LinearRegression) InitModel(dataPoints []RegressionJob) error {
	lr.models = make(map[string]*regression.Regression)
	dataPointMap := make(map[string]regression.DataPoints)

	for _, j := range dataPoints {
		if j.DataSize <= 0 {
			continue
		}

		tag := metapipe.GetTag(j.Tag)
		parVal := generateParamBin(j.Params.MP)
		dataPoint := regression.DataPoint(j.ExecTime, []float64{j.DataSize, parVal, float64(j.Params.MP.InputContigsCutoff)})
		dataPointMap[tag] = append(dataPointMap[tag], dataPoint)
	}

	for key, val := range dataPointMap {
		r := new(regression.Regression)
		r.SetVar(0, "executionTime")
		r.SetVar(1, "datasize")
		r.SetVar(2, "parameters")
		r.SetVar(3, "contigs")
		for _, p := range val {
			r.Train(p)
		}
		r.Run()
		lr.models[key] = r
		//log.Printf("__________________________________________________")
		//log.Printf("Regression formula for %s:\n%v\n", key, r.Formula)
		//log.Printf("Regression:\n%s\n", r)
		//log.Printf("__________________________________________________")
	}
	return nil
}

func (lr *LinearRegression) estimateJob(params metapipe.Parameters, tag string, dataSize int64) (int64, error) {
	pred, err := lr.models[metapipe.GetTag(tag)].Predict([]float64{float64(dataSize), generateParamBin(params), float64(params.InputContigsCutoff)})
	if err != nil {
		return 0, err
	}
	return int64(pred), nil
}

func (lr *LinearRegression) ProcessQueue(jobs []autoscale.AlgorithmJob) ([]autoscale.AlgorithmJob, error) {
	out := make([]autoscale.AlgorithmJob, 0)
	client := metapipe.RetryClient{
		Auth:        lr.Auth,
		MaxAttempts: 3,
		Client: http.Client{
			Timeout: time.Second * 5,
		},
	}

	for _, j := range jobs {
		if j.State == "CANCELLED" {
			continue
		}

		newTag := metapipe.GetTag(j.Tag)

		if newTag == "undefined" {
			continue
		}


		dataSize, err := client.GetMetapipeJobSize(j.Id)
		if err != nil {
			continue
		}
		execMap := make(map[string]int64)
		var execTime int64
		if newTag == "" {
			execTime, err = lr.estimateJob(metapipe.ConvertToMetapipeParamaters(j.Parameters), metapipe.AWS, dataSize)
			if err != nil {
				return nil, err
			}
			execMap[metapipe.AWS] = execTime
			execTime, err = lr.estimateJob(metapipe.ConvertToMetapipeParamaters(j.Parameters), metapipe.CPouta, dataSize)
			if err != nil {
				return nil, err
			}
			execMap[metapipe.CPouta] = execTime
			execTime, err = lr.estimateJob(metapipe.ConvertToMetapipeParamaters(j.Parameters), metapipe.Stallo, dataSize)
			if err != nil {
				return nil, err
			}
			execMap[metapipe.Stallo] = execTime
		} else {
			execTime, err = lr.estimateJob(metapipe.ConvertToMetapipeParamaters(j.Parameters), j.Tag, dataSize)
			execMap[j.Tag] = execTime
		}

		outputJob := autoscale.AlgorithmJob{
			Id:            j.Id,
			Tag:           newTag,
			Parameters:    j.Parameters,
			State:         j.State,
			Priority:      j.Priority,
			ExecutionTime: execMap,
			Deadline:      time.Time{},
			Created:       j.Created,
			Started:       j.Started,
		}
		out = append(out, outputJob)
	}
	return out, nil
}
