package helpers

import (
	"io/ioutil"
	"os"
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
