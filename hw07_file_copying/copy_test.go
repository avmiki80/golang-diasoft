package main

import (
	"errors"
	"os"
	"testing"
)

const content = "test content"

func setupTestFiles(t *testing.T, content string) (string, string) {
	t.Helper()

	fromFile, err := os.CreateTemp("", "test_from_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp from file: %v", err)
	}
	defer fromFile.Close()

	toFile, err := os.CreateTemp("", "test_to_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp to file: %v", err)
	}
	toPath := toFile.Name()
	toFile.Close()

	// Записываем контент в исходный файл
	if _, err := fromFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to from file: %v", err)
	}

	return fromFile.Name(), toPath
}

func TestNewFromCopy(t *testing.T) {
	fromPath, _ := setupTestFiles(t, content)
	defer os.Remove(fromPath)

	fromCopy, err := NewFromCopy(fromPath)
	if err != nil {
		t.Fatalf("NewFromCopy failed: %v", err)
	}
	defer fromCopy.Close()

	if fromCopy.From == nil {
		t.Error("From file should not be nil")
	}
}

func TestNewFromCopy_FileNotExists(t *testing.T) {
	_, err := NewFromCopy("nonexistent_file.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestFromCopy_CheckFile(t *testing.T) {
	t.Run("ValidFile", func(t *testing.T) {
		fromPath, _ := setupTestFiles(t, content)
		defer os.Remove(fromPath)

		fromCopy, err := NewFromCopy(fromPath)
		if err != nil {
			t.Fatalf("NewFromCopy failed: %v", err)
		}
		defer fromCopy.Close()

		err = fromCopy.CheckFile()
		if err != nil {
			t.Errorf("CheckFile should not fail for valid file: %v", err)
		}
	})

	t.Run("Directory", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "test_dir")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		fromCopy, err := NewFromCopy(dir)
		if err != nil {
			t.Fatalf("NewFromCopy failed: %v", err)
		}
		defer fromCopy.Close()

		err = fromCopy.CheckFile()
		if !errors.Is(err, ErrUnsupportedFile) {
			t.Errorf("Expected ErrUnsupportedFile, got: %v", err)
		}
	})
}

func TestFromCopy_CheckOffset(t *testing.T) {
	fromPath, _ := setupTestFiles(t, content)
	defer os.Remove(fromPath)

	fromCopy, err := NewFromCopy(fromPath)
	if err != nil {
		t.Fatalf("NewFromCopy failed: %v", err)
	}
	defer fromCopy.Close()

	tests := []struct {
		name     string
		offset   int64
		expected error
	}{
		{"ValidOffset", 5, nil},
		{"ZeroOffset", 0, nil},
		{"OffsetEqualToSize", int64(len(content)), nil},
		{"OffsetExceedsSize", int64(len(content) + 10), ErrOffsetExceedsFileSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fromCopy.CheckOffset(tt.offset)
			if !errors.Is(err, tt.expected) {
				t.Errorf("CheckOffset(%d) = %v, expected %v", tt.offset, err, tt.expected)
			}
		})
	}
}

func TestFromCopy_CheckLimit(t *testing.T) {
	content := "test content for limit checking"
	fileSize := int64(len(content))
	fromPath, _ := setupTestFiles(t, content)
	defer os.Remove(fromPath)

	fromCopy, err := NewFromCopy(fromPath)
	if err != nil {
		t.Fatalf("NewFromCopy failed: %v", err)
	}
	defer fromCopy.Close()

	tests := []struct {
		name          string
		offset        int64
		limit         int64
		expectedLimit int64
		expectError   bool
	}{
		{"LimitWithinBounds", 5, 10, 10, false},
		{"LimitExceedsBounds", 5, 100, fileSize - 5, false},
		{"NegativeLimit", 5, -1, fileSize - 5, false},
		{"ZeroLimit", 5, 0, 0, false},
		{"FullFileCopy", 0, -1, fileSize, false},
		{"OffsetAtEnd", fileSize, 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fromCopy.CheckLimit(tt.offset, tt.limit)
			if (err != nil) != tt.expectError {
				t.Errorf("CheckLimit(%d, %d) error = %v, expectError = %v", tt.offset, tt.limit, err, tt.expectError)
				return
			}
			if result != tt.expectedLimit {
				t.Errorf("CheckLimit(%d, %d) = %d, expected %d", tt.offset, tt.limit, result, tt.expectedLimit)
			}
		})
	}
}

