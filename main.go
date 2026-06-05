package main

import (
	"bytes"
	"context"
	db "eiv-ranking/Database"
	elo "eiv-ranking/cmd/XGboost/EloCalculator"
	t "eiv-ranking/cmd/XGboost/types"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/jackc/pgx/v5"
	//testing "eiv-ranking/cmd/XGboost/Testing"
)
func main() {
	con,_ := db.InitDB()
	BuildFiles(con)
	RunModelAttr("Modeled_Sold.csv", "Models/JsonModelData/ModelAttr_Sold.json")
	RunGeoPremiumCalc("Models/JsonModelData/ModelAttr_Sold.json", "Models/JsonModelData/GeoPremium_Sold.json")
	err := RunModelGeo("Modeled_Sold.csv", "Models/JsonModelData/GeoPremium_Sold.json", "Models/JsonModelData/ModelGeo_Sold.json")
	if err != nil {
		fmt.Printf("Error running ModelGeo: %v\n", err)
	}
	elo.MergeMLDataToCSV("Modeled_Sold.csv", "Models/JsonModelData/ModelGeo_Sold.json", "Modeled_Sold_ML.csv")
	err = elo.RunEloPipeline(context.Background(), con, "Modeled_Sold_ML.csv", "Models/JsonModelData/Elo2_Sold.json")
	if err != nil {
		fmt.Printf("Error running Elo pipeline: %v\n", err)
	}

	


}


func BuildFiles(con *pgx.Conn) {
	
	
	query := t.BuildQuery()
	Builder := t.Builder{
		Db:    con,
		Query: query,
	}
	err := Builder.Initialize("sell")
	if err != nil {
		panic(err)
	}
	ExportToCSV("Modeled_Sold.csv", Builder.PropertiesModeled)
	err = Builder.Initialize("rent")
	if err != nil {
		panic(err)
	}
	ExportToCSV("Modeled_Rent.csv", Builder.PropertiesModeled)


}


func ExportToCSV(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return fmt.Errorf("data must be a non-empty slice")
	}

	elemType := val.Index(0).Type()
	numFields := elemType.NumField()

	var headers []string
	for i := 0; i < numFields; i++ {
		headers = append(headers, elemType.Field(i).Name)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		var row []string
		for j := 0; j < numFields; j++ {
			fieldVal := elem.Field(j).Interface()
			row = append(row, fmt.Sprintf("%v", fieldVal))
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
func RunModelAttr(inputCSV string, outputJSON string) error {
	cmd := exec.Command(
		"python",
		"Models/ModelAttr/AttrTree.py",
		"--input_csv", inputCSV,
		"--output_json", outputJSON,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("python execution failed: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}
func RunGeoPremiumCalc(inputJSON string, outputJSON string) error {
	cmd := exec.Command(
		"python",
		"Models/GeoPremiumCalc/GeoPremium.py",
		"--input_json", inputJSON,
		"--output_json", outputJSON,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("python execution failed: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}
func RunModelGeo(inputCSV string, inputJSON string, outputJSON string) error {
	cmd := exec.Command(
		"python",
		"Models/ModelGeo/GeoTree.py",
		"--input_csv", inputCSV,
		"--input_json", inputJSON,
		"--output_json", outputJSON,
	)

	
	var stderr bytes.Buffer
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("python execution failed: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}
func RunReport(inputCSV string, outputHTML string) error {
	cmd := exec.Command(
		"python",
		"Reports/ReportResults.py",
		"--input_csv", inputCSV,
		"--output_html", outputHTML,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("python execution failed: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}