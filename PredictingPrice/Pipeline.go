package predictingprice

import (
	"bufio"

	t "eiv-ranking/cmd/XGboost/types"

	"os"

	"strconv"
	"strings"

	"github.com/yalue/onnxruntime_go"
)

// 1. A global map to hold your encodings

// 2. Load your text file ONCE when the app starts
func LoadCategoryMappings(filepath string) (map[string]map[string]float32, error) {
	mappings := make(map[string]map[string]float32)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) == 3 {
			col := parts[0]
			val := parts[1]
			idx, _ := strconv.ParseFloat(parts[2], 32)

			if mappings[col] == nil {
				mappings[col] = make(map[string]float32)
			}
			mappings[col][val] = float32(idx)
		}
	}
	return mappings, nil
}


func PrepareFeatures(p t.PropertyModel, mappings map[string]map[string]float32) []float32 {
	features := make([]float32, 0)

	features = append(features, float32(p.Age))
	features = append(features, float32(p.RenovAge))
	features = append(features, float32(p.IsRenovated))
	features = append(features, mappings["SubType"][p.SubType])
	features = append(features, float32(p.SquareMeters))
	features = append(features, float32(p.LandplotSize))
	features = append(features, float32(p.NoView))
	features = append(features, float32(p.ViewForest))
	features = append(features, float32(p.ViewMountain))
	features = append(features, float32(p.ViewSea))
	features = append(features, float32(p.ViewUnlimited))
	features = append(features, float32(p.NoFacade))
	features = append(features, float32(p.FacadeCorner))
	features = append(features, float32(p.FacadeFrontage))
	features = append(features, float32(p.FacadeInterior))
	features = append(features, float32(p.FacadeThrough))
	features = append(features, mappings["HeatType"][p.HeatType])
	features = append(features, mappings["HeatFrom"][p.HeatFrom])
	features = append(features, mappings["EnergyPerformance"][p.EnergyPerformance])
	features = append(features, float32(p.WC))
	features = append(features, float32(p.AC))
	features = append(features, float32(p.Elevator))
	features = append(features, float32(p.Parking))
	features = append(features, float32(p.StorageRoom))
	features = append(features, float32(p.AluminumFrames))
	features = append(features, float32(p.Bedrooms))
	features = append(features, float32(p.Pool))
	features = append(features, float32(p.Garden))
	features = append(features, float32(p.FloorLevel))
	features = append(features, float32(p.FloorCount))
	features = append(features, float32(p.ZoneValue))
	features = append(features, float32(p.MarketabillityFactor))
	features = append(features, mappings["AdminName"][p.AdminName])

	return features
}



func GetProperty() t.PropertyModel {
	return t.PropertyModel{
		UID:                      "property_01KP6BYMQS9Z0X49DPBG92KB69",
		Age:                      126,
		RenovAge:                 6,
		Distance:                 120,
		IsRenovated:              1,
		AgeSq:                    15876,
		InteractionAgeDistance:   15120,
		LogInteraction:           581.302450375031,
		SubType:                  "apartment",
		AdCreationAge:            6,
		AskingPricePerSqm:        2245.762711864407,
		SquareMeters:             118,
		LandplotSize:             118,
		NoView:                   0,
		ViewForest:               0,
		ViewMountain:             0,
		ViewSea:                  0,
		ViewUnlimited:            0,
		NoFacade:                 0,
		FacadeCorner:             1,
		FacadeFrontage:           1,
		FacadeInterior:           0,
		FacadeThrough:            1,
		HeatType:                 "autonomous_individual",
		HeatFrom:                 "",
		EnergyPerformance:        "Ζ",
		WC:                       2,
		AC:                       1,
		Elevator:                 0,
		Parking:                  0,
		StorageRoom:              0,
		AluminumFrames:           1,
		Bedrooms:                 3,
		Pool:                     0,
		Garden:                   0,
		FloorLevel:               1,
		FloorCount:               1,
		GeoLat:                   37.9854314,
		GeoLon:                   23.7199172,
		ZoneValue:                1850,
		MarketabillityFactor:     1.0,
		AskingPrice:              0,
		AdminName:                "Περιφερειακή Ενότητα Κεντρικού Τομέα Αθηνών",
	}

	
}
func PredictSinglePoint(modelPath string, features []float32) (float32, error) {
	onnxruntime_go.SetSharedLibraryPath("onnxengine.dll")

	err := onnxruntime_go.InitializeEnvironment()
	if err != nil {
		return 0, err
	}
	defer onnxruntime_go.DestroyEnvironment()

	inputShape := onnxruntime_go.NewShape(1, int64(len(features)))
	inputTensor, err := onnxruntime_go.NewTensor(inputShape, features)
	if err != nil {
		return 0, err
	}
	defer inputTensor.Destroy()

	outputShape := onnxruntime_go.NewShape(1, 1)
	outputData := make([]float32, 1)
	outputTensor, err := onnxruntime_go.NewTensor(outputShape, outputData)
	if err != nil {
		return 0, err
	}
	defer outputTensor.Destroy()

	session, err := onnxruntime_go.NewAdvancedSession(
		modelPath,
		[]string{"input"},
		[]string{"variable"},
		[]onnxruntime_go.Value{inputTensor},
		[]onnxruntime_go.Value{outputTensor},
		nil,
	)
	if err != nil {
		return 0, err
	}
	defer session.Destroy()

	err = session.Run()
	if err != nil {
		return 0, err
	}

	return outputData[0], nil
}

