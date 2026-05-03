package pages

import "fmt"

// numericResult coerces a Page.Evaluate JSON result into an int. Page.Evaluate
// returns numbers as float64 (JSON's default numeric type).
func numericResult(raw interface{}) (int, error) {
	if raw == nil {
		return 0, nil
	}
	if obj, ok := raw.(map[string]interface{}); ok {
		// CDP wraps results as {type: "number", value: 6, ...}; try the value field.
		if v, exists := obj["value"]; exists {
			return numericResult(v)
		}
	}
	if f, ok := raw.(float64); ok {
		return int(f), nil
	}
	return 0, fmt.Errorf("evaluate returned non-numeric result: %T = %v", raw, raw)
}

// boolResult coerces a Page.Evaluate JSON result into a bool.
func boolResult(raw interface{}) bool {
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	if b, ok := raw.(bool); ok {
		return b
	}
	return false
}

// stringSliceResult coerces a Page.Evaluate result that should be an array of strings.
// Handles both the raw JS array form and the CDP-wrapped {type: "object", ...} form.
func stringSliceResult(raw interface{}) ([]string, error) {
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			return stringSliceResult(v)
		}
	}
	if arr, ok := raw.([]interface{}); ok {
		out := make([]string, len(arr))
		for i, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("array element %d is not a string: %T = %v", i, item, item)
			}
			out[i] = s
		}
		return out, nil
	}
	return nil, fmt.Errorf("evaluate returned non-array result: %T = %v", raw, raw)
}
