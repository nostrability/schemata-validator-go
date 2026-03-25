# schemata-validator-go

[![Test](https://github.com/nostrability/schemata-validator-go/actions/workflows/test.yml/badge.svg)](https://github.com/nostrability/schemata-validator-go/actions/workflows/test.yml)
[![License](https://img.shields.io/badge/license-GPL--3.0--or--later-blue?style=flat-square)](LICENSE)

Go validator for [Nostr](https://nostr.com/) protocol JSON schemas. Built on [`schemata-go`](https://github.com/nostrability/schemata-go) and [`santhosh-tekuri/jsonschema`](https://github.com/santhosh-tekuri/jsonschema) (Draft 7).

## Overview

`schemata-validator-go` wraps the `schemata-go` embedded JSON Schema definitions with jsonschema validation, exposing ready-to-use validation functions for common Nostr data structures. It validates Nostr events by kind, NIP-11 relay information documents, and relay/client protocol messages.

Validation results include both hard errors (schema violations) and soft warnings (additional properties not defined in the schema).

## When to use this

JSON Schema validation is [not suited for runtime hot paths](https://github.com/nostrability/schemata#what-is-it-not-good-for). Use this in:

- **CI pipelines** catching schema drift during builds
- **Integration tests** for clients and relays
- **Fuzz testing** to identify malformed event structures

## Installation

```bash
go get github.com/nostrability/schemata-validator-go
```

Since `schemata-go` is not published to a Go module proxy, you'll need a replace directive:

```bash
git clone https://github.com/nostrability/schemata-go.git ../schemata-go
go mod edit -replace github.com/nostrability/schemata-go=../schemata-go
go mod tidy
```

## Quick Start

```go
package myapp_test

import (
	"encoding/json"
	"testing"

	validator "github.com/nostrability/schemata-validator-go"
)

func TestMyEvent(t *testing.T) {
	event, _ := json.Marshal(map[string]interface{}{
		"id":         strings.Repeat("a", 64),
		"pubkey":     strings.Repeat("b", 64),
		"created_at": 1700000000,
		"kind":       1,
		"tags":       []interface{}{},
		"content":    "hello world",
		"sig":        strings.Repeat("c", 128),
	})

	result := validator.ValidateNote(event)
	if !result.Valid {
		t.Fatalf("errors: %+v", result.Errors)
	}
	// result.Warnings may flag additional properties
}
```

## API

### `Validate(schema, data)`

```go
func Validate(schema json.RawMessage, data json.RawMessage) ValidationResult
```

Low-level validator. Compiles a JSON Schema (Draft 7) and validates `data` against it. Strips nested `$id` fields from the schema to prevent resolution issues. Use `ValidateNote`, `ValidateNip11`, or `ValidateMessage` for common cases.

| Parameter | Type | Description |
|-----------|------|-------------|
| `schema` | `json.RawMessage` | A JSON Schema document |
| `data` | `json.RawMessage` | The data to validate |

### `ValidateNote(event)`

```go
func ValidateNote(event json.RawMessage) ValidationResult
```

Validates a Nostr event against the schema for its `kind`. The schema is looked up from `schemata-go` using the key `kind{N}Schema`. Returns a warning (not an error) if no schema exists for the given kind.

| Parameter | Type | Description |
|-----------|------|-------------|
| `event` | `json.RawMessage` | A Nostr event as JSON bytes |

### `ValidateNip11(doc)`

```go
func ValidateNip11(doc json.RawMessage) ValidationResult
```

Validates a NIP-11 relay information document — the metadata object a relay serves at its HTTP endpoint — against the `nip11Schema`.

| Parameter | Type | Description |
|-----------|------|-------------|
| `doc` | `json.RawMessage` | A NIP-11 relay info document as JSON bytes |

### `ValidateMessage(msg, subject, slug)`

```go
func ValidateMessage(msg json.RawMessage, subject Subject, slug string) ValidationResult
```

Validates a Nostr protocol message against the schema for the given subject and message type. The schema key is constructed as `{subject}{Slug}Schema` (e.g., `relayNoticeSchema` for `subject=Relay`, `slug="Notice"`).

| Parameter | Type | Description |
|-----------|------|-------------|
| `msg` | `json.RawMessage` | The protocol message as JSON bytes |
| `subject` | `Subject` | Message origin: `Relay` or `Client` |
| `slug` | `string` | Message type name (e.g., `"Notice"`, `"Event"`, `"Ok"`) |

### `GetSchema(key)`

```go
func GetSchema(key string) (json.RawMessage, bool)
```

Looks up a schema by key from the `schemata-go` registry. Returns the schema and `true` if found, or `nil` and `false` if not.

| Parameter | Type | Description |
|-----------|------|-------------|
| `key` | `string` | Schema registry key (e.g., `"kind1Schema"`, `"pTagSchema"`) |

### `ValidationResult`

```go
type ValidationResult struct {
    Valid    bool
    Errors   []ValidationError
    Warnings []ValidationError
}
```

- `Valid` — `true` if the data passes all schema constraints
- `Errors` — schema violations; empty when `Valid` is `true`
- `Warnings` — additional property alerts; populated even when `Valid` is `true`

### `ValidationError`

```go
type ValidationError struct {
    InstancePath string
    Keyword      string
    Message      string
    SchemaPath   string
}
```

### `Subject`

```go
type Subject int

const (
    Relay  Subject = iota
    Client
)
```

## Usage Examples

**Event validation:**

```go
event, _ := json.Marshal(map[string]interface{}{
    "id": strings.Repeat("a", 64), "pubkey": strings.Repeat("b", 64),
    "created_at": 1700000000, "kind": 1, "tags": []interface{}{},
    "content": "hello world", "sig": strings.Repeat("c", 128),
})
result := validator.ValidateNote(event)
```

**NIP-11 validation:**

```go
doc, _ := json.Marshal(map[string]interface{}{
    "name": "My Relay", "supported_nips": []int{1, 11},
})
result := validator.ValidateNip11(doc)
```

**Protocol message validation:**

```go
msg := json.RawMessage(`["NOTICE", "rate limited"]`)
result := validator.ValidateMessage(msg, validator.Relay, "Notice")
```

**Direct schema lookup:**

```go
schema, ok := validator.GetSchema("pTagSchema")
if ok {
    tag, _ := json.Marshal([]string{"p", strings.Repeat("a", 64)})
    result := validator.Validate(schema, tag)
}
```

## Known Limitations

- **Partial kind coverage:** Only event kinds with a corresponding schema in `@nostrability/schemata` can be validated. `ValidateNote` returns a warning (not an error) when no schema exists for the given kind.
- **No recursive content validation:** The `content` field of events containing stringified JSON (e.g., kind 0 metadata) is not recursively validated.
- **Alpha accuracy:** False positives and negatives are possible. The underlying schemas are in active development.

## Related Packages

- [`schemata-go`](https://github.com/nostrability/schemata-go) — Go data package containing embedded schemas and registry
- [`@nostrability/schemata`](https://github.com/nostrability/schemata) — canonical language-agnostic schema definitions
- [`@nostrwatch/schemata-js-ajv`](https://github.com/sandwichfarm/nostr-watch/tree/next/libraries/schemata-js-ajv) — JavaScript/TypeScript validator implementation

## License

[GPL-3.0-or-later](LICENSE)
