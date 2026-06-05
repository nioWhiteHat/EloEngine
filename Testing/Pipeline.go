package testing

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func TestingPipeline(pgp *pgx.Conn, area string, validUidsFile string) {
	ValidUids, err := LoadValidUids(validUidsFile)
	if err != nil {
		panic(err)
	}
	TestingData := GetFilteredUidsForArea(pgp, area, ValidUids)
	FillAttributesFromCSV(TestingData, "Modeled_Sold.csv")

	// 4. Output
	WriteResultsToCSV(area, TestingData)
	fmt.Printf("Pipeline complete for %s. Report generated.\n", area)

}