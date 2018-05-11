package models

import (
	"database/sql"
	"github.com/tteige/uit-go/autoscale"
)

func GetInstanceTypes(db *sql.DB) ([]autoscale.InstanceType, error) {
	rows, err := db.Query("SELECT * FROM instance_types")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	iTypes := make([]autoscale.InstanceType, 0)

	for rows.Next() {
		iType := autoscale.InstanceType{}
		err := rows.Scan(&iType.Name, &iType.PriceIncrement)
		if err != nil {
			return nil, err
		}
		iTypes = append(iTypes, iType)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return iTypes, nil
}

func InsertInstanceType(db *sql.DB, instanceType autoscale.InstanceType, clusterName string) error {
	sqlStmt :=
		`INSERT INTO instance_types (name, priceincrement)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO NOTHING`

	_, err := db.Exec(sqlStmt, instanceType.Name, instanceType.PriceIncrement)
	if err != nil {
		return err
	}

	sqlStmt =
		`INSERT INTO cluster_to_instance_type (cluster_name, instance_name)
		VALUES ($1, $2)
		ON CONFLICT (instance_name)
		DO NOTHING`

	_, err = db.Exec(sqlStmt, clusterName, instanceType.Name)
	if err != nil {
		return err
	}

	return nil
}