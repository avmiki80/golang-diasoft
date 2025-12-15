package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

var (
	ErrInvalidLength  = errors.New("invalid length")
	ErrNotInSet       = errors.New("value not in set")
	ErrRegexpMismatch = errors.New("regexp mismatch")
	ErrLessThanMin    = errors.New("value less than min")
	ErrGreaterThanMax = errors.New("value greater than max")
)

var (
	ErrNegativeLength     = errors.New("length limit must be >= 0")
	ErrToManyParams       = errors.New("too many params")
	ErrInsufficientParams = errors.New("insufficient params")
	ErrConvertParam       = errors.New("param is not converted to int")
)

type Numbers interface {
	~int64 | ~int | ~int32 | ~int16 | ~int8 | ~uint64 | ~uint32 | ~uint16 | ~uint8 |
		~float64 | ~float32
}

func GetMaxFunc[T Numbers | string]() func(T, []string) error {
	return func(value T, params []string) error {
		err := checkEmptyParams(params)
		if err != nil {
			return err
		}

		err = checkTooManyParam(params)
		if err != nil {
			return err
		}

		convertedParams, err := convertParamsToType(value, params)
		if err != nil {
			return ErrConvertParam
		}

		if value <= convertedParams[0] {
			return nil
		}
		return ErrGreaterThanMax
	}
}

func GetMinFunc[T Numbers | string]() func(T, []string) error {
	return func(value T, params []string) error {
		err := checkEmptyParams(params)
		if err != nil {
			return err
		}

		err = checkTooManyParam(params)
		if err != nil {
			return err
		}

		convertedParams, err := convertParamsToType(value, params)
		if err != nil {
			return ErrConvertParam
		}

		if value >= convertedParams[0] {
			return nil
		}
		return ErrLessThanMin
	}
}

func GetInFunc[T comparable]() func(T, []string) error {
	return func(value T, params []string) error {
		err := checkEmptyParams(params)
		if err != nil {
			return err
		}
		var convParams []T
		convParams, err = convertParamsToType(value, params)
		if err != nil {
			return ErrConvertParam
		}
		for _, param := range convParams {
			if param == value {
				return nil
			}
		}
		return ErrNotInSet
	}
}

func GetLenFunc() func(string, []string) error {
	return func(value string, params []string) error {
		err := checkEmptyParams(params)
		if err != nil {
			return err
		}

		err = checkTooManyParam(params)
		if err != nil {
			return err
		}
		limit, err := strconv.Atoi(params[0])
		if err != nil {
			return ErrConvertParam
		}

		if limit < 0 {
			return ErrNegativeLength
		}
		if len(value) == limit {
			return nil
		}
		return ErrInvalidLength
	}
}

func GetRegFunc() func(string, []string) error {
	return func(value string, params []string) error {
		err := checkEmptyParams(params)
		if err != nil {
			return err
		}

		err = checkTooManyParam(params)
		if err != nil {
			return err
		}
		param := params[0]
		re, err := regexp.Compile(param)
		if err != nil {
			return fmt.Errorf("invalid regexp pattern: %w", err)
		}
		if !re.MatchString(value) {
			return ErrRegexpMismatch
		}
		return nil
	}
}

func checkEmptyParams[T any](param []T) error {
	if len(param) == 0 {
		return ErrInsufficientParams
	}
	return nil
}

func checkTooManyParam[T any](param []T) error {
	if len(param) > 1 {
		return ErrToManyParams
	}
	return nil
}

//nolint:exhaustive
func convertParamsToType[T any](value T, params []string) ([]T, error) {
	if len(params) == 0 {
		return nil, ErrInsufficientParams
	}

	result := make([]T, 0, len(params))
	valueType := reflect.TypeOf(value)

	for i, param := range params {
		if valueType.Kind() == reflect.String {
			result = append(result, any(param).(T))
			continue
		}

		var converted reflect.Value

		switch valueType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, parseErr := strconv.ParseInt(param, 10, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("param at index %d: cannot parse '%s' as int: %w", i, param, parseErr)
			}
			converted = reflect.ValueOf(val).Convert(valueType)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, parseErr := strconv.ParseUint(param, 10, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("param at index %d: cannot parse '%s' as uint: %w", i, param, parseErr)
			}
			converted = reflect.ValueOf(val).Convert(valueType)

		case reflect.Float32, reflect.Float64:
			val, parseErr := strconv.ParseFloat(param, 64)
			if parseErr != nil {
				return nil, fmt.Errorf("param at index %d: cannot parse '%s' as float: %w", i, param, parseErr)
			}
			converted = reflect.ValueOf(val).Convert(valueType)

		case reflect.Bool:
			val, parseErr := strconv.ParseBool(param)
			if parseErr != nil {
				return nil, fmt.Errorf("param at index %d: cannot parse '%s' as bool: %w", i, param, parseErr)
			}
			converted = reflect.ValueOf(val)

		default:
			return nil, fmt.Errorf("param at index %d: unsupported type conversion from string to %v", i, valueType)
		}

		result = append(result, converted.Interface().(T))
	}

	return result, nil
}
