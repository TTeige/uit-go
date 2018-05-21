package estimator

import (
	"github.com/tteige/uit-go/autoscale"
	"github.com/sajari/regression"
	"github.com/tteige/uit-go/models"
	"database/sql"
	"time"
	"net/http"
)

type LinearRegression struct {
	models map[string]*regression.Regression
	DB     *sql.DB
	Auth   autoscale.Oath2
}

func (lr *LinearRegression) Init() error {
	jobs, err := models.GetAllJobs(lr.DB)
	if err != nil {
		return err
	}
	var dp []RegJob
	for _, j := range jobs {
		if j.QueueDuration <= 0 {
			continue
		}
		param, err := models.GetParameter(lr.DB, j.JobId)
		if err != nil {
			return err
		}
		var regJ RegJob
		regJ.Params = param
		regJ.ExecTime = float64(j.Runtime)
		regJ.DataSize = float64(j.InputDataSize)
		regJ.Tag = j.Tag
		dp = append(dp, regJ)
	}
	lr.InitModel(dp)
	return nil
}

type RegJob struct {
	Params   models.Parameters
	DataSize float64
	ExecTime float64
	Tag      string
}

func generateParamBin(params autoscale.MetapipeParameter) float64 {
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

func (lr *LinearRegression) InitModel(dataPoints []RegJob) error {
	lr.models = make(map[string]*regression.Regression)
	dataPointMap := make(map[string]regression.DataPoints)

	for _, j := range dataPoints {
		if j.DataSize <= 0 {
			continue
		}

		tag := autoscale.GetTag(j.Tag)
		params := autoscale.MetapipeParameter{
			InputContigsCutoff:     j.Params.InputContigsCutoff,
			UseBlastUniref50:       j.Params.UseBlastUniref50,
			UseInterproScan5:       j.Params.UseInterproScan5,
			UsePriam:               j.Params.UsePriam,
			RemoveNonCompleteGenes: j.Params.RemoveNonCompleteGenes,
			ExportMergedGenbank:    j.Params.ExportMergedGenbank,
			UseBlastMarRef:         j.Params.UseBlastMarRef,
		}
		parVal := generateParamBin(params)
		dataPoint := regression.DataPoint(j.ExecTime, []float64{j.DataSize, parVal, float64(j.Params.InputContigsCutoff)})
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

func (lr *LinearRegression) estimateJob(params autoscale.MetapipeParameter, tag string, dataSize int64) (int64, error) {
	pred, err := lr.models[autoscale.GetTag(tag)].Predict([]float64{float64(dataSize), generateParamBin(params), float64(params.InputContigsCutoff)})
	if err != nil {
		return 0, err
	}
	return int64(pred), nil
}

func (lr *LinearRegression) ProcessQueue(jobs []autoscale.MetapipeJob) ([]autoscale.AlgorithmJob, error) {
	out := make([]autoscale.AlgorithmJob, 0)
	client := autoscale.RetryClient{
		Auth:        lr.Auth,
		MaxAttempts: 3,
		Client: http.Client{
			Timeout: time.Second * 5,
		},
	}
	for _, j := range jobs {
		if j.State == "CANCELLED" || j.State == "DELAYED" {
			continue
		}

		newTag := autoscale.GetTag(j.Tag)

		if newTag == "undefined" {
			continue
		}
		t, err := autoscale.ParseMetapipeTimestamp(j.TimeSubmitted)
		if err != nil {
			return nil, err
		}

		dataSize, err := client.GetMetapipeJobSize(j.Id)
		if err != nil {
			return nil, err
		}
		execTime, err := lr.estimateJob(j.Parameters, j.Tag, dataSize)
		if err != nil {
			return nil, err
		}

		outputJob := autoscale.AlgorithmJob{
			Id:            j.Id,
			Tag:           newTag,
			Parameters:    j.Parameters,
			State:         j.State,
			Priority:      j.Priority,
			ExecutionTime: execTime,
			Deadline:      time.Time{},
			Created:       t,
		}
		out = append(out, outputJob)
	}
	return out, nil
}
