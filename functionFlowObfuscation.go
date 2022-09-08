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

// type FunctionBody struct {
// 	bodyContent   string
// 	indexInSource int
// }

// type FunctionDefinition struct {
// 	body                        FunctionBody
// 	parameterNames              []string
// 	retParameterTypes           []string
// 	independentStatements       []string
// 	topLevelDeclarationsIndexes []int
// }

// type FunctionCall struct {
// 	name          string
// 	args          []string
// 	indexInSource int
// 	callLen       int
// }

// type ManipulatedFunction struct {
// 	body FunctionBody
// }

func checkElseStmtFollows(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) < 5 {
		return false
	}
	codeToTest := code[0:5]
	matched, _ := regexp.MatchString(`else\s`, codeToTest)
	return matched
}

func findIndependentStatements(functionBody string) map[int]string {
	referentBodyCopy, _ := helpers.CopyString(functionBody)
	bodyCopy, _ := helpers.CopyString(functionBody)
	fmt.Println(bodyCopy)

	statements := make(map[int]string, 0)

	stmtStart := 0
	parenthesesCounter := 0
	whitespaceDiff := 0

	for ind, character := range bodyCopy {
		if character == ';' && parenthesesCounter == 0 && !checkElseStmtFollows(bodyCopy[ind+1:]) {
			independentStmt := strings.TrimSpace(bodyCopy[stmtStart : ind+1])
			i := 0
			for referentBodyCopy[stmtStart+i] != independentStmt[0] {
				i++
			}
			whitespaceDiff = i
			statements[stmtStart+whitespaceDiff] = strings.TrimSpace(bodyCopy[stmtStart : ind+1])
			fmt.Print(stmtStart + whitespaceDiff)
			fmt.Print(": ")
			fmt.Println(statements[stmtStart+whitespaceDiff])
			stmtStart = ind + 1
		} else if character == '{' {
			parenthesesCounter++
		} else if character == '}' {
			parenthesesCounter--
			if parenthesesCounter == 0 && !checkElseStmtFollows(bodyCopy[ind+1:]) {
				independentStmt := strings.TrimSpace(bodyCopy[stmtStart : ind+1])
				i := 0
				for referentBodyCopy[stmtStart+i] != independentStmt[0] {
					i++
				}
				whitespaceDiff = i
				statements[stmtStart+whitespaceDiff] = strings.TrimSpace(bodyCopy[stmtStart : ind+1])
				fmt.Print(stmtStart + whitespaceDiff)
				fmt.Print(": ")
				fmt.Println(statements[stmtStart+whitespaceDiff])
				stmtStart = ind + 1
			}
		}
	}

	return statements
}

func replaceFunctionParametersWithArguments(functionBody string, functionParameters []string, functionArguments []string) string {
	newBody, _ := helpers.CopyString(functionBody)

	i := 0
	for _, parameter := range functionParameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		newBody = re.ReplaceAllString(newBody, functionArguments[i])
		i++
	}

	return newBody
}

func replaceReturnStmtWithVariables(functionBody string, retVarNames []string, retParameterTypes []string) string {
	newBody, _ := helpers.CopyString(functionBody)
	re, _ := regexp.Compile("\\breturn\\b")
	retStmtIndexes := re.FindAllStringIndex(newBody, -1)
	if retStmtIndexes == nil {
		return newBody
	}

	stringIncrease := 0
	fullPrependLen := 0
	var insertString string
	var prependString string

	for _, indexPair := range retStmtIndexes {
		retStmtStartIndex := indexPair[0]
		retStmtEndIndex := retStmtStartIndex
		for newBody[retStmtEndIndex] != ';' {
			retStmtEndIndex++
		}

		retValuesList := strings.Split(newBody[retStmtStartIndex+len("return")+stringIncrease:retStmtEndIndex+stringIncrease], ",;")

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

		newBody = prependString + newBody[:retStmtStartIndex+stringIncrease] + insertString + newBody[retStmtEndIndex+stringIncrease+1:]
		stringIncrease += len(insertString) + len(prependString)
		fullPrependLen += len(prependString)

	}

	newBody = newBody[:fullPrependLen] + "\n{\n" + newBody[fullPrependLen:] + "\n}\n"

	return newBody
}

