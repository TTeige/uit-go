package models

import "database/sql"

type Parameters struct {
	InputContigsCutoff     int  `json:"inputContigsCutoff"`
	UseBlastUniref50       bool `json:"useBlastUniref50"`
	UseInterproScan5       bool `json:"useInterproScan5"`
	UsePriam               bool `json:"usePriam"`
	RemoveNonCompleteGenes bool `json:"removeNonCompleteGenes"`
	ExportMergedGenbank    bool `json:"exportMergedGenbank"`
	UseBlastMarRef         bool `json:"useBlastMarRef"`
	JobId                  string
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
