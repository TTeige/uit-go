package models

import (
	"time"
	"database/sql"
	"github.com/lib/pq"
)

type AutoscalingRun struct {
	AutoscalingRunStats
	Events []CloudEvent
}

type AutoscalingRunStats struct {
	Id       int
	Name     string
	Started  time.Time
	Finished pq.NullTime
}

func GetAllAutoscalingRunStats(db *sql.DB) ([]AutoscalingRunStats, error) {
	stats := make([]AutoscalingRunStats, 0)
	rows, err := db.Query("SELECT * FROM autoscaling_run")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var run AutoscalingRunStats
		err = rows.Scan(&run.Id, &run.Name, &run.Started, &run.Finished)
		if err != nil {
			return nil, err
		}
		stats = append(stats, run)
	}

	return stats, nil
}

func GetAutoscalingRunEvents(db *sql.DB, runId string) ([]CloudEvent, error) {
	rows, err := db.Query("SELECT * FROM cloud_events WHERE run_name = $1", runId)
	if err != nil {
		return nil, err
	}

	events := make([]CloudEvent, 0)

	for rows.Next() {
		var event CloudEvent
		err = rows.Scan(&event.Id, &event.RunId, &event.Created, &event.Instance.Id, &event.Type, &event.InstanceType.Name, &event.InstanceType.PriceIncrement, &event.Instance.State, &event.CloudName)
		if err != nil {
			return nil, err
		}
		event.Instance.Type = event.InstanceType.Name
		events = append(events, event)
	}

	return events, nil
}

func GetAutoscalingRun(db *sql.DB, runName string) (*AutoscalingRun, error) {
	run := new(AutoscalingRun)
	err := db.QueryRow("SELECT * FROM autoscaling_run WHERE name = $1", runName).Scan(
		&run.Id, &run.Name, &run.Started, &run.Finished)
	if err != nil {
		return nil, err
	}

	events, err := GetAutoscalingRunEvents(db, runName)
	if err != nil {
		return nil, err
	}
	run.Events = events

	return run, nil
}

func UpdateAutoscalingRun(db *sql.DB, run string, finTime time.Time) error {
	_, err := db.Exec("UPDATE autoscaling_run SET finished = $1 WHERE name = $2", finTime, run)
	if err != nil {
		return err
	}
	return nil
}

func CreateAutoscalingRun(db *sql.DB, runName string, startTime time.Time) (string, error) {
	_, err := db.Exec("INSERT INTO autoscaling_run (name, started) VALUES ($1, $2)",
		runName, startTime)
	if err != nil {
		return "", err
	}
	sim, err := GetAutoscalingRun(db, runName)
	if err != nil {
		return "", err
	}
	return sim.Name, nil
}
