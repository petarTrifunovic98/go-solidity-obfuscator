package obfuscator

import (
	"regexp"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/contractprovider"
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
	blockStartsIndex := 0

	for i := 0; i < len(blockEnds); i++ {
		sourceCodeString = sourceCodeString[:blockStarts[blockStartsIndex][0]-stringReduction] + sourceCodeString[blockEnds[i][1]-stringReduction:]
		stringReduction += blockEnds[i][1] - blockStarts[blockStartsIndex][0]
		for blockStartsIndex < len(blockStarts) && blockStarts[blockStartsIndex][0] <= blockEnds[i][1] {
			blockStartsIndex++
		}
	}

	contract.SetSourceCode(sourceCodeString)

	return sourceCodeString
}
