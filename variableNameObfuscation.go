package main

import (
	"regexp"
	contractprovider "solidity-obfuscator/contractProvider"
)

func getVarNames(jsonAST map[string]interface{}) []string {
	nodes := jsonAST["nodes"]
	namesList := make([]string, 0)
	namesList = storeVarNames(nodes, namesList)
	return namesList
}

func storeVarNames(node interface{}, namesList []string) []string {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			namesList = storeVarNames(element, namesList)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "VariableDeclaration" {
				if name, ok := nodeMap["name"]; ok && name.(string) != "" {
					namesList = append(namesList, name.(string))
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					namesList = storeVarNames(value, namesList)
				}
			}
		}
	}

	return namesList
}

func ReplaceVarNames() string {

	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()
	namesList := getVarNames(jsonAST)

	// starting name can not be one dash, since that is a reserved name
	var newVarName string = "__"
	nameIsUsed := make(map[string]bool)

	for _, name := range namesList {
		if !nameIsUsed[name] {
			re, _ := regexp.Compile("\\b" + name + "\\b")
			sourceCodeString = re.ReplaceAllString(sourceCodeString, newVarName)
			nameIsUsed[name] = true
			newVarName += "_"
		}
	}

	contract.SetSourceCode(sourceCodeString)

	return sourceCodeString
}
