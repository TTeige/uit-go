package estimator

import (
	"github.com/tteige/uit-go/autoscale"
	"github.com/sajari/regression"
	"github.com/tteige/uit-go/models"
	"strings"
	"database/sql"
)

type LinearRegression struct {
	models map[string]*regression.Regression
	DB     *sql.DB
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

func generateParamBin(regJob RegJob) float64 {
	x := 1
	if regJob.Params.UseBlastMarRef {
		x = x | 1<<1
	}
	if regJob.Params.ExportMergedGenbank {
		x = x | 1<<2
	}
	if regJob.Params.RemoveNonCompleteGenes {
		x = x | 1<<3
	}
	if regJob.Params.UsePriam {
		x = x | 1<<4
	}
	if regJob.Params.UseInterproScan5 {
		x = x | 1<<5
	}
	if regJob.Params.UseBlastUniref50 {
		x = x | 1<<6
	}
	return float64(x)
}

func getTag(tag string) (string) {
	if strings.Contains(tag, autoscale.AWS) {
		return autoscale.AWS
	}
	if strings.Contains(tag, autoscale.CPouta) {
		return autoscale.CPouta
	}
	if strings.Contains(tag, autoscale.Stallo) && !strings.Contains(tag, autoscale.AWS) && ! strings.Contains(tag, autoscale.CPouta) {
		return autoscale.Stallo
	}
	return "undefined"
}

func (lr *LinearRegression) InitModel(dataPoints []RegJob) error {
	lr.models = make(map[string]*regression.Regression)
	dataPointMap := make(map[string]regression.DataPoints)

	for _, j := range dataPoints {
		if j.DataSize <= 0 {
			continue
		}

		tag := getTag(j.Tag)
		parVal := generateParamBin(j)
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

func (lr *LinearRegression) ProcessQueue(jobs []autoscale.MetapipeJob) ([]autoscale.AlgorithmJob, error) {

	return nil, nil
}
