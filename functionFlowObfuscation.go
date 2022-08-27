package main

import (
	"fmt"
	"regexp"
	contractprovider "solidity-obfuscator/contractProvider"
	"solidity-obfuscator/helpers"
	processinformation "solidity-obfuscator/processInformation"
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

type ManipulatedFunction struct {
	body string
}

func getFunctionCalls(jsonAST map[string]interface{}, sourceString string) []*FunctionCall {

	nodes := jsonAST["nodes"]
	functionsCalls := make([]*FunctionCall, 0)
	functionsCalls = storeFunctionCalls(nodes, functionsCalls, sourceString)
	return functionsCalls
}

func storeFunctionCalls(node interface{}, functionsCalls []*FunctionCall, sourceString string) []*FunctionCall {
	switch node.(type) {
	case []interface{}:
		nodeArr := node.([]interface{})
		for _, element := range nodeArr {
			functionsCalls = storeFunctionCalls(element, functionsCalls, sourceString)
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
					functionsCalls = storeFunctionCalls(value, functionsCalls, sourceString)
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

func (mf *ManipulatedFunction) replaceFunctionParametersWithArguments(functionParameters []string, functionArguments []string) {
	body := mf.body

	i := 0
	for _, parameter := range functionParameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		body = re.ReplaceAllString(body, functionArguments[i])
		i++
	}

	mf.body = body
}

func (mf *ManipulatedFunction) replaceReturnStmtWithVariables(retVarNames []string, retParameterTypes []string) {
	body := mf.body
	re, _ := regexp.Compile("\\breturn\\b")
	retStmtIndexes := re.FindAllStringIndex(body, -1)
	if retStmtIndexes == nil {
		return
	}

	stringIncrease := 0
	fullPrependLen := 0
	var insertString string
	var prependString string

	for _, indexPair := range retStmtIndexes {
		retStmtStartIndex := indexPair[0]
		retStmtEndIndex := retStmtStartIndex
		for body[retStmtEndIndex] != ';' {
			retStmtEndIndex++
		}

		retValuesList := strings.Split(body[retStmtStartIndex+len("return")+stringIncrease:retStmtEndIndex+stringIncrease], ",;")

		if len(retValuesList) > 0 {
			insertString = "\n{\n"
			prependString = "\n"
			for i := 0; i < len(retValuesList); i++ {
				retValue := strings.Trim(retValuesList[i], " \t\n")

				retVarDeclaration := retParameterTypes[i] + " " + retVarNames[i] + ";\n"
				retVarValueAssignment := retVarNames[i] + " = " + retValue + ";\n"

				prependString += retVarDeclaration
				insertString += retVarValueAssignment
			}
			insertString += "}\n"
		}

		body = prependString + body[:retStmtStartIndex+stringIncrease] + insertString + body[retStmtEndIndex+stringIncrease+1:]
		stringIncrease += len(insertString) + len(prependString)
		fullPrependLen += len(prependString)

	}

	body = body[:fullPrependLen] + "\n{\n" + body[fullPrependLen:] + "\n}\n"

	mf.body = body
}

func ManipulateCalledFunctionsBodies() string {

	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()
	functionCalls := getFunctionCalls(jsonAST, sourceCodeString)

	sourceCodeChangeInfo := processinformation.SourceCodeChangeInformation()

	nodes := jsonAST["nodes"]

	sort.Slice(functionCalls, func(i, j int) bool {
		return functionCalls[i].indexInSource < functionCalls[j].indexInSource
	})

	stringIndexIncrease := 0

	var sb strings.Builder
	if _, err := sb.WriteString(sourceCodeString); err != nil {
		fmt.Println("error copying string!")
		fmt.Println(err)
		return ""
	}

	originalSourceString := sb.String()

	variableInfo := processinformation.VariableInformation()
	namesSet := variableInfo.GetVariableNamesSet()
	if namesSet == nil {
		namesSet = getVarNames(jsonAST) //move to another place from VariableNameObfuscation.go
		variableInfo.SetVariableNamesSet(namesSet)
	}

	for _, functionCall := range functionCalls {
		//functionDef, exists := extractedFunctionDefinitions[functionCall.name]
		//if !exists {
		functionDef := extractFunctionDefinition(nodes, functionCall.name, originalSourceString)
		//}
		if functionDef != nil {

			manipulatedFunc := ManipulatedFunction{}
			manipulatedFunc.body, _ = helpers.CopyString(functionDef.body)

			manipulatedFunc.replaceFunctionParametersWithArguments(functionDef.parameterNames, functionCall.args)
			retVarNames := make([]string, len(functionDef.retParameterTypes))

			newVarName := variableInfo.GetLatestDashVariableName() + "_"

			for i := 0; i < len(functionDef.retParameterTypes); i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				retVarNames[i] = newVarName
			}
			newVarName += "_"

			manipulatedFunc.replaceReturnStmtWithVariables(retVarNames, functionDef.retParameterTypes)

			funcCallStart := functionCall.indexInSource
			funcCallEnd := functionCall.indexInSource + functionCall.callLen

			numToAdd := sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)
			newSourceCodeIndex := funcCallStart + numToAdd

			i := funcCallStart + stringIndexIncrease

			fmt.Print("i: ")
			fmt.Print(i)
			fmt.Print("; RBTreeWithDLList calculated i: ")
			fmt.Println(newSourceCodeIndex)
			for sourceCodeString[i] != ';' && sourceCodeString[i] != '{' && sourceCodeString[i] != '}' {
				i--
			}
			sourceCodeString = sourceCodeString[:i+1] + manipulatedFunc.body + sourceCodeString[i+1:]
			stringIndexIncrease += len(manipulatedFunc.body)
			sourceCodeChangeInfo.ReportSourceCodeChange(i+1, len(manipulatedFunc.body))

			insertString := "("
			for ind, varName := range retVarNames {
				if ind > 0 {
					insertString += ", "
				}
				insertString += varName
			}
			insertString += ")"

			numToAdd = sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)
			newSourceCodeIndex = funcCallStart + numToAdd
			fmt.Print("i: ")
			fmt.Print(funcCallStart + stringIndexIncrease)
			fmt.Print("; RBTreeWithDLList calculated i: ")
			fmt.Println(newSourceCodeIndex)

			sourceCodeString = sourceCodeString[:funcCallStart+stringIndexIncrease] + insertString + sourceCodeString[funcCallEnd+stringIndexIncrease:]
			stringLenDiff := len(insertString) - functionCall.callLen
			smallerStringLen := functionCall.callLen
			if stringLenDiff < 0 {
				smallerStringLen = len(insertString)
				fmt.Println("Smaller")
			} else {
				fmt.Println("Bigger")
			}
			if stringLenDiff != 0 {
				sourceCodeChangeInfo.ReportSourceCodeChange(funcCallStart+stringIndexIncrease+smallerStringLen, stringLenDiff)
			}
			stringIndexIncrease += len(insertString) - functionCall.callLen

			variableInfo.SetLatestDashVariableName(newVarName)
		}
	}

	sourceCodeChangeInfo.DisplayTree()

	contract.SetSourceCode(sourceCodeString)

	return sourceCodeString

}
