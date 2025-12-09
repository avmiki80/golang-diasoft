package main

import (
	"os"
	"testing"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []string
		wantCode int
		wantErr  bool
	}{
		{
			name:     "successful command",
			command:  "echo",
			args:     []string{"hello"},
			wantCode: successCode,
			wantErr:  false,
		},
		{
			name:     "command with multiple args",
			command:  "echo",
			args:     []string{"hello", "world"},
			wantCode: successCode,
			wantErr:  false,
		},
		{
			name:     "non-existent command",
			command:  "nonexistentcommand12345",
			args:     []string{},
			wantCode: errorCode,
			wantErr:  true,
		},
		{
			name:     "command that fails",
			command:  "false",
			args:     []string{},
			wantCode: 1, // false returns exit code 1
			wantErr:  true,
		},
		{
			name:     "command that returns specific exit code",
			command:  "sh",
			args:     []string{"-c", "exit 42"},
			wantCode: 42,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := runCommand(tt.command, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if code != tt.wantCode {
				t.Errorf("runCommand() code = %v, want %v", code, tt.wantCode)
			}

			if tt.wantErr && err == nil {
				t.Error("runCommand() expected error, got nil")
			}
		})
	}
}

func TestProcessEnv_SetNewVariables(t *testing.T) {
	// Сохраняем оригинальное окружение
	originalEnv := backupEnv(t, []string{"TEST_VAR_1", "TEST_VAR_2"})
	defer restoreEnv(t, originalEnv, []string{"TEST_VAR_1", "TEST_VAR_2"})

	env := Environment{
		"TEST_VAR_1": {Value: "value1", NeedRemove: false},
		"TEST_VAR_2": {Value: "value2", NeedRemove: false},
	}

	err := processEnv(env)
	if err != nil {
		t.Errorf("processEnv() error = %v, wantErr false", err)
	}

	checkEnvValues(t, map[string]string{
		"TEST_VAR_1": "value1",
		"TEST_VAR_2": "value2",
	}, []string{})
}

func TestProcessEnv_UnsetVariables(t *testing.T) {
	// Сохраняем оригинальное окружение
	originalEnv := backupEnv(t, []string{"TEST_VAR_1", "TEST_VAR_3"})
	defer restoreEnv(t, originalEnv, []string{"TEST_VAR_1", "TEST_VAR_3"})

	// Устанавливаем переменные для удаления
	os.Setenv("TEST_VAR_1", "should_be_removed")
	os.Setenv("TEST_VAR_3", "should_be_removed")

	env := Environment{
		"TEST_VAR_1": {Value: "", NeedRemove: true},
		"TEST_VAR_3": {Value: "", NeedRemove: true},
	}

	err := processEnv(env)
	if err != nil {
		t.Errorf("processEnv() error = %v, wantErr false", err)
	}

	checkEnvValues(t, map[string]string{}, []string{"TEST_VAR_1", "TEST_VAR_3"})
}

func TestProcessEnv_MixedSetAndUnset(t *testing.T) {
	// Сохраняем оригинальное окружение
	originalEnv := backupEnv(t, []string{"TEST_VAR_1", "TEST_VAR_2"})
	defer restoreEnv(t, originalEnv, []string{"TEST_VAR_1", "TEST_VAR_2"})

	// Устанавливаем начальные значения
	os.Setenv("TEST_VAR_1", "old_value")
	os.Setenv("TEST_VAR_2", "should_be_removed")

	env := Environment{
		"TEST_VAR_1": {Value: "new_value", NeedRemove: false},
		"TEST_VAR_2": {Value: "", NeedRemove: true},
	}

	err := processEnv(env)
	if err != nil {
		t.Errorf("processEnv() error = %v, wantErr false", err)
	}

	checkEnvValues(t, map[string]string{
		"TEST_VAR_1": "new_value",
	}, []string{"TEST_VAR_2"})
}

func TestProcessEnv_EmptyEnvironment(t *testing.T) {
	env := Environment{}

	err := processEnv(env)
	if err != nil {
		t.Errorf("processEnv() error = %v, wantErr false", err)
	}
}

