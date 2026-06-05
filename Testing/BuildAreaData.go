package testing

import (
	"context"
	t "eiv-ranking/cmd/XGboost/types"
	"encoding/json"
	"os"
	"slices"

	"github.com/jackc/pgx/v5"
)
type TestingProperty struct {
	UID            string
	Attributes     t.PropertyModel
	GeoPremium     float64
	PredictedPrice float64 
	FinalResidual  float64
}

func LoadValidUids(filePath string) (map[string]*TestingProperty, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Add the PredictedPrice field to the JSON parser
	var properties []struct {
		UID            string  `json:"UID"`
		GeoPremium     float64 `json:"geo_premium"`
		PredictedPrice float64 `json:"final_predicted_sqm_price"` // <-- ADD THIS
		FinalResidual  float64 `json:"final_residual_percentage"`
	}

	if err := json.Unmarshal(data, &properties); err != nil {
		return nil, err
	}

	uidMap := make(map[string]*TestingProperty)
	for _, p := range properties {
		uidMap[p.UID] = &TestingProperty{
			UID:            p.UID,
			GeoPremium:     p.GeoPremium,
			PredictedPrice: p.PredictedPrice, // <-- ADD THIS
			FinalResidual:  p.FinalResidual,
		}
	}

	return uidMap, nil
}

func GetFilteredUidsForArea(con *pgx.Conn, area string, validUids map[string]*TestingProperty) []TestingProperty {
	query := `
		SELECT p.uid
		FROM properties p 
		JOIN property_geo_data loc ON p.uid = loc.property_uid 
		JOIN admin_boundaries ab ON ST_Within(loc.geom, ab.geom)
		WHERE p.item_type = 'residence' AND ab.name = $1
	`
	rows, err := con.Query(context.Background(), query, area)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []TestingProperty
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			panic(err)
		}

		if data, exists := validUids[uid]; exists {
			result = append(result, *data)
		}
	}

	// Sort Descending by Residual
	slices.SortFunc(result, func(a, b TestingProperty) int {
		if a.FinalResidual > b.FinalResidual { return -1 }
		if a.FinalResidual < b.FinalResidual { return 1 }
		return 0
	})

	return result
}

