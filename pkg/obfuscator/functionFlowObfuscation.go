package obfuscator

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/contractprovider"
	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/helpers"
	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/processinfo"
)

func replaceFunctionParametersWithArguments(functionBody string, sourceString string, functionParameters []string, /*functionArguments []string,*/
	functionArgs [][2]int) string {
	sourceCodeChangeInfo := processinfo.SourceCodeChangeInformation()

	argumentsList := make([]string, 0)

	for _, argsDetails := range functionArgs {
		reallArgIndex := argsDetails[0] + sourceCodeChangeInfo.NumToAddToSearch(argsDetails[0])
		argString := sourceString[reallArgIndex : reallArgIndex+argsDetails[1]]
		argumentsList = append(argumentsList, argString)
	}

	newBody, _ := helpers.CopyString(functionBody)

	i := 0
	for _, parameter := range functionParameters {
		re, _ := regexp.Compile("\\b" + parameter + "\\b")
		newBody = re.ReplaceAllString(newBody, argumentsList[i])
		i++
	}

	return newBody
}

func replaceReturnStmtWithVariables(functionBody string, retVarNames []string, retParameterTypes []string) string {
	newBody, _ := helpers.CopyString(functionBody)
	re, _ := regexp.Compile(`\breturn\b`)
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
		//fmt.Println("Done ret value split: ", retValuesList)

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

func insertOpaquePredicates(functionBody string, bodyIndexInSource int, uselessArrayNames [2]string, topLevelDecls [][2]int, independentStmts [][2]int) string {
	newBody, _ := helpers.CopyString(functionBody)

	sourceCodeChangeInfo := processinfo.SourceCodeChangeInformation()
	realBodyIndexInSource := bodyIndexInSource + sourceCodeChangeInfo.NumToAddToSearch(bodyIndexInSource)

	topLevelDeclarations := make([]string, 0)
	independentStatements := make([]string, 0)

	for _, topLevelDeclParameters := range topLevelDecls {
		realTopLevelDeclIndex := topLevelDeclParameters[0] + sourceCodeChangeInfo.NumToAddToSearch(topLevelDeclParameters[0])
		declIndexInBody := realTopLevelDeclIndex - realBodyIndexInSource
		topLevelDeclString := functionBody[declIndexInBody : declIndexInBody+topLevelDeclParameters[1]]
		if functionBody[declIndexInBody+topLevelDeclParameters[1]] == ';' {
			topLevelDeclString += ";"
		}
		topLevelDeclarations = append(topLevelDeclarations, topLevelDeclString)
	}

	for _, independentStmtParameters := range independentStmts {
		realIndependentStmtIndex := independentStmtParameters[0] + sourceCodeChangeInfo.NumToAddToSearch(independentStmtParameters[0])
		stmtIndexInBody := realIndependentStmtIndex - realBodyIndexInSource
		independentStmtString := functionBody[stmtIndexInBody : stmtIndexInBody+independentStmtParameters[1]]
		if functionBody[stmtIndexInBody+independentStmtParameters[1]] == ';' {
			independentStmtString += ";"
		}
		independentStatements = append(independentStatements, independentStmtString)
	}

	independentStatementsLen := len(independentStatements)

	if independentStatementsLen < 2 {
		return newBody
	}

	fmt.Println("Indeps: ", independentStatements)
	fmt.Println("Decls: ", topLevelDeclarations)

	randomSeeder := rand.NewSource(time.Now().UnixNano())
	randomGenerator := rand.New(randomSeeder)

	var statementsSplitIndex1 int
	var statementsSplitIndex2 int

	statementsSplitIndex1 = randomGenerator.Intn(independentStatementsLen) + 1
	statementsSplitIndex2 = randomGenerator.Intn(independentStatementsLen) + 1
	for statementsSplitIndex1 == statementsSplitIndex2 {
		statementsSplitIndex2 = randomGenerator.Intn(independentStatementsLen) + 1
	}

	var linkedDeclarations string
	for _, declaration := range topLevelDeclarations {
		linkedDeclarations += declaration + "\n"
	}

	//REMOVE
	//linkedDeclarations = ""

	//replace "7" with a declared constant
	arraySize := randomGenerator.Intn(7) + 1
	firstArrayDeclaration := "uint" + "[" + strconv.Itoa(arraySize) + "] memory "
	lenToCopy := len(firstArrayDeclaration)
	firstArrayDeclaration += uselessArrayNames[0] + " = [uint("

	//replace "20" with a declared constant
	firstArrayDeclaration += strconv.Itoa(randomGenerator.Intn(20)) + ")"

	for i := 1; i < arraySize; i++ {
		firstArrayDeclaration += ", " + strconv.Itoa(randomGenerator.Intn(20))
	}
	firstArrayDeclaration += "];\n"

	secondArrayDeclaration := firstArrayDeclaration[:lenToCopy] + uselessArrayNames[1] + " = " + uselessArrayNames[0] + ";\n"

	// randomIndex := randomGenerator.Intn(arraySize)
	randomIndexStr := "uint(uint160(msg.sender)) % " + strconv.Itoa(arraySize)
	// ifStmt := "if (" + uselessArrayNames[0] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
	ifStmt := "if (" + uselessArrayNames[0] + "[" + randomIndexStr + "] % 2 == 0) {"
	for i := 0; i < statementsSplitIndex1; i++ {
		ifStmt += independentStatements[i]
	}
	ifStmt += "\n}\n"
	ifStmt += "else {"
	for i := 0; i < statementsSplitIndex2; i++ {
		ifStmt += independentStatements[i]
	}
	ifStmt += "\n}\n"

	if statementsSplitIndex1 < independentStatementsLen {
		// ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 == 0) {"
		ifStmt += "if (" + uselessArrayNames[1] + "[" + randomIndexStr + "] % 2 == 0) {"
		for i := statementsSplitIndex1; i < independentStatementsLen; i++ {
			ifStmt += independentStatements[i]
		}
		ifStmt += "\n}\n"
	}

	if statementsSplitIndex2 < independentStatementsLen {
		if statementsSplitIndex1 < independentStatementsLen {
			ifStmt += "else {"
		} else {
			// ifStmt += "if (" + uselessArrayNames[1] + "[" + strconv.Itoa(randomIndex) + "] % 2 != 0) {"
			ifStmt += "if (" + uselessArrayNames[1] + "[" + randomIndexStr + "] % 2 != 0) {"
		}
		for i := statementsSplitIndex2; i < independentStatementsLen; i++ {
			ifStmt += independentStatements[i]
		}
		ifStmt += "\n}\n"
	}

	newBody = linkedDeclarations + firstArrayDeclaration + secondArrayDeclaration + ifStmt

	return newBody
}

func ManipulateDefinedFunctionBodies() string {
	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()
	functionInfo := processinfo.FunctionInformation()
	functionDefinitions := functionInfo.ExtractAllFunctionDefinitions(jsonAST, sourceCodeString)
	sourceCodeChangeInfo := processinfo.SourceCodeChangeInformation()

	variableInfo := processinfo.VariableInformation()
	namesSet := variableInfo.GetVariableNamesSet()
	if namesSet == nil {
		namesSet = getVarNames(jsonAST) //move to another place from VariableNameObfuscation.go
		variableInfo.SetVariableNamesSet(namesSet)
	}

	var newVarName string

	for _, functionDefinition := range functionDefinitions {
		newBodyContent, _ := helpers.CopyString(functionDefinition.Body.BodyContent)
		newBodyIndex := functionDefinition.Body.IndexInSource

		var arrNames [2]string
		newVarName = variableInfo.GetLatestDashVariableName() + "_"
		for i := 0; i < 2; i++ {
			for variableInfo.NameIsUsed(newVarName) {
				newVarName += "_"
			}
			arrNames[i] = newVarName
			newVarName += "_"
		}

		newBodyContent = insertOpaquePredicates(newBodyContent, newBodyIndex, arrNames, functionDefinition.TopLevelDeclarations, functionDefinition.IndependentStatements)
		fmt.Println("newBody:", newBodyContent)
		fmt.Println("oldBody:", functionDefinition.Body.BodyContent)
		numToAdd := sourceCodeChangeInfo.NumToAddToSearch(newBodyIndex)
		fmt.Println("numToAdd: ", numToAdd)
		fmt.Println(functionDefinition.Name)
		fmt.Println(newBodyIndex)
		fmt.Println("oldScLen: ", len(sourceCodeString))
		secondSourceCodeStringPart := sourceCodeString[newBodyIndex+numToAdd+len(functionDefinition.Body.BodyContent):]
		sourceCodeString = sourceCodeString[:newBodyIndex+numToAdd] + newBodyContent + secondSourceCodeStringPart
		fmt.Println("first intermediate sclen: ", len(sourceCodeString))
		fmt.Println("second start index: ", newBodyIndex+numToAdd+len(functionDefinition.Body.BodyContent))
		fmt.Println("second adition sclen: ", len(sourceCodeString[newBodyIndex+numToAdd+len(functionDefinition.Body.BodyContent):]))
		fmt.Println("first intermediate sclen: ", len(sourceCodeString))
		fmt.Println("scLen: ", len(sourceCodeString))

		fmt.Println("newBodyLen: ", len(newBodyContent))
		fmt.Println("oldBodyLen: ", len(functionDefinition.Body.BodyContent))
		stringLenDiff := len(newBodyContent) - len(functionDefinition.Body.BodyContent)
		smallerStringLen := len(functionDefinition.Body.BodyContent)
		if stringLenDiff < 0 {
			smallerStringLen = len(newBodyContent)
		}
		if stringLenDiff != 0 {
			fmt.Println("inserting")
			sourceCodeChangeInfo.ReportSourceCodeChange(newBodyIndex+numToAdd+1+smallerStringLen, stringLenDiff)
		}
		functionDefinition.Body.BodyContent = newBodyContent

		sourceCodeChangeInfo.DisplayTree()

		fmt.Println("-------------------")
	}

	variableInfo.SetLatestDashVariableName(newVarName)
	contract.SetSourceCode(sourceCodeString)
	return sourceCodeString
}

func ManipulateCalledFunctionsBodies() string {

	contract := contractprovider.SolidityContractInstance()
	jsonAST := contract.GetJsonCompactAST()
	sourceCodeString := contract.GetSourceCode()
	functionInfo := processinfo.FunctionInformation()
	functionCalls := functionInfo.GetFunctionCalls()
	if functionCalls == nil {
		functionCalls = functionInfo.ExtractFunctionCalls(jsonAST, sourceCodeString)
	}

	sourceCodeChangeInfo := processinfo.SourceCodeChangeInformation()

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

	variableInfo := processinfo.VariableInformation()
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

			var arrNames [2]string
			newVarName := variableInfo.GetLatestDashVariableName() + "_"
			for i := 0; i < 2; i++ {
				for variableInfo.NameIsUsed(newVarName) {
					newVarName += "_"
				}
				arrNames[i] = newVarName
				newVarName += "_"
			}

			//newBodyContent = insertOpaquePredicates(newBodyContent, newBodyIndex, arrNames, functionDef.TopLevelDeclarationsIndexes)
			newBodyContent = replaceFunctionParametersWithArguments(newBodyContent, sourceCodeString, functionDef.ParameterNames /*functionCall.ArgsOld,*/, functionCall.Args)
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
