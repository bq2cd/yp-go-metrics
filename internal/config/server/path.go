package server

import (
	"fmt"
	"os"
	"path/filepath"
)

func getAbsoluteFilePath(path string) (string, error) {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		return "", fmt.Errorf("dir is not allowed")
	}

	abspath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot determine absolute path: %w", err)
	}

	return abspath, nil
}
