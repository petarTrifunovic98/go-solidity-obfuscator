package main

import (
	"fmt"
	"regexp"
)

func replaceComments(sourceString string) string {
	re, _ := regexp.Compile("//(.*)(\n)")
	sourceString = re.ReplaceAllString(sourceString, "\n")

	reStartBlock, _ := regexp.Compile("/\\*")
	reEndBlock, _ := regexp.Compile("\\*/")

	blockStarts := reStartBlock.FindAllStringIndex(sourceString, -1)
	blockEnds := reEndBlock.FindAllStringIndex(sourceString, -1)

	fmt.Println(blockStarts)
	fmt.Println(blockEnds)

	stringReduction := 0

	for i := 0; i < len(blockStarts); i++ {
		sourceString = sourceString[:blockStarts[i][0]-stringReduction] + sourceString[blockEnds[i][1]-stringReduction:]
		stringReduction += blockEnds[i][1] - blockStarts[i][0]
	}

	return sourceString
}
