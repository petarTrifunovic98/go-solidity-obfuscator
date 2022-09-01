package processinformation

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type functionBody struct {
	bodyContent   string
	indexInSource int
}

type FunctionDefinition struct {
	Body                        functionBody
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
	functionDefitions []*FunctionDefinition
	functionCalls     []*FunctionCall
}

var functionInfoOnce sync.Once
var functionInfoInstance *functionInformation

func FunctionInformation() *functionInformation {
	functionInfoOnce.Do(func() {
		functionInfoInstance = &functionInformation{
			functionDefitions: nil,
			functionCalls:     nil,
		}
	})

	return functionInfoInstance
}

func (fi *functionInformation) GetFunctionCalls() []*FunctionCall {
	return fi.functionCalls
}

func (fi *functionInformation) GetFunctionDefinitions() []*FunctionDefinition {
	return fi.functionDefitions
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
				argsList := fi.findFunctionCallArgumentValues(nodeMap, sourceString)
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

func (fi *functionInformation) findFunctionCallArgumentValues(functionCallNodeMap map[string]interface{}, sourceString string) []string {
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
