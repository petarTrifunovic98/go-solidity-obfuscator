package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/obfuscator"
)

func main() {

	outputPtr := flag.String("output", "./obfuscated.sol", "the path to the output file")
	fobfType := flag.String("fobf", "opaque", `The type of function `+
		`obfuscation to use. One of "opaque" (opaque predicates) and "inline" (inline `+
		`called functions). Throws an error in other cases.`)

	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Println("requires two positional arguments - the contract (.sol) and the contract AST, in that order")
		os.Exit(1)
	}

	if *fobfType != "opaque" && *fobfType != "inline" {
		fmt.Println(`fobf option must be equal to "opaque" or "inline"`)
		os.Exit(1)
	}

	if *fobfType == "opaque" {
		obfuscator.ManipulateDefinedFunctionBodies()
	} else {
		obfuscator.ManipulateCalledFunctionsBodies()
	}
	obfuscator.ReplaceVarNames()
	obfuscator.ReplaceComments()
	obfuscationResult := obfuscator.ReplaceLiterals()

	outputFile, errOutput := os.Create(*outputPtr)
	if errOutput != nil {
		fmt.Println(errOutput)
		return
	}
	defer outputFile.Close()

	outputFile.WriteString(obfuscationResult)
}
