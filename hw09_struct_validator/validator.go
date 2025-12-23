package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var ErrNotStruct = errors.New("input is not a struct")

type ValidationField struct {
	Name         string
	TypeName     string
	Value        interface{}
	BaseTypeName string
	BaseValue    interface{}
	Validators   map[string]ValidateRule
}

type Validator struct {
	Method interface{}
}

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation errors: ")
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %v", err.Field, err.Err))
	}
	return sb.String()
}

func initValidators() map[string]Validator {
	validatorMap := make(map[string]Validator)
	validatorMap["max_int"] = Validator{GetMaxFunc[int]()}
	validatorMap["max_int64"] = Validator{GetMaxFunc[int64]()}
	validatorMap["max_int32"] = Validator{GetMaxFunc[int32]()}
	validatorMap["max_string"] = Validator{GetMaxFunc[string]()}
	validatorMap["max_[]int"] = Validator{GetMaxFunc[int]()}
	validatorMap["max_[]int64"] = Validator{GetMaxFunc[int64]()}
	validatorMap["max_[]int32"] = Validator{GetMaxFunc[int32]()}
	validatorMap["max_[]string"] = Validator{GetMaxFunc[string]()}
	validatorMap["min_[]int"] = Validator{GetMinFunc[int]()}
	validatorMap["min_[]int64"] = Validator{GetMinFunc[int64]()}
	validatorMap["min_[]int32"] = Validator{GetMinFunc[int32]()}
	validatorMap["min_[]string"] = Validator{GetMinFunc[string]()}
	validatorMap["min_int"] = Validator{GetMinFunc[int]()}
	validatorMap["min_int64"] = Validator{GetMinFunc[int64]()}
	validatorMap["min_int32"] = Validator{GetMinFunc[int32]()}
	validatorMap["min_string"] = Validator{GetMinFunc[string]()}
	validatorMap["in_int64"] = Validator{GetInFunc[int64]()}
	validatorMap["in_int32"] = Validator{GetInFunc[int32]()}
	validatorMap["in_string"] = Validator{GetInFunc[string]()}
	validatorMap["in_[]int"] = Validator{GetInFunc[int]()}
	validatorMap["in_[]int64"] = Validator{GetInFunc[int64]()}
	validatorMap["in_[]int32"] = Validator{GetInFunc[int32]()}
	validatorMap["in_[]string"] = Validator{GetInFunc[string]()}
	validatorMap["len_string"] = Validator{GetLenFunc()}
	validatorMap["len_[]string"] = Validator{GetLenFunc()}
	validatorMap["regexp_string"] = Validator{GetRegFunc()}
	validatorMap["regexp_[]string"] = Validator{GetRegFunc()}

	return validatorMap
}

func NewValidationField(t reflect.Type, v reflect.Value, index int) ValidationField {
	field := t.Field(index)
	value := v.Field(index)

	baseValue, baseTypeName := extractBaseType(value)
	validatorField := ValidationField{
		Name:         field.Name,
		TypeName:     field.Type.Name(),
		Value:        value,
		Validators:   make(map[string]ValidateRule),
		BaseValue:    baseValue,
		BaseTypeName: baseTypeName,
	}

	if value.IsValid() && value.CanInterface() {
		validatorField.Value = value.Interface()
	}

	for _, rule := range NewValidateTag(field) {
		validatorField.Validators[rule.Name+"_"+validatorField.BaseTypeName] = rule
	}
	return validatorField
}

func Validate(v interface{}) error {
	var validationErrors ValidationErrors
	if v == nil {
		return ErrNotStruct
	}
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	validators := initValidators()

	t := reflect.TypeOf(v)

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		if t.Field(i).Tag == "" {
			continue
		}

		if !t.Field(i).IsExported() {
			continue
		}
		validField := NewValidationField(t, val, i)
		if len(validField.Validators) == 0 {
			continue
		}
		if val.Field(i).Kind() == reflect.Slice {
			for j := 0; j < val.Field(i).Len(); j++ {
				elem := val.Field(i).Index(j)
				validField.BaseValue = elem.Interface()
				errs, err := validateField(validField, validators)
				if err != nil {
					return fmt.Errorf("field %s: %w", validField.Name, err)
				}
				validationErrors = append(validationErrors, errs...)
			}
		} else {
			errs, err := validateField(validField, validators)
			if err != nil {
				return fmt.Errorf("field %s: %w", validField.Name, err)
			}
			validationErrors = append(validationErrors, errs...)
		}
	}
	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

func validateField(validField ValidationField, validators map[string]Validator) (ValidationErrors, error) {
	var validationErrors ValidationErrors
	for key, rule := range validField.Validators {
		validator, ok := validators[key]
		if !ok {
			return nil, fmt.Errorf("validator %s not found", key)
		}
		args := make([]interface{}, 0)
		args = append(args, validField.BaseValue)
		args = append(args, rule.Args)
		err := CallFuncFromMap(validator.Method, args)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidLength),
				errors.Is(err, ErrNotInSet),
				errors.Is(err, ErrLessThanMin),
				errors.Is(err, ErrGreaterThanMax),
				errors.Is(err, ErrRegexpMismatch):
				validationErrors = append(validationErrors, ValidationError{Field: validField.Name, Err: err})
			default:
				return nil, err
			}
		}
	}
	return validationErrors, nil
}

func CallFuncFromMap(fn interface{}, args []interface{}) error {
	v := reflect.ValueOf(fn)

	callArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		callArgs[i] = reflect.ValueOf(arg)
	}

	results := v.Call(callArgs)
	out := make([]interface{}, len(results))

	for i, r := range results {
		out[i] = r.Interface()
	}

	if len(out) > 0 && out[0] != nil {
		if err, ok := out[0].(error); ok {
			return err
		}
	}
	return nil
}

//nolint:exhaustive
func extractBaseType(v reflect.Value) (interface{}, string) {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, "nil"
		}
		v = v.Elem()
	}
	kind := v.Kind()
	switch kind {
	case reflect.String:
		return v.String(), "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), "int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), "uint64"
	case reflect.Float32, reflect.Float64:
		return v.Float(), "float64"
	case reflect.Bool:
		return v.Bool(), "bool"
	case reflect.Slice:
		elemKind := v.Type().Elem().Kind()
		return v.Interface(),
			fmt.Sprintf("[]%v", elemKind)
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Map, reflect.Struct, reflect.UnsafePointer, reflect.Pointer | reflect.Ptr:
		return v.Interface(), "not supported"
	default:
		return v.Interface(), kind.String()
	}
}
