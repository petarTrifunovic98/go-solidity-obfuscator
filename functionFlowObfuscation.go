package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FunctionDefinition struct {
	body              string
	parameterNames    []string
	retParameterTypes []string
}

type FunctionCall struct {
	name          string
	args          []string
	indexInSource int
	callLen       int
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
				functionSrcParts := strings.Split(functionSrc, ":")
				functionStartIndex, _ := strconv.Atoi(functionSrcParts[0])
				functionCallLen, _ := strconv.Atoi(functionSrcParts[1])
				functionCall := FunctionCall{
					name:          functionName,
					args:          argsList,
					indexInSource: functionStartIndex,
					callLen:       functionCallLen,
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

func findFunctionRetParameterTypes(functionDefinitionNode map[string]interface{}) []string {
	retParametersField := functionDefinitionNode["returnParameters"].(map[string]interface{})
	retParametersList := retParametersField["parameters"]

	retParametersTypesList := make([]string, 0)

	for _, retParameterInterface := range retParametersList.([]interface{}) {
		retParameterMap := retParameterInterface.(map[string]interface{})
		retParameterTypeDesc := retParameterMap["typeDescriptions"].(map[string]interface{})
		retParametersTypesList = append(retParametersTypesList, retParameterTypeDesc["typeString"].(string))
	}

	return retParametersTypesList
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
	retParametersNames := findFunctionRetParameterTypes(functionDefinitionNode)

	functionDefinition := FunctionDefinition{
		body:              body,
		parameterNames:    parametersNames,
		retParameterTypes: retParametersNames,
	}

	return &functionDefinition
}

func replaceFunctionParametersWithArguments(functionDefinition *FunctionDefinition, functionArguments []string) {
	body := functionDefinition.body
	parameters := functionDefinition.parameterNames

	i := 0
	for _, parameter := range parameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		body = re.ReplaceAllString(body, functionArguments[i])
		i++
	}

	functionDefinition.body = body
}

func replaceReturnStmtWithVariables(functionDefinition *FunctionDefinition) {
	body := functionDefinition.body
	retParameterTypes := functionDefinition.retParameterTypes
	re, _ := regexp.Compile("\\breturn\\b")
	retStmtIndexes := re.FindAllStringIndex(body, -1)
	if retStmtIndexes == nil {
		return
	}

	currentVarName := "__"
	stringIncrease := 0
	var insertString string

	for _, indexPair := range retStmtIndexes {
		retStmtStartIndex := indexPair[0]
		retStmtEndIndex := retStmtStartIndex
		for body[retStmtEndIndex] != ';' {
			retStmtEndIndex++
		}

		retValuesList := strings.Split(body[retStmtStartIndex+len("return")+stringIncrease:retStmtEndIndex+stringIncrease], ",;")

		if len(retValuesList) > 0 {
			insertString = "\n{\n"
			for i := 0; i < len(retValuesList); i++ {
				retValue := strings.Trim(retValuesList[i], " \t\n")
				insertString += retParameterTypes[i] + " " + currentVarName + " = " + retValue + ";\n"
				currentVarName += "_"
			}
			insertString += "}\n"
		}

		body = body[:retStmtStartIndex+stringIncrease] + insertString + body[retStmtEndIndex+stringIncrease+1:]
		stringIncrease += len(insertString)

	}
	functionDefinition.body = body

}

func manipulateCalledFunctionsBodies(jsonAST map[string]interface{}, sourceString string) map[string][]string {

	nodes := jsonAST["nodes"]
	functionCalls := make([]*FunctionCall, 0)
	functionCalls = storeCalledFunctions(nodes, functionCalls, sourceString)

	//extractedFunctionDefinitions := make(map[string]*FunctionDefinition, 0)

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
		//functionDef, exists := extractedFunctionDefinitions[functionCall.name]
		//if !exists {
		functionDef := extractFunctionDefinition(nodes, functionCall.name, originalSourceString)
		//}
		if functionDef != nil {
			replaceFunctionParametersWithArguments(functionDef, functionCall.args)
			replaceReturnStmtWithVariables(functionDef)
			newFuncBodies[functionCall.name] = append(newFuncBodies[functionCall.name], functionDef.body)

			funcCallStart := functionCall.indexInSource
			funcCallEnd := functionCall.indexInSource + functionCall.callLen

			i := funcCallStart + stringIndexIncrease
			for sourceString[i] != ';' && sourceString[i] != '{' && sourceString[i] != '}' {
				i--
			}
			sourceString = sourceString[:i+1] + "\n" + functionDef.body + sourceString[i+1:]
			stringIndexIncrease += len(functionDef.body) + 1

			sourceString = sourceString[:funcCallStart+stringIndexIncrease] + "__" + sourceString[funcCallEnd+stringIndexIncrease:]
			fmt.Println(funcCallStart + stringIndexIncrease)
			stringIndexIncrease += len("__") - functionCall.callLen

		}
	}

	fmt.Println(sourceString)
	//fmt.Println(newFuncBodies)

	return newFuncBodies

}
