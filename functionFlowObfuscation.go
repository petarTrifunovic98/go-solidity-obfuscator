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

type FunctionBody struct {
	bodyContent   string
	indexInSource int
}

type FunctionDefinition struct {
	body                        FunctionBody
	parameterNames              []string
	retParameterTypes           []string
	independentStatements       []string
	topLevelDeclarationsIndexes []int
}

type FunctionCall struct {
	name          string
	args          []string
	indexInSource int
	callLen       int
}

type ManipulatedFunction struct {
	body FunctionBody
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

func (mf *ManipulatedFunction) replaceFunctionParametersWithArguments(functionParameters []string, functionArguments []string) {
	body := mf.body

	i := 0
	for _, parameter := range functionParameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		body.bodyContent = re.ReplaceAllString(body.bodyContent, functionArguments[i])
		i++
	}

	mf.body = body
}

func (mf *ManipulatedFunction) replaceReturnStmtWithVariables(retVarNames []string, retParameterTypes []string) {
	bodyContent := mf.body.bodyContent
	re, _ := regexp.Compile("\\breturn\\b")
	retStmtIndexes := re.FindAllStringIndex(bodyContent, -1)
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
		for bodyContent[retStmtEndIndex] != ';' {
			retStmtEndIndex++
		}

		retValuesList := strings.Split(bodyContent[retStmtStartIndex+len("return")+stringIncrease:retStmtEndIndex+stringIncrease], ",;")

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

		bodyContent = prependString + bodyContent[:retStmtStartIndex+stringIncrease] + insertString + bodyContent[retStmtEndIndex+stringIncrease+1:]
		stringIncrease += len(insertString) + len(prependString)
		fullPrependLen += len(prependString)

	}

	bodyContent = bodyContent[:fullPrependLen] + "\n{\n" + bodyContent[fullPrependLen:] + "\n}\n"

	mf.body.bodyContent = bodyContent
}

func (mf *ManipulatedFunction) insertOpaquePredicates(uselessArrayNames [2]string, topLevelDeclIndexes []int) {
	independentStatements := findIndependentStatements(mf.body.bodyContent)
	independentStatementsLen := len(independentStatements)
	if independentStatementsLen < 2 {
		return
	}

	//extractedDeclarations := make([]string, 0)

	sourceCodeChangeInfo := processinformation.SourceCodeChangeInformation()
	realBodyIndexInSource := mf.body.indexInSource + sourceCodeChangeInfo.NumToAddToSearch(mf.body.indexInSource)

	for _, topLevelDeclIndex := range topLevelDeclIndexes {
		realTopLevelDeclIndex := topLevelDeclIndex + sourceCodeChangeInfo.NumToAddToSearch(topLevelDeclIndex)
		declIndexInBody := realTopLevelDeclIndex - realBodyIndexInSource
		i := declIndexInBody
		for mf.body.bodyContent[i] != ';' {
			i++
		}
		fmt.Println("###############")
		fmt.Println(mf.body.bodyContent[declIndexInBody-1 : i])
		fmt.Println("###############")
	}

	randomSeeder := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSeeder)

	var statementsSplitIndex1 int
	var statementsSplitIndex2 int

	if independentStatementsLen == 2 {
		statementsSplitIndex1 = 2
	} else {
		statementsSplitIndex1 = randomGenerator.Intn(independentStatementsLen) + 1
		statementsSplitIndex2 = randomGenerator.Intn(independentStatementsLen) + 1
		for statementsSplitIndex1 == statementsSplitIndex2 {
			statementsSplitIndex2 = randomGenerator.Intn(independentStatementsLen) + 1
		}
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
	for i := 0; i < statementsSplitIndex1; i++ {
		ifStmt += independentStatements[i]
	}
	ifStmt += "\n}\n"

	if statementsSplitIndex1 < independentStatementsLen {
		ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
		for i := statementsSplitIndex1; i < independentStatementsLen; i++ {
			ifStmt += independentStatements[i]
		}
		ifStmt += "\n}\n"
	}

	body := firstArrayDeclaration + secondArrayDeclaration + ifStmt
	fmt.Println(body)
	fmt.Println("---------------------")

}

