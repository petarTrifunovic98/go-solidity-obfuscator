package processinfo

import (
	"strconv"
	"strings"
	"sync"
)

type FunctionBody struct {
	BodyContent   string
	IndexInSource int
}

type FunctionDefinition struct {
	Name                  string
	Body                  FunctionBody
	ParameterNames        []string
	RetParameterTypes     []string
	TopLevelDeclarations  [][2]int
	IndependentStatements [][2]int
}

type FunctionCall struct {
	Name                string
	ArgsOld             []string
	Args                [][2]int
	IndexInSource       int
	CallLen             int
	CallingFunctionName string
}

type functionInformation struct {
	functionDefinitions map[string]*FunctionDefinition
	functionCalls       []*FunctionCall
}

var functionInfoOnce sync.Once
var functionInfoInstance *functionInformation

func FunctionInformation() *functionInformation {
	functionInfoOnce.Do(func() {
		functionInfoInstance = &functionInformation{
			functionDefinitions: make(map[string]*FunctionDefinition),
			functionCalls:       nil,
		}
	})

	return functionInfoInstance
}

func (fi *functionInformation) GetFunctionCalls() []*FunctionCall {
	return fi.functionCalls
}

func (fi *functionInformation) GetFunctionDefinitions() map[string]*FunctionDefinition {
	return fi.functionDefinitions
}

func (fi *functionInformation) GetSpecificFunctionDefinition(name string) (*FunctionDefinition, bool) {
	functionDef, ok := fi.functionDefinitions[name]
	return functionDef, ok
}

func (fi *functionInformation) ExtractFunctionCalls(jsonAST map[string]interface{}, sourceString string) []*FunctionCall {
	nodes := jsonAST["nodes"]
	fi.functionCalls = make([]*FunctionCall, 0)
	var functionDef string
	fi.functionCalls = fi.storeFunctionCalls(nodes, sourceString, functionDef)
	// fmt.Println(fi.functionCalls)
	return fi.functionCalls
}

func (fi *functionInformation) storeFunctionCalls(nodeInterface interface{}, sourceString string, latestFunctionDef string) []*FunctionCall {
	switch node := nodeInterface.(type) {
	case []interface{}:
		for _, element := range node {
			fi.functionCalls = fi.storeFunctionCalls(element, sourceString, latestFunctionDef)
		}
	case map[string]interface{}:
		if fieldType, ok := node["nodeType"]; ok && fieldType == "FunctionDefinition" {
			latestFunctionDef = node["name"].(string)
		}
		for key, value := range node {
			if key == "nodeType" && value == "FunctionCall" {
				expressionNode := node["expression"]
				expressionNodeMap := expressionNode.(map[string]interface{})
				if _, ok := expressionNodeMap["name"]; !ok {
					continue
				}
				functionName := expressionNodeMap["name"].(string)
				argsList := findFunctionCallArgumentValuesOld(node, sourceString)
				args := findFunctionCallArgumentValues(node)
				functionSrc := node["src"].(string)
				functionSrcParts := strings.Split(functionSrc, ":")
				functionStartIndex, _ := strconv.Atoi(functionSrcParts[0])
				functionCallLen, _ := strconv.Atoi(functionSrcParts[1])
				functionCall := FunctionCall{
					Name:                functionName,
					ArgsOld:             argsList,
					Args:                args,
					IndexInSource:       functionStartIndex,
					CallLen:             functionCallLen,
					CallingFunctionName: latestFunctionDef,
				}
				fi.functionCalls = append(fi.functionCalls, &functionCall)

			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					fi.functionCalls = fi.storeFunctionCalls(value, sourceString, latestFunctionDef)
				}
			}
		}
	}

	return fi.functionCalls
}

