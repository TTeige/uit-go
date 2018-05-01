package models

import (
	"database/sql"
	"time"
)

type SimEvent struct {
	Id         int
	SimId      string
	Created    time.Time
	InstanceId string
	Type       string
}

func GetSimEvent(db *sql.DB, id string) (*SimEvent, error) {
	event := new(SimEvent)
	err := db.QueryRow("SELECT * FROM sim_events WHERE id = $1", id).Scan(&event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func WriteSimEvent(db *sql.DB, event SimEvent) error {
	_, err := db.Exec("INSERT INTO sim_events (sim_name, created, instance_id, type) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		event.SimId, event.Created, event.InstanceId, event.Type)
	if err != nil {
		return err
	}
	return nil
}