func ManipulateDefinedFunctionBodies() {

	// contract := contractprovider.SolidityContractInstance()
	// jsonAST := contract.GetJsonCompactAST()
	// sourceCodeString := contract.GetSourceCode()
	// functionDefinitions := extractAllFunctionDefinitions(jsonAST, sourceCodeString)

}

func ManipulateCalledFunctionsBodies() string {

	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()
	functionInfo := processinformation.FunctionInformation()
	functionCalls := functionInfo.GetFunctionCalls()
	if functionCalls == nil {
		functionCalls = functionInfo.ExtractFunctionCalls(jsonAST, sourceCodeString)
	}

	sourceCodeChangeInfo := processinformation.SourceCodeChangeInformation()

	sort.Slice(functionCalls, func(i, j int) bool {
		return functionCalls[i].IndexInSource < functionCalls[j].IndexInSource
	})

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
		functionDef, exists := functionInfo.GetSpecificFunctionDefinition(functionCall.Name)
		if !exists {
			functionDef = functionInfo.ExtractFunctionDefinition(jsonAST, functionCall.Name, originalSourceString)
		}
		if functionDef != nil {
			manipulatedFuncBodyContent, _ := helpers.CopyString(functionDef.Body.BodyContent)
			manipulatedFunc := ManipulatedFunction{
				body: FunctionBody{
					bodyContent:   manipulatedFuncBodyContent,
					indexInSource: functionDef.Body.IndexInSource,
				},
			}

			var arrNames [2]string
			newVarName := variableInfo.GetLatestDashVariableName() + "_"
			for i := 0; i < 2; i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				arrNames[i] = newVarName
			}
			newVarName += "_"

			manipulatedFunc.insertOpaquePredicates(arrNames, functionDef.TopLevelDeclarationsIndexes)
			manipulatedFunc.replaceFunctionParametersWithArguments(functionDef.ParameterNames, functionCall.Args)
			retVarNames := make([]string, len(functionDef.RetParameterTypes))
			for i := 0; i < len(functionDef.RetParameterTypes); i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				retVarNames[i] = newVarName
			}
			manipulatedFunc.replaceReturnStmtWithVariables(retVarNames, functionDef.RetParameterTypes)

			funcCallStart := functionCall.IndexInSource
			funcCallEnd := functionCall.IndexInSource + functionCall.CallLen
			numToAdd := sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)
			i := funcCallStart + numToAdd
			for sourceCodeString[i] != ';' && sourceCodeString[i] != '{' && sourceCodeString[i] != '}' {
				i--
			}
			sourceCodeString = sourceCodeString[:i+1] + manipulatedFunc.body.bodyContent + sourceCodeString[i+1:]
			sourceCodeChangeInfo.ReportSourceCodeChange(i+1, len(manipulatedFunc.body.bodyContent))

			insertString := "("
			for ind, varName := range retVarNames {
				if ind > 0 {
					insertString += ", "
				}
				insertString += varName
			}
			insertString += ")"

			numToAdd = sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)

			sourceCodeString = sourceCodeString[:funcCallStart+numToAdd] + insertString + sourceCodeString[funcCallEnd+numToAdd:]
			stringLenDiff := len(insertString) - functionCall.CallLen
			smallerStringLen := functionCall.CallLen
			if stringLenDiff < 0 {
				smallerStringLen = len(insertString)
			}
			if stringLenDiff != 0 {
				sourceCodeChangeInfo.ReportSourceCodeChange(funcCallStart+numToAdd+smallerStringLen, stringLenDiff)
			}

			variableInfo.SetLatestDashVariableName(newVarName)
		}
	}

	sourceCodeChangeInfo.DisplayTree()
	contract.SetSourceCode(sourceCodeString)
	return sourceCodeString
}
