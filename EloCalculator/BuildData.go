package elocalculator

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Property struct {
	Uid            string  `json:"uid"`
	ResidualValue  float64 `json:"residual_value"`
	GeoPremium     float64 `json:"geo_premium"`
	PredictedPrice float64 `json:"predicted_price"`
	Mean500m       float64 `json:"mean_500m"`
	Std500m        float64 `json:"std_500m"`
	Elo            float64 `json:"elo"`
}

var PropertyMap = make(map[string]*Property)

const (
	EloBase          = 10000.0
	WeightResidual   = 2000.0
	WeightGeoPremium = 2500.0
	WeightZScore     = 500.0
)

func RunEloPipeline(ctx context.Context, conn *pgx.Conn, inputCSV string, outputJSON string) error {
	if err := TranslateCsvToProperties(inputCSV); err != nil {
		return err
	}

	var uids []string
	for uid := range PropertyMap {
		uids = append(uids, uid)
	}

	if err := FillAllNeighborhoodStats(ctx, conn, uids); err != nil {
		return err
	}

	for uid := range PropertyMap {
		CalculateElo(uid)
	}

	file, err := os.Create(outputJSON)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(PropertyMap)
}

var AllowedRegions = map[string]bool{
	"Περιφερειακή Ενότητα Κεντρικού Τομέα Αθηνών": true,
	"Περιφερειακή Ενότητα Βορείου Τομέα Αθηνών":   true,
	"Περιφερειακή Ενότητα Νοτίου Τομέα Αθηνών":    true,

	
	
}

func TranslateCsvToProperties(inputCSV string) error {
	file, err := os.Open(inputCSV)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV is empty or only contains a header")
	}

	header := records[0]
	idxUID, idxGeoPremium, idxPred, idxResidual, idxLat, idxLon, idxAdmin := -1, -1, -1, -1, -1, -1, -1

	for i, h := range header {
		cleanHeader := strings.TrimSpace(strings.TrimPrefix(h, "\xef\xbb\xbf"))

		switch cleanHeader {
		case "UID":
			idxUID = i
		case "GeoPremium":
			idxGeoPremium = i
		case "PredictedPricePerSqm":
			idxPred = i
		case "FinalResidual":
			idxResidual = i
		case "GeoLat":
			idxLat = i
		case "GeoLon":
			idxLon = i
		case "AdminName":
			idxAdmin = i
		}
	}

	// Safety check ensures AdminName is actually in your CSV
	if idxUID == -1 || idxLat == -1 || idxLon == -1 || idxPred == -1 || idxAdmin == -1 {
		return fmt.Errorf("missing required columns in CSV (make sure AdminName exists)")
	}

	for _, row := range records[1:] {
		if len(row) <= idxUID || len(row) <= idxLat || len(row) <= idxPred || len(row) <= idxAdmin {
			continue
		}

		// 1. Check the Region First (Instant Rejection for outside Attica)
		adminName := strings.TrimSpace(row[idxAdmin])
		if !AllowedRegions[adminName] {
			continue // Skip this property entirely
		}

		// 2. Process the allowed properties
		uid := strings.TrimSpace(row[idxUID])
		if uid == "" {
			continue
		}

	
		geoPremium, _ := strconv.ParseFloat(strings.TrimSpace(row[idxGeoPremium]), 64)
		predPrice, _ := strconv.ParseFloat(strings.TrimSpace(row[idxPred]), 64)
		residual, _ := strconv.ParseFloat(strings.TrimSpace(row[idxResidual]), 64)

		PropertyMap[uid] = &Property{
			Uid:            uid,
			
			ResidualValue:  residual,
			GeoPremium:     geoPremium,
			PredictedPrice: predPrice,
		}
	}

	return nil
}

func FillAllNeighborhoodStats(ctx context.Context, conn *pgx.Conn, uids []string) error {
	totalUids := len(uids)
	if totalUids == 0 {
		return nil
	}

	batchSize := 2000
	startTime := time.Now()

	fmt.Printf("Starting Spatial ELO Math for %d properties...\n", totalUids)

	for i := 0; i < totalUids; i += batchSize {
		end := i + batchSize
		if end > totalUids {
			end = totalUids
		}
		
		batchUids := uids[i:end]

		query := `
			SELECT 
				t1.property_uid AS target_uid,
				json_agg(t2.property_uid) AS neighbors
			FROM property_geo_data t1
			JOIN property_geo_data t2 
				ON ST_DWithin(t1.geom, t2.geom, 0.0016) 
			WHERE t1.property_uid = ANY($1)
			GROUP BY t1.property_uid;
		`

		rows, err := conn.Query(ctx, query, batchUids)
		if err != nil {
			return fmt.Errorf("database error on batch %d-%d: %v", i, end, err)
		}

		for rows.Next() {
			var targetUID string
			var neighbors []string

			if err := rows.Scan(&targetUID, &neighbors); err != nil {
				rows.Close()
				return err
			}

			prop, exists := PropertyMap[targetUID]
			if !exists || prop == nil {
				continue
			}

			var prices []float64
			for _, neighborUID := range neighbors {
				if targetUID == neighborUID {
					continue
				}
				if neighbor, ok := PropertyMap[neighborUID]; ok && neighbor != nil {
					prices = append(prices, neighbor.PredictedPrice)
				}
			}

			if len(prices) == 0 {
				prop.Mean500m = 0
				prop.Std500m = 0
				continue
			}

			var sum float64
			for _, p := range prices {
				sum += p
			}
			prop.Mean500m = sum / float64(len(prices))

			var varianceSum float64
			for _, p := range prices {
				varianceSum += math.Pow(p-prop.Mean500m, 2)
			}
			prop.Std500m = math.Sqrt(varianceSum / float64(len(prices)))
		}
		rows.Close()

		percentDone := (float64(end) / float64(totalUids)) * 100
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%.2f%%] Processed %d / %d | Elapsed: %.1fs\n", percentDone, end, totalUids, elapsed)
	}

	fmt.Printf("✅ Finished Spatial ELO Math in %.1f seconds.\n", time.Since(startTime).Seconds())
	return nil
}

func CalculateElo(uid string) {
	prop := PropertyMap[uid]

	zScore := 0.0

	if prop.Std500m > 0 {
		zScore = (prop.PredictedPrice - prop.Mean500m) / prop.Std500m
		if zScore > 3.0 {
			zScore = 3.0
		} else if zScore < -3.0 {
			zScore = -3.0
		}
	}

	prop.Elo = EloBase - (prop.ResidualValue * WeightResidual) + (prop.GeoPremium * WeightGeoPremium) + (zScore * WeightZScore)
}