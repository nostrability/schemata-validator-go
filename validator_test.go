package validator

import (
	"encoding/json"
	"strings"
	"testing"
)

func hex64(c byte) string  { return strings.Repeat(string(c), 64) }
func sig128(c byte) string { return strings.Repeat(string(c), 128) }

func TestValidKind1Note(t *testing.T) {
	event, _ := json.Marshal(map[string]interface{}{
		"id": hex64('a'), "pubkey": hex64('b'), "created_at": 1700000000,
		"kind": 1, "tags": []interface{}{}, "content": "hello world", "sig": sig128('c'),
	})
	result := ValidateNote(event)
	if !result.Valid {
		t.Fatalf("Expected valid, errors: %+v", result.Errors)
	}
}

func TestWrongKindFails(t *testing.T) {
	event, _ := json.Marshal(map[string]interface{}{
		"id": hex64('a'), "pubkey": hex64('b'), "created_at": 1700000000,
		"kind": 1, "tags": []interface{}{}, "content": "hello", "sig": sig128('c'),
	})
	schema, ok := GetSchema("kind0Schema")
	if !ok {
		t.Fatal("kind0Schema not found")
	}
	result := Validate(schema, event)
	if result.Valid {
		t.Fatal("Expected invalid - kind mismatch")
	}
}

func TestMissingPubkeyFails(t *testing.T) {
	event := []byte(`{"kind":1,"content":"hello"}`)
	result := ValidateNote(event)
	if result.Valid {
		t.Fatal("Expected invalid - missing fields")
	}
}

func TestPTagValidation(t *testing.T) {
	schema, ok := GetSchema("pTagSchema")
	if !ok {
		t.Fatal("pTagSchema not found")
	}
	tag, _ := json.Marshal([]string{"p", hex64('a')})
	result := Validate(schema, tag)
	if !result.Valid {
		t.Fatalf("Expected valid p tag, errors: %+v", result.Errors)
	}
}

func TestNip11Validation(t *testing.T) {
	doc, _ := json.Marshal(map[string]interface{}{
		"name": "Test Relay", "supported_nips": []int{1, 11},
	})
	result := ValidateNip11(doc)
	if !result.Valid {
		t.Fatalf("Expected valid NIP-11, errors: %+v", result.Errors)
	}
}

func TestUnknownKindWarning(t *testing.T) {
	event, _ := json.Marshal(map[string]interface{}{
		"id": hex64('a'), "pubkey": hex64('b'), "created_at": 1700000000,
		"kind": 99999, "tags": []interface{}{}, "content": "", "sig": sig128('c'),
	})
	result := ValidateNote(event)
	if result.Valid {
		t.Fatal("Expected valid=false for unknown kind")
	}
	if len(result.Warnings) == 0 {
		t.Fatal("Expected warning for unknown kind")
	}
}

func TestValidateMessage(t *testing.T) {
	msg := []byte(`["NOTICE","rate limited"]`)
	result := ValidateMessage(msg, Relay, "Notice")
	if !result.Valid {
		t.Fatalf("Expected valid NOTICE, errors: %+v", result.Errors)
	}
}

func TestAdditionalPropsWarning(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	data := json.RawMessage(`{"name":"Alice","extra":"surprise"}`)
	result := Validate(schema, data)
	if !result.Valid {
		t.Fatal("Should be valid - extra props are warnings")
	}
	if len(result.Warnings) == 0 {
		t.Fatal("Should have warnings for extra props")
	}
}
