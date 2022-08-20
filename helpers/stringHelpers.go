package helpers

import (
	"io/ioutil"
	"os"
	"strings"
)

func ReadFileToString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return "", err
	}

	byteValue, _ := ioutil.ReadAll(file)
	fileString := string(byteValue)

	return fileString, nil
}

func CopyString(original string) (string, error) {
	var sb strings.Builder
	if _, err := sb.WriteString(original); err != nil {
		return "", err
	}

	return sb.String(), nil
}
