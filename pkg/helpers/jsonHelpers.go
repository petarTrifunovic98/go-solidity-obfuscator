package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadJsonToMap(jsonFilePath string) (map[string]interface{}, error) {
	jsonFile, errJson := os.Open(jsonFilePath)
	defer jsonFile.Close()
	if errJson != nil {
		return nil, errJson
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonMap)

	return jsonMap, nil
}
