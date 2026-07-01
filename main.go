package main

import (
	"bytes"
	"log"
	onnxruntime "github.com/yalue/onnxruntime_go"

	"context"
	db "eiv-ranking/Database"
	elo "eiv-ranking/cmd/XGboost/EloCalculator"
	t "eiv-ranking/cmd/XGboost/types"
	"encoding/csv"

	//"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/jackc/pgx/v5"
	//testing "eiv-ranking/cmd/XGboost/Testing"
	//p "eiv-ranking/cmd/XGboost/PredictingPrice"
)
func main() {
	con,_ := db.InitDB()
	//BuildFiles(con)
	err := RunModelAttr("Modeled_Sold.csv", "Models/JsonModelData/ModelAttr_Sold.json")
	if err != nil {
		fmt.Printf("Error running ModelAttr: %v\n", err)
	}
	err = RunGeoPremiumCalc("Models/JsonModelData/ModelAttr_Sold.json", "Models/JsonModelData/GeoPremium_Sold.json")
	if err != nil {
		fmt.Printf("Error running GeoCalc: %v\n", err)
	}
	err = RunModelGeo("Modeled_Sold.csv", "Models/JsonModelData/GeoPremium_Sold.json", "Models/JsonModelData/ModelGeo_Sold.json")
	if err != nil {
		fmt.Printf("Error running Final: %v\n", err)
	}
	elo.MergeMLDataToCSV("Modeled_Sold.csv","Models/JsonModelData/ModelGeo_Sold.json","Modeled_Sold_ML.csv")
	elo.RunEloPipeline(context.Background(),con,"Modeled_Sold_ML.csv","TestingElo2.json")
	/*Mappings, err := p.LoadCategoryMappings("Models/Saved/CategoryMappings.txt")
	if err != nil {
		log.Fatal(err)
	}
	/*prop := p.GetProperty()
	features := p.PrepareFeatures(prop, Mappings)
	predictedPrice, err := p.PredictSinglePoint("Models/Saved/ModelAttr_Trees.onnx", features)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Predicted Price per Sqm: %.2f\n", predictedPrice)*/
	
	



}


func TestPrediction(prop t.PropertyModel) {
	onnxruntime.SetSharedLibraryPath("./onnxengine.dll")

	err := onnxruntime.InitializeEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	defer onnxruntime.DestroyEnvironment()

	catMap := map[string]float32{
		"apartment":                                   0.0,
		"autonomous_individual":                       1.0,
		"":                                            -1.0,
		"Ζ":                                           19.0,
		"Περιφερειακή Ενότητα Κεντρικού Τομέα Αθηνών": 34.0,
	}

	inputData := []float32{
		float32(prop.Age),
		float32(prop.RenovAge),
		float32(prop.IsRenovated),
		catMap[prop.SubType],
		float32(prop.SquareMeters),
		float32(prop.LandplotSize),
		float32(prop.NoView),
		float32(prop.ViewForest),
		float32(prop.ViewMountain),
		float32(prop.ViewSea),
		float32(prop.ViewUnlimited),
		float32(prop.NoFacade),
		float32(prop.FacadeCorner),
		float32(prop.FacadeFrontage),
		float32(prop.FacadeInterior),
		float32(prop.FacadeThrough),
		catMap[prop.HeatType],
		catMap[prop.HeatFrom],
		catMap[prop.EnergyPerformance],
		float32(prop.WC),
		float32(prop.AC),
		float32(prop.Elevator),
		float32(prop.Parking),
		float32(prop.StorageRoom),
		float32(prop.AluminumFrames),
		float32(prop.Bedrooms),
		float32(prop.Pool),
		float32(prop.Garden),
		float32(prop.FloorLevel),
		float32(prop.FloorCount),
		float32(prop.ZoneValue),
		float32(prop.MarketabillityFactor),
		catMap[prop.AdminName],
	}
	fmt.Printf("GO_INPUT: %v\n", inputData)
	inputTensor, err := onnxruntime.NewTensor(onnxruntime.Shape{1, 33}, inputData)
	if err != nil {
		log.Fatal(err)
	}
	defer inputTensor.Destroy()

	outputData := make([]float32, 1)
	outputTensor, err := onnxruntime.NewTensor(onnxruntime.Shape{1, 1}, outputData)
	if err != nil {
		log.Fatal(err)
	}
	defer outputTensor.Destroy()

	

	session, err := onnxruntime.NewAdvancedSession("Models/Saved/ModelAttr_Trees.onnx",
		[]string{"input"},
		[]string{"variable"},
		[]onnxruntime.Value{inputTensor},
		[]onnxruntime.Value{outputTensor},
		nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Destroy()

	err = session.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Predicted Price per Sqm: %.2f\n", outputData[0])
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
func RunSeeAttrNames() error {
	cmd := exec.Command(
		"python",
		"GetCategoricalAttrNames.py",
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



