package main

import (
	"strconv"
	"strings"
)

func getCalledFunctionsNames(jsonAST map[string]interface{}) []string {

	nodes := jsonAST["nodes"]
	functionsNamesList := make([]string, 0)
	functionsNamesList = storeCalledFunctionsNames(nodes, functionsNamesList)
	return functionsNamesList
}

func storeCalledFunctionsNames(node interface{}, functionsNamesList []string) []string {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			functionsNamesList = storeCalledFunctionsNames(element, functionsNamesList)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionCall" {
				if expressionNode, ok := nodeMap["expression"]; ok {
					expressionNodeMap := expressionNode.(map[string]interface{})
					functionName := expressionNodeMap["name"].(string)
					functionsNamesList = append(functionsNamesList, functionName)
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					functionsNamesList = storeCalledFunctionsNames(value, functionsNamesList)
				}
			}
		}
	}

	return functionsNamesList
}

func findFunctionDefinitionStart(node interface{}, functionName string) int {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			res := findFunctionDefinitionStart(element, functionName)
			if res != -1 {
				return res
			}
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionDefinition" {
				if name, ok := nodeMap["name"]; ok && name.(string) == functionName {
					nameLocationField := nodeMap["nameLocation"]
					nameLocationFieldParts := strings.Split((nameLocationField.(string)), ":")
					nameLocationStart, _ := strconv.Atoi(nameLocationFieldParts[0])
					return nameLocationStart
				}
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					res := findFunctionDefinitionStart(value, functionName)
					if res != -1 {
						return res
					}
				}
			}
		}
	}
	return -1
}

func findFunctionDefinition(sourceString string, jsonAST map[string]interface{}, functionName string) string {
	functionDefinitionStart := findFunctionDefinitionStart(jsonAST["nodes"], functionName)
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
