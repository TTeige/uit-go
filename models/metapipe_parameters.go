package models

import (
	"database/sql"
	"github.com/tteige/uit-go/metapipe"
)

type Parameters struct {
	MP metapipe.Parameters
	JobId string
}

func InsertParameter(db *sql.DB, par Parameters) error {

	sqlStmt :=
		`INSERT INTO metapipe_parameters (inputcontigscutoff, useblastuniref50, useinterproscan5, usepriam, 
			removenoncompletegenes, exportmergedgenbank, useblastmarref, jobid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (jobid)
		DO NOTHING`

	_, err := db.Exec(sqlStmt, par.MP.InputContigsCutoff, par.MP.UseBlastUniref50, par.MP.UseBlastUniref50, par.MP.UsePriam,
		par.MP.RemoveNonCompleteGenes, par.MP.ExportMergedGenbank, par.MP.UseBlastMarRef, par.JobId)
	if err != nil {
		return err
	}
	return nil
}

func GetParameters(db *sql.DB, job string) (Parameters, error) {
	var par Parameters
	err := db.QueryRow("SELECT * FROM metapipe_parameters WHERE jobid = $1", job).Scan(&par.MP.InputContigsCutoff,
		&par.MP.UseBlastUniref50, &par.MP.UseInterproScan5, &par.MP.UsePriam, &par.MP.RemoveNonCompleteGenes, &par.MP.ExportMergedGenbank,
		&par.MP.UseBlastMarRef, &par.JobId)
	if err != nil {
		return Parameters{}, nil
	}
	return par, nil
}
