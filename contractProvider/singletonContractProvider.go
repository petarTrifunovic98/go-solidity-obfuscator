package contractprovider

import (
	"fmt"
	"solidity-obfuscator/helpers"
	"sync"
)

type solidityContract struct {
	sourceCode     string
	jsonCompactAST map[string]interface{}
}

var once sync.Once
var instance *solidityContract

func SolidityContractInstance(sourceCodePath string, jsonCompactASTPath string) *solidityContract {
	once.Do(func() {

		configFile, errConfig := helpers.ReadJsonToMap("../config/config.json")
		if errConfig != nil {
			fmt.Println(errConfig)
			instance = nil
			return
		}

		sourceCodeString, errSource := helpers.ReadFileToString(configFile[sourceCodePath].(string))
		if errSource != nil {
			fmt.Println(errSource)
			instance = nil
			return
		}

		jsonCompactASTString, errJson := helpers.ReadJsonToMap(configFile["jsonCompactASTPath"].(string))
		if errJson != nil {
			fmt.Println(errJson)
			instance = nil
			return
		}

		instance = &solidityContract{
			sourceCode:     sourceCodeString,
			jsonCompactAST: jsonCompactASTString,
		}
	})

	return instance
}
