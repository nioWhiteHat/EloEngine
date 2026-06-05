package elocalculator

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type MLData struct {
	UID            string  `json:"UID"`
	GeoPremium     float64 `json:"geo_premium"`
	PredictedPrice float64 `json:"final_predicted_sqm_price"`
	Residual       float64 `json:"final_residual_percentage"`
}

// MergeMLDataToCSV takes the raw CSV and the Python JSON,
// and creates a new complete CSV with the appended ML columns.
func MergeMLDataToCSV(inputCSVPath, jsonPath, outputCSVPath string) error {

	// 1. Load JSON into a Map for lightning-fast O(1) lookups
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON: %v", err)
	}

	var mlRecords []MLData
	if err := json.Unmarshal(jsonFile, &mlRecords); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	mlMap := make(map[string]MLData, len(mlRecords))
	for _, record := range mlRecords {
		mlMap[record.UID] = record
	}

	// 2. Open Input CSV for streaming
	csvFile, err := os.Open(inputCSVPath)
	if err != nil {
		return fmt.Errorf("failed to open input CSV: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1 // Forgive messy rows

	// 3. Create Output CSV
	outFile, err := os.Create(outputCSVPath)
	if err != nil {
		return fmt.Errorf("failed to create output CSV: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// 4. Read Header & Find the exact UID index safely
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}

	uidIndex := -1
	for i, colName := range header {
		// Strip Python's invisible UTF-8 BOM just to be safe
		cleanName := strings.TrimSpace(strings.TrimPrefix(colName, "\xef\xbb\xbf"))
		if cleanName == "UID" {
			uidIndex = i
			break
		}
	}

	if uidIndex == -1 {
		return fmt.Errorf("could not find UID column in the CSV")
	}

	// Write the new expanded header
	newHeader := append(header, "GeoPremium", "PredictedPricePerSqm", "FinalResidual")
	if err := writer.Write(newHeader); err != nil {
		return fmt.Errorf("failed to write new header: %v", err)
	}

	// 5. Stream the rows, couple the data, and write out
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip broken rows
		}
		if len(row) <= uidIndex {
			continue // Skip rows missing a UID
		}

		uid := strings.TrimSpace(row[uidIndex])

		// Inner Join Logic:
		// If the UID exists in the ML JSON, attach the 3 columns and write the row.
		// If it DOES NOT exist in the JSON, we drop the row entirely.
		if mlData, exists := mlMap[uid]; exists {
			row = append(row,
				strconv.FormatFloat(mlData.GeoPremium, 'f', -1, 64),
				strconv.FormatFloat(mlData.PredictedPrice, 'f', -1, 64),
				strconv.FormatFloat(mlData.Residual, 'f', -1, 64),
			)

			writer.Write(row)
		}
	}

	return nil
}