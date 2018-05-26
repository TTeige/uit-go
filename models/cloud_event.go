package models

import (
	"database/sql"
	"time"
	"github.com/tteige/uit-go/autoscale"
)

type CloudEvent struct {
	Id           int
	RunId        string
	Created      time.Time
	Instance     autoscale.Instance
	InstanceType autoscale.InstanceType
	Type         string
	CloudName    string
}

func GetSimEvent(db *sql.DB, id string) (*CloudEvent, error) {
	event := new(CloudEvent)
	err := db.QueryRow("SELECT * FROM cloud_events WHERE id = $1", id).Scan(&event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func WriteSimEvent(db *sql.DB, event CloudEvent) error {
	_, err := db.Exec("INSERT INTO cloud_events (run_name, created, instance_id, type, instance_type, price, instance_state, cloud_name)"+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)ON CONFLICT DO NOTHING",
		event.RunId, event.Created, event.Instance.Id, event.Type, event.Instance.Type, event.InstanceType.PriceIncrement, event.Instance.State, event.CloudName)
	if err != nil {
		return err
	}
	return nil
}
