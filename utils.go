package sklib

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func insertAll(from map[string]float64, to map[string]float64) {
	for k, v := range from {
		to[k] = v
	}
}

func formatKey(key string) string {
	return apiKeyTag + key
}

func ReadFromFile(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadKey(fileName string) string {
	data, err := ReadFromFile(fileName)
	if err != nil {
		panic(fmt.Errorf("Could not read key from %s", fileName))
	}
	return strings.TrimSpace(data)
}
