package consts

import (
	"fmt"
	"os"
	"path/filepath"
)

func StateDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home dir failed: %w", err)
	}

	return filepath.Join(homeDir, ".local", "state", AppName), nil
}

func EnsureStateDir() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create state dir failed: %w", err)
	}

	return dir, nil
}

func PIDFilePath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, PIDFileName), nil
}

func LogFilePath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, LogFileName), nil
}

func StateFilePath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, StateFileName), nil
}
