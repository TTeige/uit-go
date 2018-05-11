package models

import (
	"database/sql"
	"github.com/tteige/uit-go/autoscale"
)

func GetDefaultClusters(db *sql.DB) ([]autoscale.Cluster, error) {
	cPouta, err := GetCluster(db, "SimcPouta")
	if err != nil {
		return nil, err
	}
	aws, err := GetCluster(db, "SimAWS")
	if err != nil {
		return nil, err
	}
	stallo, err := GetCluster(db, "SimStallo")
	if err != nil {
		return nil, err
	}

	return []autoscale.Cluster{cPouta, aws, stallo}, nil
}

func GetCluster(db *sql.DB, name string) (autoscale.Cluster, error) {
	type partialCluster struct {
		Name  string
		Limit int
	}

	var pc partialCluster
	var cluster autoscale.Cluster

	err := db.QueryRow("SELECT * FROM simcluster WHERE name=$1", name).Scan(&pc)
	if err != nil {
		return cluster, err
	}

	tags, err := GetAcceptTags(db, name)
	if err != nil {
		return cluster, err
	}

	types, err := GetTypes(db, name)
	if err != nil {
		return cluster, err
	}

	instances, err := GetInstances(db, name)
	if err != nil {
		return cluster, err
	}

	cluster = autoscale.Cluster{
		Name:            pc.Name,
		Limit:           pc.Limit,
		AcceptTags:      tags,
		Types:           types,
		ActiveInstances: instances,
	}

	return cluster, nil
}

func GetAcceptTags(db *sql.DB, cluster_name string) ([]string, error) {
	rows, err := db.Query("SELECT tag_name FROM cluster_to_tag WHERE cluster_name = $1", cluster_name)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)

	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func GetInstances(db *sql.DB, cluster_name string) ([]autoscale.Instance, error) {
	rows, err := db.Query("SELECT id, type, state FROM cluster_instance WHERE cluster_name = $1", cluster_name)
	instances := make([]autoscale.Instance, 0)
	for rows.Next() {
		var instance autoscale.Instance
		err = rows.Scan(&instance.Id, &instance.Type, &instance.State)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, err
}

func GetTypes(db *sql.DB, cluster_name string) (map[string]autoscale.InstanceType, error) {
	stmt := `SELECT s2.name, s2.priceincrement 
	FROM cluster_to_instance_type 
	INNER JOIN instance_types s2 ON cluster_to_instance_type.instance_name 
	WHERE cluster_to_instance_type.cluster_name = $1`

	rows, err := db.Query(stmt, cluster_name)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	types := make(map[string]autoscale.InstanceType, 0)
	for rows.Next() {
		iType := autoscale.InstanceType{}
		err = rows.Scan(&iType.Name, &iType.PriceIncrement)
		if err != nil {
			return nil, err
		}
		types[iType.Name] = iType
	}
	return types, nil
}