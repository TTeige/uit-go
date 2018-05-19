package models

import (
	"time"
	"database/sql"
	"github.com/lib/pq"
)

type Simulation struct {
	SimulationStats
	Events []SimEvent
}

type SimulationStats struct {
	Id       int
	Name     string
	Started  time.Time
	Finished pq.NullTime
}

func GetAllSimulationStats(db *sql.DB) ([]SimulationStats, error) {
	stats := make([]SimulationStats, 0)
	rows, err := db.Query("SELECT * FROM simulation")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sim SimulationStats
		err = rows.Scan(&sim.Id, &sim.Name, &sim.Started, &sim.Finished)
		if err != nil {
			return nil, err
		}
		stats = append(stats, sim)
	}

	return stats, nil
}

func GetSimEvents(db *sql.DB, sim_id string) ([]SimEvent, error) {
	rows, err := db.Query("SELECT * FROM sim_events WHERE sim_name = $1", sim_id)
	if err != nil {
		return nil, err
	}

	events := make([]SimEvent, 0)

	for rows.Next() {
		var event SimEvent
		err = rows.Scan(&event.Id, &event.SimId, &event.Created, &event.Instance.Id, &event.Type, &event.InstanceType.Name, &event.InstanceType.PriceIncrement, &event.Instance.State)
		if err != nil {
			return nil, err
		}
		event.Instance.Type = event.InstanceType.Name
		events = append(events, event)
	}

	return events, nil
}

func GetSimulation(db *sql.DB, sim_name string) (*Simulation, error) {
	sim := new(Simulation)
	err := db.QueryRow("SELECT * FROM simulation WHERE name = $1", sim_name).Scan(
		&sim.Id, &sim.Name, &sim.Started, &sim.Finished)
	if err != nil {
		return nil, err
	}

	events, err := GetSimEvents(db, sim_name)
	if err != nil {
		return nil, err
	}
	sim.Events = events

	return sim, nil
}

func UpdateSim(db *sql.DB, sim string, finTime time.Time) error {
	_, err := db.Exec("UPDATE simulation SET finished = $1 WHERE name = $2", finTime, sim)
	if err != nil {
		return err
	}
	return nil
}

func CreateSimulation(db *sql.DB, sim_name string, startTime time.Time) (string, error) {
	_, err := db.Exec("INSERT INTO simulation (name, started) VALUES ($1, $2)",
		sim_name, startTime)
	if err != nil {
		return "", err
	}
	sim, err := GetSimulation(db, sim_name)
	if err != nil {
		return "", err
	}
	return sim.Name, nil
}
