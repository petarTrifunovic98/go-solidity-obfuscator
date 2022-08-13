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

	byteValue, _ = ioutil.ReadAll(sourceFile)
	sourceString := string(byteValue)

	sourceString = ManipulateCalledFunctionsBodies()
	sourceString = ReplaceVarNames()
	sourceString = ReplaceComments()
	sourceString = ReplaceLiterals()

	outputFile, errOutput := os.Create("../contract_examples/obfuscated.sol")
	defer outputFile.Close()
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}

	outputFile.WriteString(sourceString)
	//generateTargetAST(12)

}
