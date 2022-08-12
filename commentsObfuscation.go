package main

import (
	"regexp"
	contractprovider "solidity-obfuscator/contractProvider"
)

func ReplaceComments() string {
	contract := contractprovider.SolidityContractInstance()
	sourceCodeString := contract.GetSourceCode()

	re, _ := regexp.Compile("//(.*)(\n)")
	sourceCodeString = re.ReplaceAllString(sourceCodeString, "\n")

	reStartBlock, _ := regexp.Compile("/\\*")
	reEndBlock, _ := regexp.Compile("\\*/")

	blockStarts := reStartBlock.FindAllStringIndex(sourceCodeString, -1)
	blockEnds := reEndBlock.FindAllStringIndex(sourceCodeString, -1)

	stringReduction := 0

	for i := 0; i < len(blockStarts); i++ {
		sourceCodeString = sourceCodeString[:blockStarts[i][0]-stringReduction] + sourceCodeString[blockEnds[i][1]-stringReduction:]
		stringReduction += blockEnds[i][1] - blockStarts[i][0]
	}

	contract.SetSourceCode(sourceCodeString)

	return sourceCodeString
}