func TestNewToCopy(t *testing.T) {
	toFile, err := os.CreateTemp("", "test_to_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	toPath := toFile.Name()
	toFile.Close()
	defer os.Remove(toPath)

	toCopy, err := NewToCopy(toPath)
	if err != nil {
		t.Fatalf("NewToCopy failed: %v", err)
	}
	defer toCopy.Close()

	if toCopy.To == nil {
		t.Error("To file should not be nil")
	}
}

func TestCopy_FullFile(t *testing.T) {
	content := "this is a test content for full file copy"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	err := Copy(fromPath, toPath, 0, -1)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Проверяем что файл скопирован полностью
	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != content {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", content, string(copiedContent))
	}
}

func TestCopy_WithOffset(t *testing.T) {
	content := "this is a test content"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	// Копируем с offset 5
	err := Copy(fromPath, toPath, 5, -1)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "is a test content"
	if string(copiedContent) != expected {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", expected, string(copiedContent))
	}
}

func TestCopy_WithLimit(t *testing.T) {
	content := "this is a test content"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	// Копируем только 4 байта с offset 5
	err := Copy(fromPath, toPath, 5, 4)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "is a"
	if string(copiedContent) != expected {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", expected, string(copiedContent))
	}
}

func TestCopy_OffsetAndLimit(t *testing.T) {
	content := "this is a test content for offset and limit"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	// Копируем 10 байт с offset 8
	err := Copy(fromPath, toPath, 8, 10)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "a test con"
	if string(copiedContent) != expected {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", expected, string(copiedContent))
	}
}

func TestCopy_OffsetExceedsFileSize(t *testing.T) {
	content := "short content"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	err := Copy(fromPath, toPath, int64(len(content)+10), 5)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Errorf("Expected ErrOffsetExceedsFileSize, got: %v", err)
	}
}

func TestCopy_EmptyFile(t *testing.T) {
	content := ""
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	err := Copy(fromPath, toPath, 0, -1)
	if err != nil {
		t.Fatalf("Copy failed for empty file: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != "" {
		t.Errorf("Expected empty content, got: %s", string(copiedContent))
	}
}

func TestCopy_NonexistentSource(t *testing.T) {
	toFile, err := os.CreateTemp("", "test_to_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	toPath := toFile.Name()
	toFile.Close()
	defer os.Remove(toPath)

	err = Copy("nonexistent_file.txt", toPath, 0, -1)
	if err == nil {
		t.Error("Expected error for nonexistent source file")
	}
}

func TestCopy_ToInvalidPath(t *testing.T) {
	content := "test content"
	fromPath, _ := setupTestFiles(t, content)
	defer os.Remove(fromPath)

	// Пытаемся записать в невалидный путь
	err := Copy(fromPath, "/invalid/path/file.txt", 0, -1)
	if err == nil {
		t.Error("Expected error for invalid destination path")
	}
}

func TestProcessCopy_ZeroLimit(t *testing.T) {
	content := "test content for zero limit"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	fromCopy, err := NewFromCopy(fromPath)
	if err != nil {
		t.Fatalf("NewFromCopy failed: %v", err)
	}
	defer fromCopy.Close()

	toCopy, err := NewToCopy(toPath)
	if err != nil {
		t.Fatalf("NewToCopy failed: %v", err)
	}
	defer toCopy.Close()

	// Устанавливаем offset
	err = fromCopy.Seek(0)
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}

	// Копируем с limit = 0 (весь файл)
	err = processCopy(fromCopy, toCopy, 0)
	if err != nil {
		t.Fatalf("processCopy failed: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != content {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", content, string(copiedContent))
	}
}

func TestProcessCopy_WithLimit(t *testing.T) {
	content := "test content for process copy with limit"
	fromPath, toPath := setupTestFiles(t, content)
	defer os.Remove(fromPath)
	defer os.Remove(toPath)

	fromCopy, err := NewFromCopy(fromPath)
	if err != nil {
		t.Fatalf("NewFromCopy failed: %v", err)
	}
	defer fromCopy.Close()

	toCopy, err := NewToCopy(toPath)
	if err != nil {
		t.Fatalf("NewToCopy failed: %v", err)
	}
	defer toCopy.Close()

	// Устанавливаем offset
	err = fromCopy.Seek(5)
	if err != nil {
		t.Fatalf("Seek failed: %v", err)
	}

	// Копируем только 10 байт
	err = processCopy(fromCopy, toCopy, 10)
	if err != nil {
		t.Fatalf("processCopy failed: %v", err)
	}

	copiedContent, err := os.ReadFile(toPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "content fo"
	if string(copiedContent) != expected {
		t.Errorf("Copied content mismatch.\nExpected: %s\nGot: %s", expected, string(copiedContent))
	}
}
