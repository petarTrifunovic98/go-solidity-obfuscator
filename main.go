package main

import (
	"fmt"
	datastructs "solidity-obfuscator/dataStructs"
)

func main() {

	// jsonFile, errJson := os.Open("../contract_examples/contract_example_0813_2.sol_json.ast")
	// defer jsonFile.Close()
	// if errJson != nil {
	// 	fmt.Println(errJson)
	// 	return
	// }

	// byteValue, _ := ioutil.ReadAll(jsonFile)
	// var jsonStringMap map[string]interface{}
	// json.Unmarshal([]byte(byteValue), &jsonStringMap)

	// sourceFile, errSource := os.Open("../contract_examples/contract_example_0813_2.sol")
	// defer sourceFile.Close()
	// if errSource != nil {
	// 	fmt.Println(errSource)
	// 	return
	// }

	// byteValue, _ = ioutil.ReadAll(sourceFile)
	// sourceString := string(byteValue)

	// sourceString = ManipulateCalledFunctionsBodies()
	// sourceString = ReplaceVarNames()
	// sourceString = ReplaceComments()
	// sourceString = ReplaceLiterals()

	// outputFile, errOutput := os.Create("../contract_examples/obfuscated.sol")
	// defer outputFile.Close()
	// if errOutput != nil {
	// 	fmt.Println(errOutput)
	// 	return
	// }

	// outputFile.WriteString(sourceString)
	//generateTargetAST(12)

	asd := datastructs.RBTree[int]{}

	rootNode := datastructs.RBNode[int]{
		Data: 1,
		Key:  7,
	}
	asd.InsertAndGetParent(&rootNode)

	newNode := new(datastructs.RBNode[int])
	newNode.Key = 15
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 6
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 18
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 12
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 24
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 16
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 28
	asd.InsertAndGetParent(newNode)

	newNode = new(datastructs.RBNode[int])
	newNode.Key = 5
	asd.InsertAndGetParent(newNode)

	datastructs.InOrderTraversal(asd.Root)
	fmt.Println()
	datastructs.PreOrderTraversal(asd.Root)
	fmt.Println()
}
