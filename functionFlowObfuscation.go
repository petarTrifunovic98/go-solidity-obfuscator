package main

import (
	"fmt"
	"strconv"
	"strings"
)

func getCalledFunctionsNames(jsonAST map[string]interface{}) []string {

	nodes := jsonAST["nodes"]
	functionsNamesList := make([]string, 0)
	functionsNamesList = storeCalledFunctionsNames(nodes, nodes, functionsNamesList)
	return functionsNamesList
}

func storeCalledFunctionsNames(rootNode interface{}, node interface{}, functionsNamesList []string) []string {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			functionsNamesList = storeCalledFunctionsNames(rootNode, element, functionsNamesList)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionCall" {
				expressionNode := nodeMap["expression"]
				expressionNodeMap := expressionNode.(map[string]interface{})
				functionName := expressionNodeMap["name"].(string)
				functionsNamesList = append(functionsNamesList, functionName)
				if findFunctionDefinitionNode(rootNode, functionName) != nil {
					fmt.Println(findFunctionCallArgumentValues(nodeMap))
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					functionsNamesList = storeCalledFunctionsNames(rootNode, value, functionsNamesList)
				}
			}
		}
	}

	return functionsNamesList
}

func findFunctionDefinitionNode(node interface{}, functionName string) map[string]interface{} {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			res := findFunctionDefinitionNode(element, functionName)
			if res != nil {
				return res
			}
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionDefinition" {
				if name, ok := nodeMap["name"]; ok && name.(string) == functionName {
					return nodeMap
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					res := findFunctionDefinitionNode(value, functionName)
					if res != nil {
						return res
					}
				}
			}
		}
	}
	return nil
}

func findFunctionDefinitionBody(sourceString string, jsonAST map[string]interface{}, functionName string) string {
	functionDefinitionNode := findFunctionDefinitionNode(jsonAST["nodes"], functionName)
	nameLocationField := functionDefinitionNode["nameLocation"]
	nameLocationFieldParts := strings.Split((nameLocationField.(string)), ":")
	functionDefinitionStart, _ := strconv.Atoi(nameLocationFieldParts[0])
	sourceString = sourceString[functionDefinitionStart:]
	functionBodyStartIndex := strings.Index(sourceString, "{")
	index := functionBodyStartIndex + 1
	counter := 1

	for counter > 0 {
		if sourceString[index] == '{' {
			counter++
		} else if sourceString[index] == '}' {
			counter--
		}
		index++
	}

	return sourceString[functionBodyStartIndex+1 : index-1]
}

func findFunctionParametersNames(jsonAST map[string]interface{}, functionName string) []string {
	functionDefinitionNode := findFunctionDefinitionNode(jsonAST["nodes"], functionName)
	parametersField := functionDefinitionNode["parameters"].(map[string]interface{})
	parametersList := parametersField["parameters"]

	parameterNamesList := make([]string, 0)

	for _, parameterInterface := range parametersList.([]interface{}) {
		parameterMap := parameterInterface.(map[string]interface{})
		parameterNamesList = append(parameterNamesList, parameterMap["name"].(string))
	}

	return parameterNamesList
}

func findFunctionCallArgumentValues(functionCallNodeMap map[string]interface{}) []string {
	argumentsList := functionCallNodeMap["arguments"].([]interface{})

	argumentValuesList := make([]string, 0)

	for _, argumentInterface := range argumentsList {
		argumentMap := argumentInterface.(map[string]interface{})
		argumentValuesList = append(argumentValuesList, argumentMap["value"].(string))
	}

	return argumentValuesList
}
