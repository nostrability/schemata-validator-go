package validator

import (
	"fmt"
	"regexp"
)

// collectAdditionalProps recursively detects undeclared properties as warnings.
func collectAdditionalProps(schema interface{}, data interface{}, path string) []ValidationError {
	dataObj, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}
	schemaObj, ok := schema.(map[string]interface{})
	if !ok {
		return nil
	}
	schemaType, _ := schemaObj["type"].(string)
	if schemaType != "object" {
		return nil
	}
	if ap, ok := schemaObj["additionalProperties"]; ok {
		if apBool, ok := ap.(bool); ok && !apBool {
			return nil
		}
	}

	var warnings []ValidationError
	allowed := make(map[string]bool)

	if props, ok := schemaObj["properties"].(map[string]interface{}); ok {
		for key := range props {
			allowed[key] = true
		}
	}
	if patterns, ok := schemaObj["patternProperties"].(map[string]interface{}); ok {
		for pattern := range patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				for key := range dataObj {
					if re.MatchString(key) {
						allowed[key] = true
					}
				}
			}
		}
	}

	for key := range dataObj {
		if !allowed[key] {
			warnings = append(warnings, ValidationError{
				InstancePath: path,
				Keyword:      "additionalProperties",
				Message:      fmt.Sprintf("additional property \"%s\" exists", key),
			})
		}
	}

	if props, ok := schemaObj["properties"].(map[string]interface{}); ok {
		for prop, propSchema := range props {
			if propData, ok := dataObj[prop]; ok {
				if _, isObj := propData.(map[string]interface{}); isObj {
					childPath := fmt.Sprintf("%s/%s", path, prop)
					warnings = append(warnings, collectAdditionalProps(propSchema, propData, childPath)...)
				}
			}
		}
	}

	return warnings
}
