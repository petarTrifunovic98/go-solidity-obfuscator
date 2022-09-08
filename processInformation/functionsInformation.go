package processinformation

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type FunctionBody struct {
	BodyContent   string
	IndexInSource int
}

type FunctionDefinition struct {
	Name                        string
	Body                        FunctionBody
	ParameterNames              []string
	RetParameterTypes           []string
	IndependentStatements       []string
	TopLevelDeclarationsIndexes []int
}

type FunctionCall struct {
	Name          string
	Args          []string
	IndexInSource int
	CallLen       int
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
	fi.functionCalls = fi.storeFunctionCalls(nodes, sourceString)
	fmt.Println(fi.functionCalls)
	return fi.functionCalls
}

func (fi *functionInformation) storeFunctionCalls(node interface{}, sourceString string) []*FunctionCall {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			fi.functionCalls = fi.storeFunctionCalls(element, sourceString)
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
					Name:          functionName,
					Args:          argsList,
					IndexInSource: functionStartIndex,
					CallLen:       functionCallLen,
				}
				fi.functionCalls = append(fi.functionCalls, &functionCall)

			} else {
				_, okArr := value.([]interface{})
				_, okMap := value.(map[string]interface{})

				if okArr || okMap {
					fi.functionCalls = fi.storeFunctionCalls(value, sourceString)
				}
			}
		}
	}

	return fi.functionCalls
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
	topLevelDeclarationsIndexes := findFunctionTopLevelDeclarationStatements(functionDefinitionNode)
	//independentStatements := findIndependentStatements(body)

	functionDefinition := FunctionDefinition{
		Name:                        functionName,
		Body:                        body,
		ParameterNames:              parametersNames,
		RetParameterTypes:           retParametersNames,
		TopLevelDeclarationsIndexes: topLevelDeclarationsIndexes,
		//independentStatements: independentStatements,
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
			fmt.Println("name: ", name)
			body := findFunctionDefinitionBody(functionDefinitionNode, sourceCode)
			parameterNames := findFunctionParametersNames(functionDefinitionNode)
			retParameterNames := findFunctionRetParameterTypes(functionDefinitionNode)
			topLevelDeclarationsIndexes := findFunctionTopLevelDeclarationStatements(functionDefinitionNode)

			fi.functionDefinitions[name] = &FunctionDefinition{
				Name:                        name,
				Body:                        body,
				ParameterNames:              parameterNames,
				RetParameterTypes:           retParameterNames,
				TopLevelDeclarationsIndexes: topLevelDeclarationsIndexes,
			}
		}
	}

	return fi.functionDefinitions
}

func storeAllFunctionDefinitionNodes(node interface{}, functionNodes []map[string]interface{}) []map[string]interface{} {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			functionNodes = storeAllFunctionDefinitionNodes(element, functionNodes)
		}
	case map[string]interface{}:
		nodeMap := node.(map[string]interface{})
		for key, value := range nodeMap {
			if key == "nodeType" && value == "FunctionDefinition" {
				functionNodes = append(functionNodes, nodeMap)
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
		retParametersTypesList = append(retParametersTypesList, retParameterTypeDesc["typeString"].(string))
	}

	return retParametersTypesList
}

func findFunctionTopLevelDeclarationStatements(functionDefinitionNodeMap map[string]interface{}) []int {
	bodyField := functionDefinitionNodeMap["body"].(map[string]interface{})
	statementsList := bodyField["statements"].([]interface{})
	ret := make([]int, 0)

	for _, statementInterface := range statementsList {
		statementMap := statementInterface.(map[string]interface{})
		if statementMap["nodeType"].(string) == "VariableDeclarationStatement" {
			statementSrc := statementMap["src"].(string)
			statementSrcParts := strings.Split(statementSrc, ":")
			statementStart, _ := strconv.Atoi(statementSrcParts[0])
			adjustedStatementStart := statementStart
			ret = append(ret, adjustedStatementStart)
		}
	}

	return ret

}
