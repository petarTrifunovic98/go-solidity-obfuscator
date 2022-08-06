package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FunctionDefinition struct {
	body           string
	parameterNames []string
}

type FunctionCall struct {
	name          string
	args          []string
	indexInSource int
}

func getCalledFunctionsNames(jsonAST map[string]interface{}, sourceString string) []*FunctionCall {

	nodes := jsonAST["nodes"]
	functionsCalls := make([]*FunctionCall, 0)
	functionsCalls = storeCalledFunctions(nodes, functionsCalls, sourceString)
	return functionsCalls
}

func storeCalledFunctions(node interface{}, functionsCalls []*FunctionCall, sourceString string) []*FunctionCall {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			functionsCalls = storeCalledFunctions(element, functionsCalls, sourceString)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionCall" {
				expressionNode := nodeMap["expression"]
				expressionNodeMap := expressionNode.(map[string]interface{})
				functionName := expressionNodeMap["name"].(string)
				argsList := findFunctionCallArgumentValues(nodeMap, sourceString)
				functionSrc := nodeMap["src"].(string)
				functionStartIndex, _ := strconv.Atoi(strings.Split(functionSrc, ":")[0])
				functionCall := FunctionCall{
					name:          functionName,
					args:          argsList,
					indexInSource: functionStartIndex,
				}
				functionsCalls = append(functionsCalls, &functionCall)

			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					functionsCalls = storeCalledFunctions(value, functionsCalls, sourceString)
				}
			}
		}
	}

	return functionsCalls
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

func findFunctionDefinitionBody(functionDefinitionNode map[string]interface{}, sourceString string) string {
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

func findFunctionParametersNames(functionDefinitionNode map[string]interface{}) []string {
	parametersField := functionDefinitionNode["parameters"].(map[string]interface{})
	parametersList := parametersField["parameters"]

	parameterNamesList := make([]string, 0)

	for _, parameterInterface := range parametersList.([]interface{}) {
		parameterMap := parameterInterface.(map[string]interface{})
		parameterNamesList = append(parameterNamesList, parameterMap["name"].(string))
	}

	return parameterNamesList
}

func findFunctionCallArgumentValues(functionCallNodeMap map[string]interface{}, sourceString string) []string {
	argumentsList := functionCallNodeMap["arguments"].([]interface{})

	argumentValuesList := make([]string, 0)

	for _, argumentInterface := range argumentsList {
		argumentMap := argumentInterface.(map[string]interface{})
		argumentSrc := argumentMap["src"].(string)
		argumentSrcParts := strings.Split(argumentSrc, ":")
		argumentStart, _ := strconv.Atoi(argumentSrcParts[0])
		argumentLen, _ := strconv.Atoi(argumentSrcParts[1])

		argumentValuesList = append(argumentValuesList, sourceString[argumentStart:argumentStart+argumentLen])
	}

	return argumentValuesList
}

func extractFunctionDefinition(node interface{}, functionName string, sourceString string) *FunctionDefinition {
	functionDefinitionNode := findFunctionDefinitionNode(node, functionName)
	if functionDefinitionNode == nil {
		return nil
	}

	body := findFunctionDefinitionBody(functionDefinitionNode, sourceString)
	parametersNames := findFunctionParametersNames(functionDefinitionNode)

	functionDefinition := FunctionDefinition{
		body:           body,
		parameterNames: parametersNames,
	}

	return &functionDefinition
}

func replaceFunctionParametersForArguments(functionDefinition *FunctionDefinition, functionArguments []string) string {
	body := functionDefinition.body
	parameters := functionDefinition.parameterNames

	i := 0
	for _, parameter := range parameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		body = re.ReplaceAllString(body, functionArguments[i])
		i++
	}

	return body
}

func manipulateCalledFunctionsBodies(jsonAST map[string]interface{}, sourceString string) map[string][]string {

	nodes := jsonAST["nodes"]
	functionCalls := make([]*FunctionCall, 0)
	functionCalls = storeCalledFunctions(nodes, functionCalls, sourceString)

	extractedFunctionDefinitions := make(map[string]*FunctionDefinition, 0)

	newFuncBodies := make(map[string][]string, 0)

	sort.Slice(functionCalls, func(i, j int) bool {
		return functionCalls[i].indexInSource < functionCalls[j].indexInSource
	})

	stringIndexIncrease := 0

	var sb strings.Builder
	if _, err := sb.WriteString(sourceString); err != nil {
		fmt.Println("error copying string!")
		fmt.Println(err)
		return nil
	}

	originalSourceString := sb.String()

	for _, functionCall := range functionCalls {
		functionDef, exists := extractedFunctionDefinitions[functionCall.name]
		if !exists {
			functionDef = extractFunctionDefinition(nodes, functionCall.name, originalSourceString)
		}
		if functionDef != nil {
			body := replaceFunctionParametersForArguments(functionDef, functionCall.args)
			fmt.Print(functionDef)
			fmt.Print(" : ")
			fmt.Println(functionCall.args)
			newFuncBodies[functionCall.name] = append(newFuncBodies[functionCall.name], body)

			i := functionCall.indexInSource + stringIndexIncrease
			for sourceString[i] != ';' && sourceString[i] != '{' {
				i--
			}
			sourceString = sourceString[:i+1] + "\n" + body + sourceString[i+1:]
			stringIndexIncrease += len(body) + 1
		}
	}

	fmt.Println(sourceString)

	return newFuncBodies

}
