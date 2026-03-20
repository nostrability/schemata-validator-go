package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	schemata "github.com/nostrability/schemata-go"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidationResult holds the outcome of a schema validation.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
}

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	InstancePath string
	Keyword      string
	Message      string
	SchemaPath   string
}

// Subject identifies the sender of a protocol message.
type Subject int

const (
	Relay Subject = iota
	Client
)

func (s Subject) String() string {
	if s == Relay {
		return "relay"
	}
	return "client"
}

// stripNestedIDs removes $id fields from nested objects, keeping root $id.
func stripNestedIDs(v interface{}, depth int) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		if depth > 0 {
			delete(val, "$id")
		}
		for k, child := range val {
			val[k] = stripNestedIDs(child, depth+1)
		}
		return val
	case []interface{}:
		for i, child := range val {
			val[i] = stripNestedIDs(child, depth+1)
		}
		return val
	default:
		return v
	}
}

// Validate validates data against a JSON schema.
func Validate(schema json.RawMessage, data json.RawMessage) ValidationResult {
	var schemaObj interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "compilation", Message: fmt.Sprintf("Schema parse error: %v", err)}},
		}
	}
	schemaObj = stripNestedIDs(schemaObj, 0)

	var dataObj interface{}
	if err := json.Unmarshal(data, &dataObj); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "compilation", Message: fmt.Sprintf("Data parse error: %v", err)}},
		}
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft7)
	if err := c.AddResource("schema.json", schemaObj); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "compilation", Message: fmt.Sprintf("Schema compilation error: %v", err)}},
		}
	}

	sch, err := c.Compile("schema.json")
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "compilation", Message: fmt.Sprintf("Schema compilation error: %v", err)}},
		}
	}

	var errors []ValidationError
	if err := sch.Validate(dataObj); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			errors = flattenErrors(ve)
		} else {
			errors = []ValidationError{{Keyword: "validation", Message: err.Error()}}
		}
	}

	var origSchema interface{}
	json.Unmarshal(schema, &origSchema)
	warnings := collectAdditionalProps(origSchema, dataObj, "")

	return ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

func flattenErrors(ve *jsonschema.ValidationError) []ValidationError {
	var result []ValidationError
	doFlatten(ve, &result)
	return result
}

func doFlatten(ve *jsonschema.ValidationError, result *[]ValidationError) {
	if len(ve.Causes) == 0 {
		keyword := ""
		if ve.ErrorKind != nil {
			kp := ve.ErrorKind.KeywordPath()
			if len(kp) > 0 {
				keyword = kp[len(kp)-1]
			}
		}
		instancePath := "/" + strings.Join(ve.InstanceLocation, "/")
		if len(ve.InstanceLocation) == 0 {
			instancePath = ""
		}
		*result = append(*result, ValidationError{
			InstancePath: instancePath,
			Keyword:      keyword,
			Message:      ve.Error(),
			SchemaPath:   ve.SchemaURL,
		})
	}
	for _, cause := range ve.Causes {
		doFlatten(cause, result)
	}
}

// ValidateNote validates a Nostr event by looking up kind{N}Schema.
func ValidateNote(event json.RawMessage) ValidationResult {
	var obj map[string]interface{}
	if err := json.Unmarshal(event, &obj); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "note", Message: "Event is not valid JSON"}},
		}
	}

	kindVal, ok := obj["kind"]
	if !ok {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "note", Message: "Event missing 'kind' field"}},
		}
	}

	kindFloat, ok := kindVal.(float64)
	if !ok {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "note", Message: "Event 'kind' field is not a number"}},
		}
	}

	key := fmt.Sprintf("kind%dSchema", int(kindFloat))
	schema, found := schemata.Get(key)
	if !found {
		return ValidationResult{
			Valid:    false,
			Warnings: []ValidationError{{Keyword: "note", Message: fmt.Sprintf("No schema found for kind %d", int(kindFloat))}},
		}
	}
	return Validate(schema, event)
}

// ValidateNip11 validates a NIP-11 relay information document.
func ValidateNip11(doc json.RawMessage) ValidationResult {
	schema, found := schemata.Get("nip11Schema")
	if !found {
		return ValidationResult{
			Valid:  false,
			Errors: []ValidationError{{Keyword: "nip11", Message: "nip11Schema not found in registry"}},
		}
	}
	return Validate(schema, doc)
}

// ValidateMessage validates a protocol message (relay or client).
func ValidateMessage(msg json.RawMessage, subject Subject, slug string) ValidationResult {
	key := fmt.Sprintf("%s%sSchema", subject, capitalize(strings.ToLower(slug)))
	schema, found := schemata.Get(key)
	if !found {
		return ValidationResult{
			Valid:    false,
			Warnings: []ValidationError{{Keyword: "message", Message: fmt.Sprintf("No schema found for %s %s", subject, slug)}},
		}
	}
	return Validate(schema, msg)
}

// GetSchema looks up a schema by key.
func GetSchema(key string) (json.RawMessage, bool) {
	return schemata.Get(key)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
