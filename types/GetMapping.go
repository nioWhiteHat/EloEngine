package types

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func loadMappings(filename string) (map[string]map[string]float32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	mappings := make(map[string]map[string]float32)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) == 3 {
			category, key, valStr := parts[0], parts[1], parts[2]
			val, _ := strconv.ParseFloat(valStr, 32)
			if _, ok := mappings[category]; !ok {
				mappings[category] = make(map[string]float32)
			}
			mappings[category][key] = float32(val)
		}
	}
	return mappings, nil
}