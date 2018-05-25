package metapipe

import (
	"github.com/tteige/uit-go/autoscale"
	"strconv"
	"time"
	"strings"
)

const (
	AWS    = "aws"
	Stallo = "metapipe"
	CPouta = "csc"
)

type InputFas struct {
	Url string `json:"url"`
}

type dataUrl struct {
	InputFas InputFas `json:"input.fas"`
}

type Attempt struct {
	ExecutorId          string  `json:"executorId"`
	State               string  `json:"state"`
	AttemptId           string  `json:"attemptId"`
	Tag                 string  `json:"tag"`
	TimeCreated         string  `json:"timeCreated"`
	TimeStarted         string  `json:"timeStarted"`
	TimeEnded           string  `json:"timeEnded"`
	LastHeartbeat       string  `json:"lastHeartbeat"`
	RuntimeMillis       int     `json:"runtimeMillis"`
	QueueDurationMillis int     `json:"queueDurationMillis"`
	Outputs             dataUrl `json:"outputs"`
	Priority            int     `json:"priority"`
}
type Parameters struct {
	InputContigsCutoff     int  `json:"inputContigsCutoff"`
	UseBlastUniref50       bool `json:"useBlastUniref50"`
	UseInterproScan5       bool `json:"useInterproScan5"`
	UsePriam               bool `json:"usePriam"`
	RemoveNonCompleteGenes bool `json:"removeNonCompleteGenes"`
	ExportMergedGenbank    bool `json:"exportMergedGenbank"`
	UseBlastMarRef         bool `json:"useBlastMarRef"`
}

type Job struct {
	Id                       string     `json:"jobId"`
	TimeSubmitted            string     `json:"timeSubmitted"`
	State                    string     `json:"state"`
	UserId                   string     `json:"userId"`
	Tag                      string     `json:"tag"`
	Priority                 int        `json:"priority"`
	Hold                     bool       `json:"hold"`
	Parameters               Parameters `json:"parameters"`
	Inputs                   dataUrl    `json:"inputs"`
	Outputs                  dataUrl    `json:"outputs"`
	TotalRuntimeMillis       int64      `json:"totalRuntimeMillis"`
	TotalQueueDurationMillis int64      `json:"totalQueueDurationMillis"`
	Attempts                 []Attempt  `json:"attempts"`
}

func ConvertFromMetapipeParameters(parameter Parameters) (autoscale.JobParameters) {
	out := make(autoscale.JobParameters)
	out["InputContigsCutoff"] = strconv.FormatInt(int64(parameter.InputContigsCutoff), 10)
	out["UseBlastUniref50"] = strconv.FormatBool(parameter.UseBlastUniref50)
	out["UseInterproScan5"] = strconv.FormatBool(parameter.UseInterproScan5)
	out["UsePriam"] = strconv.FormatBool(parameter.UsePriam)
	out["RemoveNonCompleteGenes"] = strconv.FormatBool(parameter.RemoveNonCompleteGenes)
	out["ExportMergedGenbank"] = strconv.FormatBool(parameter.ExportMergedGenbank)
	out["UseBlastMarRef"] = strconv.FormatBool(parameter.UseBlastMarRef)
	return out
}

func ConvertToMetapipeParamaters(parameters autoscale.JobParameters) Parameters {

	cutoff, err := strconv.ParseInt(parameters["InputContigsCutoff"], 10, 64)
	if err != nil {
		return Parameters{}
	}
	ubu50, err := strconv.ParseBool(parameters["UseBlastUniref50"])
	if err != nil {
		return Parameters{}
	}
	uis5, err := strconv.ParseBool(parameters["UseBlastUniref50"])
	if err != nil {
		return Parameters{}
	}
	up, err := strconv.ParseBool(parameters["UsePriam"])
	if err != nil {
		return Parameters{}
	}
	rncg, err := strconv.ParseBool(parameters["RemoveNonCompleteGenes"])
	if err != nil {
		return Parameters{}
	}
	emg, err := strconv.ParseBool(parameters["ExportMergedGenbank"])
	if err != nil {
		return Parameters{}
	}
	ubmr, err := strconv.ParseBool(parameters["UseBlastMarRef"])
	if err != nil {
		return Parameters{}
	}

	out := Parameters{
		InputContigsCutoff:     int(cutoff),
		UseBlastUniref50:       ubu50,
		UseInterproScan5:       uis5,
		UsePriam:               up,
		RemoveNonCompleteGenes: rncg,
		ExportMergedGenbank:    emg,
		UseBlastMarRef:         ubmr,
	}
	return out
}

