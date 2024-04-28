package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/obfuscator"
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

	sourceString = obfuscator.ManipulateDefinedFunctionBodies()
	sourceString = obfuscator.ManipulateCalledFunctionsBodies()
	sourceString = obfuscator.ReplaceVarNames()
	sourceString = obfuscator.ReplaceComments()
	sourceString = obfuscator.ReplaceLiterals()

	outputFile, errOutput := os.Create("../contract_examples/contract_example_0813_2_obf.sol")
	defer outputFile.Close()
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}

	outputFile.WriteString(sourceString)
}
