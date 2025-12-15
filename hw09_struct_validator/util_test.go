package hw09structvalidator

import (
	"errors"
	"testing"
)

func TestGetLenFunc(t *testing.T) {
	lenFunc := GetLenFunc()

	tests := []struct {
		name    string
		value   string
		params  []string
		wantErr error
	}{
		{
			name:    "correct length - valid",
			value:   "hello",
			params:  []string{"5"},
			wantErr: nil,
		},
		{
			name:    "incorrect length - invalid",
			value:   "hello",
			params:  []string{"3"},
			wantErr: ErrInvalidLength,
		},
		{
			name:    "empty string with zero length",
			value:   "",
			params:  []string{"0"},
			wantErr: nil,
		},
		{
			name:    "negative length",
			value:   "test",
			params:  []string{"-1"},
			wantErr: ErrNegativeLength,
		},
		{
			name:    "empty params",
			value:   "test",
			params:  []string{},
			wantErr: ErrInsufficientParams,
		},
		{
			name:    "too many params",
			value:   "test",
			params:  []string{"4", "5"},
			wantErr: ErrToManyParams,
		},
		{
			name:    "invalid param format",
			value:   "test",
			params:  []string{"abc"},
			wantErr: ErrConvertParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lenFunc(tt.value, tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetLenFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetRegFunc(t *testing.T) {
	regFunc := GetRegFunc()

	tests := []struct {
		name    string
		value   string
		params  []string
		wantErr error
	}{
		{
			name:    "regexp match - valid",
			value:   "hello123",
			params:  []string{`^[a-z]+\d+$`},
			wantErr: nil,
		},
		{
			name:    "regexp mismatch - invalid",
			value:   "hello",
			params:  []string{`^\d+$`},
			wantErr: ErrRegexpMismatch,
		},
		{
			name:    "email validation - valid",
			value:   "test@example.com",
			params:  []string{`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`},
			wantErr: nil,
		},
		{
			name:    "empty params",
			value:   "test",
			params:  []string{},
			wantErr: ErrInsufficientParams,
		},
		{
			name:    "too many params",
			value:   "test",
			params:  []string{`\d+`, `\w+`},
			wantErr: ErrToManyParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := regFunc(tt.value, tt.params)
			if tt.wantErr == nil && err != nil {
				t.Errorf("GetRegFunc() unexpected error = %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("GetRegFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConvertParamsToInt64(t *testing.T) {
	t.Run("convert to int64", func(t *testing.T) {
		result, err := convertParamsToType(int64(0), []string{"1", "2", "3"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := []int64{1, 2, 3}
		if len(result) != len(expected) {
			t.Errorf("length mismatch: got %d, want %d", len(result), len(expected))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("value at index %d: got %v, want %v", i, v, expected[i])
			}
		}
	})
}

func TestConvertParamsToString(t *testing.T) {
	t.Run("convert to string", func(t *testing.T) {
		result, err := convertParamsToType("", []string{"hello", "world"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := []string{"hello", "world"}
		if len(result) != len(expected) {
			t.Errorf("length mismatch: got %d, want %d", len(result), len(expected))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("value at index %d: got %v, want %v", i, v, expected[i])
			}
		}
	})
}

func TestConvertParamsToFloat64(t *testing.T) {
	t.Run("convert to float64", func(t *testing.T) {
		result, err := convertParamsToType(float64(0), []string{"1.5", "2.7", "3.14"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := []float64{1.5, 2.7, 3.14}
		if len(result) != len(expected) {
			t.Errorf("length mismatch: got %d, want %d", len(result), len(expected))
		}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("value at index %d: got %v, want %v", i, v, expected[i])
			}
		}
	})
}

func TestConvertParamsToType(t *testing.T) {
	t.Run("invalid int conversion", func(t *testing.T) {
		_, err := convertParamsToType(int64(0), []string{"abc"})
		if err == nil {
			t.Error("expected error for invalid int conversion")
		}
	})

	t.Run("empty params", func(t *testing.T) {
		_, err := convertParamsToType(int64(0), []string{})
		if !errors.Is(err, ErrInsufficientParams) {
			t.Errorf("expected ErrInsufficientParams, got %v", err)
		}
	})
}

func TestGetMaxFunc(t *testing.T) {
	maxFunc := GetMaxFunc[int]()
	tests := []struct {
		name    string
		value   int
		params  []string
		wantErr error
	}{
		{
			name:    "value less than max - valid",
			value:   5,
			params:  []string{"10"},
			wantErr: nil,
		},
		{
			name:    "value equal to max - valid",
			value:   10,
			params:  []string{"10"},
			wantErr: nil,
		},
		{
			name:    "value greater than max - invalid",
			value:   15,
			params:  []string{"10"},
			wantErr: ErrGreaterThanMax,
		},
		{
			name:    "empty params",
			value:   5,
			params:  []string{},
			wantErr: ErrInsufficientParams,
		},
		{
			name:    "too many params",
			value:   5,
			params:  []string{"10", "20"},
			wantErr: ErrToManyParams,
		},
		{
			name:    "invalid param format",
			value:   5,
			params:  []string{"abc"},
			wantErr: ErrConvertParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := maxFunc(tt.value, tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NMax() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	err := maxFunc(5, []string{"10"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = maxFunc(15, []string{"10"})
	if !errors.Is(err, ErrGreaterThanMax) {
		t.Errorf("expected ErrGreaterThanMax, got %v", err)
	}
}

func TestGetMinFunc(t *testing.T) {
	minFunc := GetMinFunc[int]()

	tests := []struct {
		name    string
		value   int
		params  []string
		wantErr error
	}{
		{
			name:    "value greater than min - valid",
			value:   15,
			params:  []string{"10"},
			wantErr: nil,
		},
		{
			name:    "value equal to min - valid",
			value:   10,
			params:  []string{"10"},
			wantErr: nil,
		},
		{
			name:    "value less than min - invalid",
			value:   5,
			params:  []string{"10"},
			wantErr: ErrLessThanMin,
		},
		{
			name:    "empty params",
			value:   5,
			params:  []string{},
			wantErr: ErrInsufficientParams,
		},
		{
			name:    "too many params",
			value:   5,
			params:  []string{"10", "20"},
			wantErr: ErrToManyParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := minFunc(tt.value, tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NMin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err := minFunc(15, []string{"10"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = minFunc(5, []string{"10"})
	if !errors.Is(err, ErrLessThanMin) {
		t.Errorf("expected ErrLessThanMin, got %v", err)
	}
}

func TestGetInFunc(t *testing.T) {
	inFunc := GetInFunc[string]()
	tests := []struct {
		name    string
		value   string
		params  []string
		wantErr error
	}{
		{
			name:    "value in set - valid",
			value:   "hello",
			params:  []string{"world", "hello", "test"},
			wantErr: nil,
		},
		{
			name:    "value not in set - invalid",
			value:   "foo",
			params:  []string{"bar", "baz"},
			wantErr: ErrNotInSet,
		},
		{
			name:    "empty params",
			value:   "test",
			params:  []string{},
			wantErr: ErrInsufficientParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inFunc(tt.value, tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("In() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err := inFunc("hello", []string{"hello", "world"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = inFunc("foo", []string{"bar", "baz"})
	if !errors.Is(err, ErrNotInSet) {
		t.Errorf("expected ErrNotInSet, got %v", err)
	}
}