func ParseMetapipeTimestamp(stamp string) (time.Time, error) {
	t, err := strconv.ParseInt(stamp[:10], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	n, err := strconv.ParseInt(stamp[11:], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(t, n), nil
}

func ConvertMetapipeQueueToAlgInputJobs(jobs []Job) ([]autoscale.AlgorithmJob, error) {
	var out []autoscale.AlgorithmJob
	for _, j := range jobs {
		t, err := ParseMetapipeTimestamp(j.TimeSubmitted)
		if err != nil {
			return out, err
		}
		start, err := ParseMetapipeTimestamp(j.Attempts[0].TimeStarted)
		if err != nil {
			return out, err
		}
		algJob := autoscale.AlgorithmJob{
			Id:            j.Id,
			Tag:           j.Tag,
			Parameters:    ConvertFromMetapipeParameters(j.Parameters),
			State:         j.State,
			Priority:      j.Priority,
			ExecutionTime: map[string]int64{j.Tag:0},
			Deadline:      time.Time{},
			Created:       t,
			Started:       start,
		}
		out = append(out, algJob)
	}
	return out, nil
}

func GetTag(tag string) (string) {
	if strings.Contains(tag, AWS) {
		return AWS
	}
	if strings.Contains(tag, CPouta) {
		return CPouta
	}
	if strings.Contains(tag, Stallo) && !strings.Contains(tag, AWS) && ! strings.Contains(tag, CPouta) {
		return Stallo
	}
	if tag == "" {
		return ""
	}
	return "undefined"
}

func GetMetapipeJobs(defaultTime time.Time) []autoscale.AlgorithmJob {
	return []autoscale.AlgorithmJob{
		{
			Id:  "1",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "RUNNING",
			Priority:      2000,
			Deadline:      defaultTime.Add(time.Hour),
			Created:       defaultTime.Add(-time.Hour),
			Started:       defaultTime.Add(-time.Hour + time.Duration(time.Minute*5)),
			ExecutionTime: map[string]int64{"aws": 91847471},
		},
		{
			Id:  "2",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 130184712},
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(-time.Hour),
			Started:       time.Time{},
		},
		{
			Id:  "3",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 14712732},
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(time.Hour),
			Started:       time.Time{},
		},
		{
			Id:  "a",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 24712734},
			Deadline:      defaultTime.Add(time.Hour * 2),
			Created:       defaultTime.Add(time.Hour),
			Started:       time.Time{},
		},
		{
			Id:  "b",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 2347127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			Started:       time.Time{},
		},
		{
			Id:  "c",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 6647127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			Started:       time.Time{},
		},
		{
			Id:  "d",
			Tag: "aws",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"aws": 6647127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime.Add(time.Minute * 30),
			Started:       time.Time{},
		},
		{
			Id:  "e",
			Tag: "csc",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"csc": 6647127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			Started:       time.Time{},
		},
		{
			Id:  "f",
			Tag: "csc",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"csc": 6647127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			Started:       time.Time{},
		},
		{
			Id:  "g",
			Tag: "csc",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"csc": 3547127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			Started:       time.Time{},
		},
		{
			Id:  "h",
			Tag: "metapipe",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      2000,
			ExecutionTime: map[string]int64{"metapipe": 3547127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			Started:       time.Time{},
		},
		{
			Id:  "i",
			Tag: "metapipe",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			ExecutionTime: map[string]int64{"metapipe": 21347127},
			Deadline:      defaultTime.Add(time.Hour * 1),
			Created:       defaultTime,
			Started:       time.Time{},
		},
		{
			Id:  "j",
			Tag: "metapipe",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			ExecutionTime: map[string]int64{"metapipe": 45347127},
			Deadline:      defaultTime.Add(time.Hour * 3),
			Created:       defaultTime.Add(time.Hour * 2),
			Started:       time.Time{},
		},
		{
			Id:  "k",
			Tag: "metapipe",
			Parameters: ConvertFromMetapipeParameters(Parameters{
				InputContigsCutoff:     500,
				UseBlastUniref50:       true,
				UseInterproScan5:       false,
				UsePriam:               false,
				RemoveNonCompleteGenes: true,
				ExportMergedGenbank:    false,
				UseBlastMarRef:         false,
			}),
			State:         "QUEUED",
			Priority:      1,
			ExecutionTime: map[string]int64{"metapipe": 34712713},
			Deadline:      defaultTime.Add(time.Hour * 2),
			Created:       defaultTime.Add(time.Hour * 1),
			Started:       time.Time{},
		},
	}
}

