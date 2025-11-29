package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir_WithTestdata(t *testing.T) {
	// Получаем абсолютный путь к testdata/env
	testDataPath := filepath.Join("testdata", "env")

	// Проверяем что директория существует
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Fatalf("Test data directory does not exist: %s", testDataPath)
	}

	env, err := ReadDir(testDataPath)
	if err != nil {
		t.Fatalf("ReadDir() failed with testdata: %v", err)
	}

	// Проверяем что все ожидаемые переменные присутствуют
	expectedVars := []string{"BAR", "FOO", "HELLO", "EMPTY", "UNSET"}
	for _, varName := range expectedVars {
		if _, exists := env[varName]; !exists {
			t.Errorf("Expected variable %s not found in result", varName)
		}
	}

	// Проверяем конкретные значения
	tests := []struct {
		name        string
		expected    EnvValue
		description string
	}{
		{
			name: "HELLO",
			expected: EnvValue{
				Value:      "\"hello\"",
				NeedRemove: false,
			},
			description: "simple value",
		},
		{
			name: "BAR",
			expected: EnvValue{
				Value:      "bar",
				NeedRemove: false,
			},
			description: "another simple value",
		},
		{
			name: "FOO",
			expected: EnvValue{
				Value:      "   foo\nwith new line",
				NeedRemove: false,
			},
			description: "value with spaces and newline",
		},
		{
			name: "EMPTY",
			expected: EnvValue{
				Value:      "",
				NeedRemove: true,
			},
			description: "empty file should be removed",
		},
		{
			name: "UNSET",
			expected: EnvValue{
				Value:      "",
				NeedRemove: true,
			},
			description: "file with only newline should be removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envValue, exists := env[tt.name]
			if !exists {
				t.Errorf("Variable %s not found", tt.name)
				return
			}

			if envValue.Value != tt.expected.Value {
				t.Errorf("%s: Value = %q, want %q", tt.name, envValue.Value, tt.expected.Value)
			}
			if envValue.NeedRemove != tt.expected.NeedRemove {
				t.Errorf("%s: NeedRemove = %v, want %v", tt.name, envValue.NeedRemove, tt.expected.NeedRemove)
			}
		})
	}
}

func TestNewEnvValue_WithTestdataFiles(t *testing.T) {
	testDataPath := filepath.Join("testdata", "env")

	tests := []struct {
		filename string
		want     EnvValue
	}{
		{
			filename: "HELLO",
			want: EnvValue{
				Value:      "\"hello\"",
				NeedRemove: false,
			},
		},
		{
			filename: "BAR",
			want: EnvValue{
				Value:      "bar",
				NeedRemove: false,
			},
		},
		{
			filename: "FOO",
			want: EnvValue{
				Value:      "   foo\nwith new line",
				NeedRemove: false,
			},
		},
		{
			filename: "EMPTY",
			want: EnvValue{
				Value:      "",
				NeedRemove: true,
			},
		},
		{
			filename: "UNSET",
			want: EnvValue{
				Value:      "",
				NeedRemove: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := NewEnvValue(testDataPath, tt.filename)
			if err != nil {
				t.Errorf("NewEnvValue() error = %v", err)
				return
			}

			if got.Value != tt.want.Value {
				t.Errorf("Value = %q, want %q", got.Value, tt.want.Value)
			}
			if got.NeedRemove != tt.want.NeedRemove {
				t.Errorf("NeedRemove = %v, want %v", got.NeedRemove, tt.want.NeedRemove)
			}
		})
	}
}

func TestReadDir_Testdata_Count(t *testing.T) {
	testDataPath := filepath.Join("testdata", "env")

	env, err := ReadDir(testDataPath)
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	// Должно быть ровно 5 переменных
	expectedCount := 5
	if len(env) != expectedCount {
		t.Errorf("Expected %d variables, got %d: %v", expectedCount, len(env), getKeys(env))
	}
}

func TestReadDir_Testdata_NoUnexpectedVars(t *testing.T) {
	testDataPath := filepath.Join("testdata", "env")

	env, err := ReadDir(testDataPath)
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	// Проверяем что нет неожиданных переменных
	allowedVars := map[string]bool{
		"BAR":   true,
		"FOO":   true,
		"HELLO": true,
		"EMPTY": true,
		"UNSET": true,
	}

	for varName := range env {
		if !allowedVars[varName] {
			t.Errorf("Unexpected variable found: %s", varName)
		}
	}
}

func getKeys(env Environment) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	return keys
}
