package testing

import (
	t "eiv-ranking/cmd/XGboost/types"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

func FillAttributesFromCSV(targetData []TestingProperty, csvPath string) {
	file, err := os.Open(csvPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 1. Create a quick lookup map of the objects we actually care about
	targetMap := make(map[string]*TestingProperty)
	for i := range targetData {
		targetMap[targetData[i].UID] = &targetData[i]
	}

	// 2. Read the clean_properties.csv (Assumes first column is UID)
	// We only hydrate the Attributes field if the UID is in our targetMap
	reader := csv.NewReader(file)
	header, _ := reader.Read() // Skip header
    
    // Logic here depends on your CSV structure. 
    // We iterate the file and if record[0] (UID) is in targetMap, 
    // we map the columns to targetMap[uid].Attributes
	for {
		record, err := reader.Read()
		if err == io.EOF { break }
		uid := record[0]
		if prop, exists := targetMap[uid]; exists {
            // Manual mapping or use a CSV-to-Struct library here
			prop.Attributes = mapRecordToModel(record, header) 
		}
	}
}

func WriteResultsToCSV(area string, data []TestingProperty) {
	if len(data) == 0 {
		return
	}

	fileName := fmt.Sprintf("%s.csv", area)
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Helper to dynamically flatten the struct using reflection
	extractRow := func(p TestingProperty) ([]string, []string) {
		var headers []string
		var values []string

		// 1. Grab the ML metrics (ignoring the top-level UID)
		// --> ADD PredictedPricePerSqm here
		headers = append(headers, "GeoPremium", "PredictedPricePerSqm", "FinalResidual")
		values = append(values, 
			fmt.Sprintf("%f", p.GeoPremium), 
			fmt.Sprintf("%f", p.PredictedPrice), // <-- ADD THIS
			fmt.Sprintf("%f", p.FinalResidual),
		)

		// 2. Automatically un-nest every field inside Attributes
		v := reflect.ValueOf(p.Attributes)
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			headers = append(headers, t.Field(i).Name)
			values = append(values, fmt.Sprintf("%v", v.Field(i).Interface()))
		}

		return headers, values
	}

	// Extract headers and the first row's data
	headers, firstRow := extractRow(data[0])
	
	// Write Header
	writer.Write(headers)
	
	// Write First Row
	writer.Write(firstRow)

	// Write the rest of the rows dynamically
	for _, p := range data[1:] {
		_, row := extractRow(p)
		writer.Write(row)
	}
}

func mapRecordToModel(record []string, header []string) t.PropertyModel {
	idx := make(map[string]int)
	for i, name := range header {
		idx[name] = i
	}

	parseInt := func(name string) int {
		if i, ok := idx[name]; ok && i < len(record) {
			val, _ := strconv.Atoi(record[i])
			return val
		}
		return 0
	}

	parseFloat := func(name string) float64 {
		if i, ok := idx[name]; ok && i < len(record) {
			val, _ := strconv.ParseFloat(record[i], 64)
			return val
		}
		return 0.0
	}

	var m t.PropertyModel
	
	m.UID = record[idx["UID"]]
	m.Age = parseInt("Age")
	m.RenovAge = parseInt("RenovAge")
	m.Distance = parseInt("Distance")
	m.IsRenovated = parseInt("IsRenovated")
	m.AgeSq = parseInt("AgeSq")
	m.InteractionAgeDistance = parseInt("InteractionAgeDistance")
	m.LogInteraction = parseFloat("LogInteraction")
	m.SubType = record[idx["SubType"]]
	m.AdCreationAge = parseInt("AdCreationAge") 
	m.AskingPricePerSqm = parseFloat("AskingPricePerSqm")
	m.SquareMeters = parseInt("SquareMeters")
	m.LandplotSize = parseInt("LandplotSize")
	m.NoView = parseInt("NoView")
	m.ViewForest = parseInt("ViewForest")
	m.ViewMountain = parseInt("ViewMountain")
	m.ViewSea = parseInt("ViewSea")
	m.ViewUnlimited = parseInt("ViewUnlimited")
	m.NoFacade = parseInt("NoFacade")
	m.FacadeCorner = parseInt("FacadeCorner")
	m.FacadeFrontage = parseInt("FacadeFrontage")
	m.FacadeInterior = parseInt("FacadeInterior")
	m.FacadeThrough = parseInt("FacadeThrough")
	m.HeatType = record[idx["HeatType"]]
	m.HeatFrom = record[idx["HeatFrom"]]
	m.EnergyPerformance = record[idx["EnergyPerformance"]]
	m.WC = parseInt("WC")
	m.AC = parseInt("AC")
	m.Elevator = parseInt("Elevator")
	m.Parking = parseInt("Parking")
	m.StorageRoom = parseInt("StorageRoom")
	m.AluminumFrames = parseInt("AluminumFrames")
	m.Bedrooms = parseInt("Bedrooms")
	m.Pool = parseInt("Pool")
	m.Garden = parseInt("Garden")
	m.FloorLevel = parseInt("FloorLevel")
	m.FloorCount = parseInt("FloorCount")
	m.GeoLat = parseFloat("GeoLat")
	m.GeoLon = parseFloat("GeoLon")
	m.AskingPrice = parseInt("AskingPrice")

	return m
}