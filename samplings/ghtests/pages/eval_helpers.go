package pages

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
