package contractprovider

import (
	"fmt"
	"strings"
	"sync"

	"github.com/petarTrifunovic98/go-solidity-obfuscator/pkg/helpers"
)

type solidityContract struct {
	originalSourceCode string
	sourceCode         string
	jsonCompactAST     map[string]interface{}
}

var once sync.Once
var instance *solidityContract

func SolidityContractInstance() *solidityContract {
	once.Do(func() {

		configFile, errConfig := helpers.ReadJsonToMap("./config/config.json")
		if errConfig != nil {
			fmt.Println(errConfig)
			instance = nil
			return
		}

		sourceCodeString, errSource := helpers.ReadFileToString(configFile["sourceCodePath"].(string))
		if errSource != nil {
			fmt.Println(errSource)
			instance = nil
			return
		}

		jsonCompactASTString, errJson := helpers.ReadJsonToMap(configFile["jsonCompactASTPath"].(string))
		if errJson != nil {
			fmt.Println(errJson)
			instance = nil
			return
		}

		instance = &solidityContract{
			originalSourceCode: sourceCodeString,
			sourceCode:         sourceCodeString,
			jsonCompactAST:     jsonCompactASTString,
		}
	})

	return instance
}

func (s *solidityContract) GetOriginalSourceCode() string {
	var sb strings.Builder
	if _, err := sb.WriteString(s.originalSourceCode); err != nil {
		fmt.Println("error copying string!")
		fmt.Println(err)
		return ""
	}

	return sb.String()
}

func (s *solidityContract) GetSourceCode() string {
	var sb strings.Builder
	if _, err := sb.WriteString(s.sourceCode); err != nil {
		fmt.Println("error copying string!")
		fmt.Println(err)
		return ""
	}

	return sb.String()
}

func (s *solidityContract) GetJsonCompactAST() map[string]interface{} {
	return s.jsonCompactAST
}

func (s *solidityContract) SetSourceCode(newSourceCode string) {
	//possibly add mutex here, and waitgroups for getters, if this turns into a concurrent program
	s.sourceCode = newSourceCode
}
