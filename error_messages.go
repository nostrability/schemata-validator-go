package validator

import "strings"

// enrichMessage tries to find a custom errorMessage in the schema.
func enrichMessage(schema interface{}, schemaPath string, keyword string, defaultMsg string) string {
	segments := splitPath(schemaPath)
	for depth := 0; depth <= 2; depth++ {
		if depth > len(segments) {
			break
		}
		checkSegments := segments[:len(segments)-depth]
		if node := walkSchema(schema, checkSegments); node != nil {
			if obj, ok := node.(map[string]interface{}); ok {
				if em, ok := obj["errorMessage"]; ok {
					switch v := em.(type) {
					case string:
						return v
					case map[string]interface{}:
						if msg, ok := v[keyword].(string); ok {
							return msg
						}
					}
				}
			}
		}
	}
	return defaultMsg
}

func splitPath(path string) []string {
	var segments []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

func walkSchema(schema interface{}, segments []string) interface{} {
	current := schema
	for _, seg := range segments {
		switch v := current.(type) {
		case map[string]interface{}:
			child, ok := v[seg]
			if !ok {
				return nil
			}
			current = child
		default:
			return nil
		}
	}
	return current
}