func insertOpaquePredicates(functionBody string, bodyIndexInSource int, uselessArrayNames [2]string, topLevelDeclIndexes []int) string {
	newBody, _ := helpers.CopyString(functionBody)
	independentStatements := findIndependentStatements(newBody)

	sourceCodeChangeInfo := processinformation.SourceCodeChangeInformation()
	realBodyIndexInSource := bodyIndexInSource + sourceCodeChangeInfo.NumToAddToSearch(bodyIndexInSource)

	topLevelDeclarations := make([]string, 0)

	for _, topLevelDeclIndex := range topLevelDeclIndexes {
		realTopLevelDeclIndex := topLevelDeclIndex + sourceCodeChangeInfo.NumToAddToSearch(topLevelDeclIndex)
		declIndexInBody := realTopLevelDeclIndex - realBodyIndexInSource
		i := declIndexInBody
		for newBody[i] != ';' {
			i++
		}
		fmt.Println("###############")
		fmt.Print(declIndexInBody)
		fmt.Print(": ")
		fmt.Println(newBody[declIndexInBody-1 : i+1])
		topLevelDeclarations = append(topLevelDeclarations, newBody[declIndexInBody-1:i+1])
		if stmt, ok := independentStatements[declIndexInBody-1]; ok {
			delete(independentStatements, declIndexInBody-1)
			fmt.Println(stmt)
		} else {
			fmt.Println("Is not independent")
		}
		fmt.Println("###############")
	}

	fmt.Println("map: ", independentStatements)
	fmt.Println("array: ", topLevelDeclarations)

	independentStmtsKeys := make([]int, 0)
	for key := range independentStatements {
		independentStmtsKeys = append(independentStmtsKeys, key)
	}
	sort.Ints(independentStmtsKeys)

	independentStmtsList := make([]string, 0)
	for _, key := range independentStmtsKeys {
		independentStmtsList = append(independentStmtsList, independentStatements[key])
	}

	independentStatementsLen := len(independentStatements)

	if independentStatementsLen < 2 {
		return newBody
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

	var linkedDeclarations string
	for _, declaration := range topLevelDeclarations {
		linkedDeclarations += declaration + "\n"
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
		ifStmt += independentStmtsList[i]
	}
	ifStmt += "\n}\n"
	ifStmt += "else {"
	for i := 0; i < statementsSplitIndex2; i++ {
		ifStmt += independentStmtsList[i]
	}
	ifStmt += "\n}\n"

	if statementsSplitIndex1 < independentStatementsLen {
		ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
		for i := statementsSplitIndex1; i < independentStatementsLen; i++ {
			ifStmt += independentStmtsList[i]
		}
		ifStmt += "\n}\n"
	}

	if statementsSplitIndex2 < independentStatementsLen {
		if statementsSplitIndex1 < independentStatementsLen {
			ifStmt += "else {"
		} else {
			ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 != 0 {"
		}
		for i := statementsSplitIndex2; i < independentStatementsLen; i++ {
			ifStmt += independentStmtsList[i]
		}
		ifStmt += "\n}\n"
	}

	newBody = linkedDeclarations + firstArrayDeclaration + secondArrayDeclaration + ifStmt
	fmt.Println(newBody)
	fmt.Println("---------------------")

	return newBody
}

func ManipulateDefinedFunctionBodies() {

	// contract := contractprovider.SolidityContractInstance()
	// jsonAST := contract.GetJsonCompactAST()
	// sourceCodeString := contract.GetSourceCode()
	// functionInfo := processinformation.FunctionInformation()
	// functionDefinitions := functionInfo.ExtractAllFunctionDefinitions(jsonAST, sourceCodeString)

	// for _, functionDefinition := range functionDefinitions {
	// 	newBody, _ := functionDefinition.Body.BodyContent
	// }

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
			newBodyContent, _ := helpers.CopyString(functionDef.Body.BodyContent)
			newBodyIndex := functionDef.Body.IndexInSource
			// manipulatedFunc := ManipulatedFunction{
			// 	body: FunctionBody{
			// 		bodyContent:   manipulatedFuncBodyContent,
			// 		indexInSource: functionDef.Body.IndexInSource,
			// 	},
			// }

			var arrNames [2]string
			newVarName := variableInfo.GetLatestDashVariableName() + "_"
			for i := 0; i < 2; i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				arrNames[i] = newVarName
				newVarName += "_"
			}

			newBodyContent = insertOpaquePredicates(newBodyContent, newBodyIndex, arrNames, functionDef.TopLevelDeclarationsIndexes)
			newBodyContent = replaceFunctionParametersWithArguments(newBodyContent, functionDef.ParameterNames, functionCall.Args)
			retVarNames := make([]string, len(functionDef.RetParameterTypes))
			for i := 0; i < len(functionDef.RetParameterTypes); i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				retVarNames[i] = newVarName
			}
			newBodyContent = replaceReturnStmtWithVariables(newBodyContent, retVarNames, functionDef.RetParameterTypes)

			funcCallStart := functionCall.IndexInSource
			funcCallEnd := functionCall.IndexInSource + functionCall.CallLen
			numToAdd := sourceCodeChangeInfo.NumToAddToSearch(funcCallStart)
			i := funcCallStart + numToAdd
			for sourceCodeString[i] != ';' && sourceCodeString[i] != '{' && sourceCodeString[i] != '}' {
				i--
			}
			sourceCodeString = sourceCodeString[:i+1] + newBodyContent + sourceCodeString[i+1:]
			sourceCodeChangeInfo.ReportSourceCodeChange(i+1, len(newBodyContent))

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