func findFunctionCallArgumentValuesOld(functionCallNodeMap map[string]interface{}, sourceString string) []string {
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

func findFunctionCallArgumentValues(functionCallNodeMap map[string]interface{}) [][2]int {
	argumentsList := functionCallNodeMap["arguments"].([]interface{})

	argumentsIndexesList := make([][2]int, 0)

	for _, argumentInterface := range argumentsList {
		argumentMap := argumentInterface.(map[string]interface{})
		argumentSrc := argumentMap["src"].(string)
		argumentSrcParts := strings.Split(argumentSrc, ":")
		argumentStart, _ := strconv.Atoi(argumentSrcParts[0])
		argumentLen, _ := strconv.Atoi(argumentSrcParts[1])

		argumentsIndexesList = append(argumentsIndexesList, [2]int{argumentStart, argumentLen})
	}

	return argumentsIndexesList
}

func (fi *functionInformation) ExtractFunctionDefinition(jsonAST map[string]interface{}, functionName string, sourceString string) *FunctionDefinition {
	node := jsonAST["nodes"]
	functionDefinitionNode := findFunctionDefinitionNode(node, functionName)
	if functionDefinitionNode == nil {
		fi.functionDefinitions[functionName] = nil
		return nil
	}

	body := findFunctionDefinitionBody(functionDefinitionNode, sourceString)
	parametersNames := findFunctionParametersNames(functionDefinitionNode)
	retParametersNames := findFunctionRetParameterTypes(functionDefinitionNode)
	independentStmts, topLevelDeclarations := findFunctionStatementsAndDeclarations(functionDefinitionNode)

	functionDefinition := FunctionDefinition{
		Name:                  functionName,
		Body:                  body,
		ParameterNames:        parametersNames,
		RetParameterTypes:     retParametersNames,
		TopLevelDeclarations:  topLevelDeclarations,
		IndependentStatements: independentStmts,
	}

	fi.functionDefinitions[functionName] = &functionDefinition

	return &functionDefinition
}

func (fi *functionInformation) ExtractAllFunctionDefinitions(jsonAST map[string]interface{}, sourceCode string) map[string]*FunctionDefinition {
	nodes := jsonAST["nodes"]
	functionDefinitionNodes := make([]map[string]interface{}, 0)
	functionDefinitionNodes = storeAllFunctionDefinitionNodes(nodes, functionDefinitionNodes)

	for _, functionDefinitionNode := range functionDefinitionNodes {
		if functionDefinitionNode != nil {
			name := functionDefinitionNode["name"].(string)
			body := findFunctionDefinitionBody(functionDefinitionNode, sourceCode)
			parameterNames := findFunctionParametersNames(functionDefinitionNode)
			retParameterNames := findFunctionRetParameterTypes(functionDefinitionNode)
			independentStmts, topLevelDeclarations := findFunctionStatementsAndDeclarations(functionDefinitionNode)

			fi.functionDefinitions[name] = &FunctionDefinition{
				Name:                  name,
				Body:                  body,
				ParameterNames:        parameterNames,
				RetParameterTypes:     retParameterNames,
				TopLevelDeclarations:  topLevelDeclarations,
				IndependentStatements: independentStmts,
			}
		}
	}

	return fi.functionDefinitions
}

func storeAllFunctionDefinitionNodes(nodeInterface interface{}, functionNodes []map[string]interface{}) []map[string]interface{} {
	switch node := nodeInterface.(type) {
	case []interface{}:
		for _, element := range node {
			functionNodes = storeAllFunctionDefinitionNodes(element, functionNodes)
		}
	case map[string]interface{}:
		for key, value := range node {
			if key == "nodeType" && value == "FunctionDefinition" {
				functionNodes = append(functionNodes, node)
			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					functionNodes = storeAllFunctionDefinitionNodes(value, functionNodes)
				}
			}
		}
	}

	return functionNodes
}

func findFunctionDefinitionNode(nodeInterface interface{}, functionName string) map[string]interface{} {
	switch node := nodeInterface.(type) {
	case []interface{}:
		for _, element := range node {
			res := findFunctionDefinitionNode(element, functionName)
			if res != nil {
				return res
			}
		}
	case map[string]interface{}:
		for key, value := range node {
			if key == "nodeType" && value == "FunctionDefinition" {
				if name, ok := node["name"]; ok && name.(string) == functionName {
					return node
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

func findFunctionDefinitionBody(functionDefinitionNodeMap map[string]interface{}, sourceString string) FunctionBody {
	funcDefLocationField := functionDefinitionNodeMap["src"]
	funcDefLocationFieldParts := strings.Split((funcDefLocationField.(string)), ":")
	functionDefinitionStart, _ := strconv.Atoi(funcDefLocationFieldParts[0])
	sourceString = sourceString[functionDefinitionStart:]
	functionBodyStartIndex := strings.Index(sourceString, "{")
	indexInSource := functionDefinitionStart + functionBodyStartIndex
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

	return FunctionBody{
		BodyContent:   sourceString[functionBodyStartIndex+1 : index-1],
		IndexInSource: indexInSource + 1,
	}
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

func findFunctionRetParameterTypes(functionDefinitionNodeMap map[string]interface{}) []string {
	retParametersField := functionDefinitionNodeMap["returnParameters"].(map[string]interface{})
	retParametersList := retParametersField["parameters"]

	retParametersTypesList := make([]string, 0)

	for _, retParameterInterface := range retParametersList.([]interface{}) {
		retParameterMap := retParameterInterface.(map[string]interface{})
		retParameterTypeDesc := retParameterMap["typeDescriptions"].(map[string]interface{})
		retParameterType := retParameterTypeDesc["typeString"].(string)
		retParameterStorageLocation := retParameterMap["storageLocation"].(string)
		if retParameterStorageLocation != "default" {
			retParameterType += " " + retParameterStorageLocation
		}
		retParametersTypesList = append(retParametersTypesList, retParameterType)
	}

	return retParametersTypesList
}

func findFunctionStatementsAndDeclarations(functionDefinitionNodeMap map[string]interface{}) ([][2]int, [][2]int) {
	bodyField := functionDefinitionNodeMap["body"].(map[string]interface{})
	statementsList := bodyField["statements"].([]interface{})
	independentStmts := make([][2]int, 0)
	topLevelDecls := make([][2]int, 0)

	for _, statementInterface := range statementsList {
		statementMap := statementInterface.(map[string]interface{})
		statementSrc := statementMap["src"].(string)
		statementSrcParts := strings.Split(statementSrc, ":")
		statementStart, _ := strconv.Atoi(statementSrcParts[0])
		statementLen, _ := strconv.Atoi(statementSrcParts[1])
		if statementMap["nodeType"].(string) == "VariableDeclarationStatement" {
			topLevelDecls = append(topLevelDecls, [2]int{statementStart, statementLen})

			//REMOVE
			//independentStmts = append(independentStmts, [2]int{statementStart, statementLen})
		} else {
			independentStmts = append(independentStmts, [2]int{statementStart, statementLen})
		}
	}

	return independentStmts, topLevelDecls
}
