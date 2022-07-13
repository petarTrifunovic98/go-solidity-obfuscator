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
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	namesList := getVarNames(result)

	sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	defer sourceFile.Close()
	if errSource != nil {
		fmt.Println(errSource)
		return
	}

	byteValue, _ = ioutil.ReadAll(sourceFile)
	sourceString := string(byteValue)

	sourceString = replaceVarNames(namesList, sourceString)
	//fmt.Println(sourceString)

	sourceString = replaceComments(sourceString)

	outputFile, errOutput := os.Create("../contract_examples/obfuscated.sol")
	defer outputFile.Close()
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}

	outputFile.WriteString(sourceString)

}
