package utils

import (
	"os"
	"path/filepath"
)

func LoadJSONFromFile(filePath string) (string, error) {
	jsonBytes, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}