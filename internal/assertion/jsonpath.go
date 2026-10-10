package assertion

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Operator string

const (
	OpEquals      Operator = "equals"
	OpNotEquals   Operator = "not_equals"
	OpContains    Operator = "contains"
	OpGreaterThan Operator = "greater_than"
	OpLessThan    Operator = "less_than"
	OpExists      Operator = "exists"
)

type JSONPathEvaluator struct{}

func NewJSONPathEvaluator() *JSONPathEvaluator {
	return &JSONPathEvaluator{}
}

func (e *JSONPathEvaluator) Extract(jsonData []byte, path string) (interface{}, error) {
	var parsed interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	trimmed := strings.TrimPrefix(path, "$.")
	trimmed = strings.TrimPrefix(trimmed, "$")
	if trimmed == "" {
		return parsed, nil
	}

	parts := strings.Split(trimmed, ".")
	current := parsed

	for _, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("key '%s' not found on nil object", part)
		}

		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			key := part[:strings.Index(part, "[")]
			idxStr := part[strings.Index(part, "[")+1 : len(part)-1]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index in '%s'", part)
			}

			if key != "" {
				m, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected map at key '%s'", key)
				}
				current = m[key]
			}

			slice, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected slice at index '%d'", idx)
			}
			if idx < 0 || idx >= len(slice) {
				return nil, fmt.Errorf("index %d out of bounds (len %d)", idx, len(slice))
			}
			current = slice[idx]
			continue
		}

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map for property '%s', got %T", part, current)
		}
		val, exists := m[part]
		if !exists {
			return nil, fmt.Errorf("property '%s' does not exist", part)
		}
		current = val
	}

	return current, nil
}

func (e *JSONPathEvaluator) Evaluate(jsonData []byte, path string, op Operator, expected string) (bool, string, error) {
	actual, err := e.Extract(jsonData, path)
	if err != nil {
		if op == OpExists && expected == "false" {
			return true, "", nil
		}
		return false, fmt.Sprintf("failed to extract '%s': %v", path, err), nil
	}

	if op == OpExists {
		return true, "", nil
	}

	actualStr := fmt.Sprintf("%v", actual)
	switch op {
	case OpEquals:
		if actualStr == expected {
			return true, "", nil
		}
		return false, fmt.Sprintf("expected %s == %s, got %s", path, expected, actualStr), nil
	case OpNotEquals:
		if actualStr != expected {
			return true, "", nil
		}
		return false, fmt.Sprintf("expected %s != %s, got %s", path, expected, actualStr), nil
	case OpContains:
		if strings.Contains(actualStr, expected) {
			return true, "", nil
		}
		return false, fmt.Sprintf("expected %s to contain '%s', got '%s'", path, expected, actualStr), nil
	case OpGreaterThan:
		aFloat, err1 := strconv.ParseFloat(actualStr, 64)
		eFloat, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil && aFloat > eFloat {
			return true, "", nil
		}
		return false, fmt.Sprintf("expected %s > %s, got %s", path, expected, actualStr), nil
	case OpLessThan:
		aFloat, err1 := strconv.ParseFloat(actualStr, 64)
		eFloat, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil && aFloat < eFloat {
			return true, "", nil
		}
		return false, fmt.Sprintf("expected %s < %s, got %s", path, expected, actualStr), nil
	default:
		return false, fmt.Sprintf("unsupported operator: %s", op), nil
	}
}
