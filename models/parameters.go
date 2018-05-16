package models

import (
	"database/sql"
	"github.com/tteige/uit-go/autoscale"
)

type Parameters struct {
	autoscale.MetapipeParameter
	JobId string
}

func InsertParameter(db *sql.DB, par Parameters) error {

	sqlStmt :=
		`INSERT INTO parameters (inputcontigscutoff, useblastuniref50, useinterproscan5, usepriam, 
			removenoncompletegenes, exportmergedgenbank, useblastmarref, jobid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (jobid)
		DO NOTHING`

	_, err := db.Exec(sqlStmt, par.InputContigsCutoff, par.UseBlastUniref50, par.UseBlastUniref50, par.UsePriam,
		par.RemoveNonCompleteGenes, par.ExportMergedGenbank, par.UseBlastMarRef, par.JobId)
	if err != nil {
		return err
	}
	return nil
}

func GetParameter(db *sql.DB, job string) (Parameters, error) {
	var par Parameters
	err := db.QueryRow("SELECT * FROM parameters WHERE jobid = $1", job).Scan(&par.InputContigsCutoff,
		&par.UseBlastUniref50, &par.UseInterproScan5, &par.UsePriam, &par.RemoveNonCompleteGenes, &par.ExportMergedGenbank,
		&par.UseBlastMarRef, &par.JobId)
	if err != nil {
		return Parameters{}, nil
	}
	return par, nil
}