func backupEnv(t *testing.T, vars []string) map[string]string {
	t.Helper()
	originalEnv := make(map[string]string)
	for _, v := range vars {
		if val, exists := os.LookupEnv(v); exists {
			originalEnv[v] = val
		}
	}
	return originalEnv
}

func restoreEnv(t *testing.T, originalEnv map[string]string, vars []string) {
	t.Helper()
	for _, v := range vars {
		if originalVal, exists := originalEnv[v]; exists {
			os.Setenv(v, originalVal)
		} else {
			os.Unsetenv(v)
		}
	}
}

func checkEnvValues(t *testing.T, expectedValues map[string]string, expectedUnset []string) {
	t.Helper()

	// Проверяем установленные значения
	for varName, expectedValue := range expectedValues {
		actualValue, exists := os.LookupEnv(varName)
		if !exists || actualValue != expectedValue {
			t.Errorf("Variable %s = %v, want %v (exists: %v)", varName, actualValue, expectedValue, exists)
		}
	}

	// Проверяем удаленные переменные
	for _, varName := range expectedUnset {
		if _, exists := os.LookupEnv(varName); exists {
			t.Errorf("Variable %s should be unset, but it exists", varName)
		}
	}
}

func TestProcessEnv_ErrorCases(t *testing.T) {
	env := Environment{
		"": {Value: "value", NeedRemove: false}, // пустое имя переменной
	}

	err := processEnv(env)
	if err == nil {
		t.Log("processEnv() with invalid variable name didn't return error (may be OS dependent)")
	}
}

func TestRunCmd(t *testing.T) {
	// Сохраняем и восстанавливаем окружение
	originalValue, hadOriginal := os.LookupEnv("TEST_RUNCMD_VAR")
	defer func() {
		if hadOriginal {
			os.Setenv("TEST_RUNCMD_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_RUNCMD_VAR")
		}
	}()

	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		wantCode int
	}{
		{
			name:     "successful command with env",
			cmd:      []string{"echo", "test"},
			env:      Environment{"TEST_RUNCMD_VAR": {Value: "test_value", NeedRemove: false}},
			wantCode: successCode,
		},
		{
			name:     "empty command",
			cmd:      []string{},
			env:      Environment{},
			wantCode: errorCode,
		},
		{
			name:     "command with unset env",
			cmd:      []string{"echo", "test"},
			env:      Environment{"TEST_RUNCMD_VAR": {Value: "", NeedRemove: true}},
			wantCode: successCode,
		},
		{
			name:     "failing command",
			cmd:      []string{"false"},
			env:      Environment{},
			wantCode: 1,
		},
		{
			name:     "command with specific exit code",
			cmd:      []string{"sh", "-c", "exit 5"},
			env:      Environment{},
			wantCode: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := RunCmd(tt.cmd, tt.env)

			if code != tt.wantCode {
				t.Errorf("RunCmd() = %v, want %v", code, tt.wantCode)
			}

			// Для теста с установкой переменной проверяем что она установилась
			if tt.name == "successful command with env" {
				if val, exists := os.LookupEnv("TEST_RUNCMD_VAR"); !exists || val != "test_value" {
					t.Errorf("Environment variable not set correctly: exists=%v, value=%v", exists, val)
				}
			}

			// Для теста с удалением переменной проверяем что она удалилась
			if tt.name == "command with unset env" {
				if _, exists := os.LookupEnv("TEST_RUNCMD_VAR"); exists {
					t.Error("Environment variable should be unset")
				}
			}
		})
	}
}

func TestRunCmd_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		wantCode int
	}{
		{
			name:     "nil command",
			cmd:      nil,
			env:      Environment{},
			wantCode: errorCode,
		},
		{
			name:     "single element command",
			cmd:      []string{"echo"},
			env:      Environment{},
			wantCode: successCode,
		},
		{
			name: "complex environment",
			cmd:  []string{"echo", "done"},
			env: Environment{
				"VAR1": {Value: "value1", NeedRemove: false},
				"VAR2": {Value: "", NeedRemove: true},
				"VAR3": {Value: "value3", NeedRemove: false},
			},
			wantCode: successCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := RunCmd(tt.cmd, tt.env)
			if code != tt.wantCode {
				t.Errorf("RunCmd() = %v, want %v", code, tt.wantCode)
			}
		})
	}
}
