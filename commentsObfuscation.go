package main

import (
	"regexp"
)

func replaceComments(sourceString string) string {
	re, _ := regexp.Compile("//(.*)(\n)")
	sourceString = re.ReplaceAllString(sourceString, "\n")

	reStartBlock, _ := regexp.Compile("/\\*")
	reEndBlock, _ := regexp.Compile("\\*/")

	blockStart := reStartBlock.FindStringIndex(sourceString)
	blockEnd := reEndBlock.FindStringIndex(sourceString)

	for blockStart != nil && blockEnd != nil {
		sourceString = sourceString[:blockStart[0]] + sourceString[blockEnd[1]:]
		blockStart = reStartBlock.FindStringIndex(sourceString)
		blockEnd = reEndBlock.FindStringIndex(sourceString)
	}

	//#TODO - change to use FindAllStringIndex

	return sourceString
}
