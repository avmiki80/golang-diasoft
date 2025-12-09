package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrEmptyDir = errors.New("dir is empty")

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

func NewEnvValue(dir string, filename string) (*EnvValue, error) {
	path := filepath.Join(dir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info %s: %w", path, err)
	}

	if fileInfo.Size() == 0 {
		return &EnvValue{Value: "", NeedRemove: true}, nil
	}

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()

		line = strings.TrimRight(line, " \t")

		line = strings.ReplaceAll(line, "\x00", "\n")

		if line == "" {
			return &EnvValue{Value: "", NeedRemove: true}, nil
		}

		return &EnvValue{Value: line, NeedRemove: false}, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	return &EnvValue{Value: "", NeedRemove: true}, nil
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	if dir == "" {
		return nil, ErrEmptyDir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	environments := make(Environment)
	var processingErrors []error

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.Contains(entry.Name(), "=") {
			continue
		}

		env, err := NewEnvValue(dir, entry.Name())
		if err != nil {
			processingErrors = append(processingErrors, err)
			continue
		}

		if env != nil {
			environments[entry.Name()] = *env
		}
	}

	if len(processingErrors) > 0 {
		return environments, fmt.Errorf("errors processing some files: %v", processingErrors)
	}

	return environments, nil
}
