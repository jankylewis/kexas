package pages

// Page.Evaluate returns interface{} wrapped in a CDP result envelope. These
// helpers coerce the common cases. Same pattern as sltests/pages/eval_helpers.go.

// numericFromEvaluate coerces to int. Returns 0 on type mismatch.
func numericFromEvaluate(raw interface{}) int {
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			return numericFromEvaluate(v)
		}
	}
	if f, ok := raw.(float64); ok {
		return int(f)
	}
	return 0
}

// stringFromEvaluate coerces to string. Returns "" on type mismatch.
func stringFromEvaluate(raw interface{}) string {
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			return stringFromEvaluate(v)
		}
	}
	if s, ok := raw.(string); ok {
		return s
	}
	return ""
}
