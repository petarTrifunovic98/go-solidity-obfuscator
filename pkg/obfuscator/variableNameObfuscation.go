package obfuscator

import (
	"regexp"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/contractprovider"
	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/processinfo"
)

func getVarNames(jsonAST map[string]interface{}) map[string]struct{} {
	nodes := jsonAST["nodes"]
	namesSet := make(map[string]struct{}, 0)
	namesSet = storeVarNames(nodes, namesSet)
	return namesSet
}

// maybe move this function fo variableInformation
func storeVarNames(node interface{}, namesSet map[string]struct{}) map[string]struct{} {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			namesSet = storeVarNames(element, namesSet)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "VariableDeclaration" {
				if name, ok := nodeMap["name"]; ok && name != "" {
					namesSet[name.(string)] = struct{}{}
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					namesSet = storeVarNames(value, namesSet)
				}
			}
		}
	}

	return namesSet
}

func ReplaceVarNames() string {

	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()

	variableInfo := processinfo.VariableInformation()
	namesSet := variableInfo.GetVariableNamesSet()
	if namesSet == nil {
		namesSet = getVarNames(jsonAST)
		variableInfo.SetVariableNamesSet(namesSet)
	}

	// starting name can not be one dash, since that is a reserved name
	var newVarName string = variableInfo.GetLatestDashVariableName() + "_"
	//nameIsUsed := make(map[string]bool)

	//if thread safety is required, change this to always check the latest var name
	for name := range namesSet {
		for variableInfo.NameIsUsed(newVarName) {
			newVarName += "_"
		}
		re, _ := regexp.Compile("\\b" + name + "\\b")
		sourceCodeString = re.ReplaceAllString(sourceCodeString, newVarName)
		newVarName += "_"
	}

	contract.SetSourceCode(sourceCodeString)
	variableInfo.SetLatestDashVariableName(newVarName)

	return sourceCodeString
}
