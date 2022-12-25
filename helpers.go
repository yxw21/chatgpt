package chatgpt

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func GetUserHomeDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("error getting user home folder: " + err.Error())
	}
	separator := string(filepath.Separator)
	if !strings.HasSuffix(userHomeDir, separator) {
		userHomeDir += separator
	}
	return userHomeDir, nil
}

func GetTempDir() string {
	tempDir := os.TempDir()
	separator := string(filepath.Separator)
	if !strings.HasSuffix(tempDir, separator) {
		tempDir += separator
	}
	return tempDir
}

func ConvertMapToStruct(data map[string]any, result any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &result)
}
