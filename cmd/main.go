package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/obfuscator"
)

func main() {

	jsonFile, errJson := os.Open("../contract_examples/contract_example_0813_2.sol_json.ast")
	if errJson != nil {
		fmt.Println(errJson)
		return
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var jsonStringMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &jsonStringMap)

	sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	if errSource != nil {
		fmt.Println(errSource)
		return
	}
	defer sourceFile.Close()

	byteValue, _ = io.ReadAll(sourceFile)
	sourceString := string(byteValue)

	sourceString = obfuscator.ManipulateDefinedFunctionBodies()
	sourceString = obfuscator.ManipulateCalledFunctionsBodies()
	sourceString = obfuscator.ReplaceVarNames()
	sourceString = obfuscator.ReplaceComments()
	sourceString = obfuscator.ReplaceLiterals()

	outputFile, errOutput := os.Create("../contract_examples/contract_example_0813_2_obf.sol")
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}
	defer outputFile.Close()

	outputFile.WriteString(sourceString)
}
