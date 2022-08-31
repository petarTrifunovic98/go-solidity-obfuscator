package main

import (
	"fmt"
	"math/rand"
	"regexp"
	contractprovider "solidity-obfuscator/contractProvider"
	"solidity-obfuscator/helpers"
	processinformation "solidity-obfuscator/processInformation"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FunctionDefinition struct {
	body                  string
	parameterNames        []string
	retParameterTypes     []string
	independentStatements []string
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

func findIndependentStatements(functionBody string) []string {
	bodyCopy, _ := helpers.CopyString(functionBody)

	statements := make([]string, 0)

	stmtStart := 0
	parenthesesCounter := 0

	for ind, character := range bodyCopy {
		if character == ';' && parenthesesCounter == 0 {
			statements = append(statements, bodyCopy[stmtStart:ind+1])
			stmtStart = ind + 1
		} else if character == '{' {
			parenthesesCounter++
		} else if character == '}' {
			parenthesesCounter--
			if parenthesesCounter == 0 {
				statements = append(statements, bodyCopy[stmtStart:ind+1])
				stmtStart = ind + 1
			}
		}
	}

	return statements

	// return nil
}

func extractFunctionDefinition(node interface{}, functionName string, sourceString string) *FunctionDefinition {
	functionDefinitionNode := findFunctionDefinitionNode(node, functionName)
	if functionDefinitionNode == nil {
		return nil
	}

	body := findFunctionDefinitionBody(functionDefinitionNode, sourceString)
	parametersNames := findFunctionParametersNames(functionDefinitionNode)
	retParametersNames := findFunctionRetParameterTypes(functionDefinitionNode)
	independentStatements := findIndependentStatements(body)

	functionDefinition := FunctionDefinition{
		body:                  body,
		parameterNames:        parametersNames,
		retParameterTypes:     retParametersNames,
		independentStatements: independentStatements,
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

func (mf *ManipulatedFunction) insertOpaquePredicates(uselessArrayNames [2]string) {
	independentStatements := findIndependentStatements(mf.body)
	independentStatementsLen := len(independentStatements)
	if independentStatementsLen < 2 {
		return
	}

	randomSeeder := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSeeder)

	var statementsSplitIndex int

	if independentStatementsLen == 2 {
		statementsSplitIndex = 2
	} else {

		statementsSplitIndex = randomGenerator.Intn(independentStatementsLen) + 1
	}

	//replace "7" with a declared constant
	arraySize := randomGenerator.Intn(7) + 1
	firstArrayDeclaration := "uint" + "[" + strconv.Itoa(arraySize) + "] "
	lenToCopy := len(firstArrayDeclaration)
	firstArrayDeclaration += uselessArrayNames[0] + " = [uint("

	//replace "20" with a declared constant
	firstArrayDeclaration += strconv.Itoa(randomGenerator.Intn(20)) + ")"

	for i := 1; i < arraySize; i++ {
		firstArrayDeclaration += ", " + strconv.Itoa(randomGenerator.Intn(20))
	}
	firstArrayDeclaration += "];\n"

	secondArrayDeclaration := firstArrayDeclaration[:lenToCopy] + uselessArrayNames[1] + " = " + uselessArrayNames[0] + ";\n"

	randomIndex := randomGenerator.Intn(arraySize)
	ifStmt := "if (" + uselessArrayNames[0] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
	for i := 0; i < statementsSplitIndex; i++ {
		ifStmt += independentStatements[i]
	}

	if statementsSplitIndex < independentStatementsLen {
		ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
		for i := statementsSplitIndex; i < independentStatementsLen; i++ {
			ifStmt += independentStatements[i]
		}
	}

	body := firstArrayDeclaration + secondArrayDeclaration + ifStmt
	fmt.Println(body)

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

	//stringIndexIncrease := 0

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

			var arrNames [2]string
			newVarName := variableInfo.GetLatestDashVariableName() + "_"
			for i := 0; i < 2; i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				arrNames[i] = newVarName
			}
			newVarName += "_"

			manipulatedFunc.insertOpaquePredicates(arrNames)

			manipulatedFunc.replaceFunctionParametersWithArguments(functionDef.parameterNames, functionCall.args)
			retVarNames := make([]string, len(functionDef.retParameterTypes))

			for i := 0; i < len(functionDef.retParameterTypes); i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				retVarNames[i] = newVarName
			}

			manipulatedFunc.replaceReturnStmtWithVariables(retVarNames, functionDef.retParameterTypes)

			funcCallStart := functionCall.indexInSource
			funcCallEnd := functionCall.indexInSource + functionCall.callLen

			numToAdd := sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)
			// newSourceCodeIndex := funcCallStart + numToAdd

			// i := funcCallStart + stringIndexIncrease
			i := funcCallStart + numToAdd

			// fmt.Print("i: ")
			// fmt.Print(i)
			// fmt.Print("; RBTreeWithDLList calculated i: ")
			// fmt.Println(newSourceCodeIndex)
			for sourceCodeString[i] != ';' && sourceCodeString[i] != '{' && sourceCodeString[i] != '}' {
				i--
			}
			sourceCodeString = sourceCodeString[:i+1] + manipulatedFunc.body + sourceCodeString[i+1:]
			//stringIndexIncrease += len(manipulatedFunc.body)
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
			// newSourceCodeIndex = funcCallStart + numToAdd
			// fmt.Print("i: ")
			// fmt.Print(funcCallStart + stringIndexIncrease)
			// fmt.Print("; RBTreeWithDLList calculated i: ")
			// fmt.Println(newSourceCodeIndex)

			sourceCodeString = sourceCodeString[:funcCallStart+numToAdd] + insertString + sourceCodeString[funcCallEnd+numToAdd:]
			stringLenDiff := len(insertString) - functionCall.callLen
			smallerStringLen := functionCall.callLen
			if stringLenDiff < 0 {
				smallerStringLen = len(insertString)
				// fmt.Println("Smaller")
			} /*else {
				fmt.Println("Bigger")
			}*/
			if stringLenDiff != 0 {
				sourceCodeChangeInfo.ReportSourceCodeChange(funcCallStart+numToAdd+smallerStringLen, stringLenDiff)
			}
			//stringIndexIncrease += len(insertString) - functionCall.callLen

			variableInfo.SetLatestDashVariableName(newVarName)
		}
	}

	sourceCodeChangeInfo.DisplayTree()

	contract.SetSourceCode(sourceCodeString)

	return sourceCodeString

}
