# schemata-validator-go

[![Test](https://github.com/nostrability/schemata-validator-go/actions/workflows/test.yml/badge.svg)](https://github.com/nostrability/schemata-validator-go/actions/workflows/test.yml)

Go validator for [Nostr](https://nostr.com/) protocol JSON schemas. Built on top of [`schemata-go`](https://github.com/nostrability/schemata-go).

## When to use this

JSON Schema validation is [not suited for runtime hot paths](https://github.com/nostrability/schemata#what-is-it-not-good-for). Use this in **CI and integration tests**.

## Usage

```go
import validator "github.com/nostrability/schemata-validator-go"

func TestMyEvent(t *testing.T) {
    event := []byte(`{"id":"aa...","pubkey":"bb...","created_at":1700000000,"kind":1,"tags":[],"content":"hello","sig":"cc..."}`)
    result := validator.ValidateNote(event)
    if !result.Valid {
        t.Fatalf("errors: %+v", result.Errors)
    }
}
```

## API

| Function | Description |
|----------|-------------|
| `Validate(schema, data)` | Validate data against any JSON schema |
| `ValidateNote(event)` | Validate a Nostr event (looks up `kind{N}Schema`) |
| `ValidateNip11(doc)` | Validate a NIP-11 relay info document |
| `ValidateMessage(msg, subject, slug)` | Validate a relay/client protocol message |
| `GetSchema(key)` | Look up a schema from the registry |

## License

GPL-3.0-or-later
