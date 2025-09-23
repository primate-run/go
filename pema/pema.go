package pema

import (
	"fmt"
	"strconv"

	"github.com/primate-run/go/types"
)

type Dict = types.Dict

type Field[T any] interface {
	Parse(value any, coerce bool) (T, error)
}

type StringType struct{}
type BooleanType struct{}
type IntType struct{}
type Int64Type struct{}
type FloatType struct{}

func (StringType) Parse(value any, coerce bool) (string, error) {
	if s, ok := value.(string); ok {
		return s, nil
	}
	if coerce {
		return fmt.Sprintf("%v", value), nil
	}
	return "", fmt.Errorf("expected string, got %T", value)
}

func (BooleanType) Parse(value any, coerce bool) (bool, error) {
	if b, ok := value.(bool); ok {
		return b, nil
	}
	if coerce {
		switch v := value.(type) {
		case string:
			if v == "" {
				return false, nil
			}
			return strconv.ParseBool(v)
		default:
			return false, fmt.Errorf("cannot coerce %T to boolean", value)
		}
	}
	return false, fmt.Errorf("expected boolean, got %T", value)
}

func (IntType) Parse(value any, coerce bool) (int, error) {
	if i, ok := value.(int); ok {
		return i, nil
	}
	if coerce {
		switch v := value.(type) {
		case float64:
			return int(v), nil
		case string:
			if v == "" {
				return 0, nil
			}
			return strconv.Atoi(v)
		default:
			return 0, fmt.Errorf("cannot coerce %T to int", value)
		}
	}

	return 0, fmt.Errorf("expected int, got %T", value)
}

func (Int64Type) Parse(value any, coerce bool) (int64, error) {
	if i, ok := value.(int64); ok {
		return i, nil
	}
	if coerce {
		switch v := value.(type) {
		case int:
			return int64(v), nil
		case float64:
			return int64(v), nil
		case string:
			if v == "" {
				return int64(0), nil
			}
			return strconv.ParseInt(v, 10, 64)
		default:
			return 0, fmt.Errorf("cannot coerce %T to int64", value)
		}
	}
	return 0, fmt.Errorf("expected int64, got %T", value)
}

func (FloatType) Parse(value any, coerce bool) (float64, error) {
	if f, ok := value.(float64); ok {
		return f, nil
	}
	if coerce {
		switch v := value.(type) {
		case int:
			return float64(v), nil
		case string:
			if v == "" {
				return 0.0, nil
			}
			return strconv.ParseFloat(v, 64)
		default:
			return 0.0, fmt.Errorf("cannot coerce %T to float", value)
		}
	}
	return 0.0, fmt.Errorf("expected float64, got %T", value)
}

func String() Field[string] { return StringType{} }
func Boolean() Field[bool]  { return BooleanType{} }
func Int() Field[int]       { return IntType{} }
func Int64() Field[int64]   { return Int64Type{} }
func Float() Field[float64] { return FloatType{} }

type AnyField interface {
	parse(value any, coerce bool) (any, error)
}

type fieldWrapper[T any] struct {
	field Field[T]
}

func (w fieldWrapper[T]) parse(value any, coerce bool) (any, error) {
	return w.field.Parse(value, coerce)
}

type Fields = map[string]AnyField

type SchemaBuilder struct {
	fields Fields
}

func Schema(fields map[string]any) *SchemaBuilder {
	wrapped := make(Fields)
	for name, field := range fields {
		switch f := field.(type) {
		case Field[string]:
			wrapped[name] = fieldWrapper[string]{f}
		case Field[bool]:
			wrapped[name] = fieldWrapper[bool]{f}
		case Field[int]:
			wrapped[name] = fieldWrapper[int]{f}
		case Field[int64]:
			wrapped[name] = fieldWrapper[int64]{f}
		case Field[float64]:
			wrapped[name] = fieldWrapper[float64]{f}
		default:
			panic(fmt.Sprintf("unsupported field type: %T", field))
		}
	}
	return &SchemaBuilder{fields: wrapped}
}

func (s *SchemaBuilder) Parse(data Dict, args ...bool) (Dict, error) {
	coerce := false
	if len(args) > 0 {
		coerce = args[0]
	}
	result := make(Dict)

	for name, field := range s.fields {
		value, exists := data[name]
		if !exists {
			value = ""
		}

		parsed, err := field.parse(value, coerce)
		if err != nil {
			return nil, fmt.Errorf("parsing failed for field '%s': %w", name, err)
		}

		result[name] = parsed
	}

	return result, nil
}
