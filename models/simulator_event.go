package models

import (
	"time"
	"database/sql"
)

type SimulatorEvent struct {
	RunName            string
	QueueDuration      int64
	AlgorithmTimestamp time.Time
	Tag                string
}

func InsertSimulatorEvent(db *sql.DB, event SimulatorEvent) error {
	_, err := db.Exec("INSERT INTO simulator_events (run_name, queue_duration, alg_timestamp, tag) VALUES ($1, $2, $3, $4)",
		event.RunName, event.QueueDuration, event.AlgorithmTimestamp, event.Tag)
	if err != nil {
		return err
	}
	return nil
}

func GetSimulatorEvents(db *sql.DB, runName string) ([]SimulatorEvent, error) {
	rows, err := db.Query("SELECT * FROM simulator_events WHERE run_name = $1", runName)
	if err != nil {
		return nil, err
	}
	var events []SimulatorEvent
	for rows.Next() {
		var event SimulatorEvent
		var id int
		err = rows.Scan(&id, &event.RunName, &event.QueueDuration, &event.AlgorithmTimestamp, &event.Tag)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
