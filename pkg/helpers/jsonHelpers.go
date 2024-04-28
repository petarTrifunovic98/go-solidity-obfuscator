package helpers

import (
	"encoding/json"
	"io"
	"os"
)

func ReadJsonToMap(jsonFilePath string) (map[string]interface{}, error) {
	jsonFile, errJson := os.Open(jsonFilePath)
	if errJson != nil {
		return nil, errJson
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonMap)

	return jsonMap, nil
}
