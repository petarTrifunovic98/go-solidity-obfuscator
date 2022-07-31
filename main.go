package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {

	jsonFile, errJson := os.Open("../contract_examples/contract_example_0813_2.sol_json.ast")
	defer jsonFile.Close()
	if errJson != nil {
		fmt.Println(errJson)
		return
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var jsonStringMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonStringMap)

	sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	defer sourceFile.Close()
	if errSource != nil {
		fmt.Println(errSource)
		return
	}

	namesList := getVarNames(jsonStringMap)
	literalsList := getLiterals(jsonStringMap)

	byteValue, _ = ioutil.ReadAll(sourceFile)
	sourceString := string(byteValue)

	fmt.Println(findFunctionDefinition(sourceString, jsonStringMap, "privateFunc"))

	sourceString = replaceVarNames(namesList, sourceString)
	sourceString = replaceComments(sourceString)
	sourceString = replaceLiterals(literalsList, sourceString)

	outputFile, errOutput := os.Create("../contract_examples/obfuscated.sol")
	defer outputFile.Close()
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}

	outputFile.WriteString(sourceString)
	//generateTargetAST(12)

}
