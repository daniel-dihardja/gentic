package gentic

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// SchemaFor returns a JSON Schema object (as bytes) for T, suitable for OpenAI
// structured outputs with strict mode: every property is listed in required,
// additionalProperties is false on objects, and pointer fields use nullable types.
// T must be a struct (or pointer-to-struct; the struct definition is used).
func SchemaFor[T any]() json.RawMessage {
	var z T
	return SchemaFromStruct(z)
}

// SchemaFromStruct returns JSON Schema for a struct sample value (same rules as [SchemaFor]).
// v must be a struct or a pointer-to-struct (the struct type is used).
func SchemaFromStruct(v any) json.RawMessage {
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		panic("gentic.SchemaFromStruct: v must be a struct or pointer to struct")
	}
	raw, err := schemaJSONFromType(t)
	if err != nil {
		panic("gentic.SchemaFromStruct: " + err.Error())
	}
	return raw
}

// SchemaFromValue builds JSON Schema for the type pointed to by v.
// v must be a non-nil pointer to a struct (e.g. &MyStruct{}).
func SchemaFromValue(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, fmt.Errorf("gentic: nil value")
	}
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("gentic: expected pointer, got %T", v)
	}
	if t.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("gentic: expected pointer to struct, got %T", v)
	}
	return schemaJSONFromType(t.Elem())
}

// SchemaTitleFromValue returns a stable name for response_format.json_schema.name.
func SchemaTitleFromValue(v any) string {
	if v == nil {
		return "response"
	}
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil {
		return "response"
	}
	if t.Name() != "" {
		return sanitizeSchemaName(t.Name())
	}
	return "response"
}

func sanitizeSchemaName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "response"
	}
	return out
}

func schemaJSONFromType(t reflect.Type) (json.RawMessage, error) {
	obj, err := buildObjectSchema(t, map[reflect.Type]struct{}{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func buildObjectSchema(t reflect.Type, seen map[reflect.Type]struct{}) (map[string]any, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}
	if _, dup := seen[t]; dup {
		return nil, fmt.Errorf("recursive struct %s is not supported", t.String())
	}
	seen[t] = struct{}{}
	defer delete(seen, t)

	properties := make(map[string]any)
	// Use non-nil slice so JSON encodes [] not null; OpenAI tool parameters require
	// "required" to be an array (empty structs e.g. parameter-less tools).
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name, skip := jsonNameFromField(f)
		if skip {
			continue
		}

		prop, err := fieldSchema(f.Type, seen)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", f.Name, err)
		}
		properties[name] = prop
		required = append(required, name)
	}

	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
		"required":             required,
	}, nil
}

func jsonNameFromField(f reflect.StructField) (name string, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name, false
	}
	if tag == "-" {
		return "", true
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return f.Name, false
	}
	if parts[0] == "-" {
		return "", true
	}
	return parts[0], false
}

func fieldSchema(t reflect.Type, seen map[reflect.Type]struct{}) (map[string]any, error) {
	// Pointer → nullable JSON types for OpenAI strict mode
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		inner, err := scalarOrObjectSchema(elem, seen)
		if err != nil {
			return nil, err
		}
		return nullableSchema(inner), nil
	}
	return scalarOrObjectSchema(t, seen)
}

func nullableSchema(inner map[string]any) map[string]any {
	// Nullable objects cannot use type: ["object","null"] with inline properties in all providers;
	// use anyOf with explicit null branch.
	if _, isObj := inner["properties"]; isObj {
		return map[string]any{
			"anyOf": []any{inner, map[string]any{"type": "null"}},
		}
	}
	typ, ok := inner["type"]
	if !ok {
		return map[string]any{
			"anyOf": []any{inner, map[string]any{"type": "null"}},
		}
	}
	switch v := typ.(type) {
	case string:
		return map[string]any{"type": []any{v, "null"}}
	case []any:
		out := append([]any(nil), v...)
		out = append(out, "null")
		return map[string]any{"type": out}
	default:
		return map[string]any{
			"anyOf": []any{inner, map[string]any{"type": "null"}},
		}
	}
}

func scalarOrObjectSchema(t reflect.Type, seen map[reflect.Type]struct{}) (map[string]any, error) {
	switch t.Kind() {
	case reflect.Struct:
		return buildObjectSchema(t, seen)
	case reflect.Slice, reflect.Array:
		elem := t.Elem()
		item, err := fieldSchema(elem, seen)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"type":  "array",
			"items": item,
		}, nil
	case reflect.Map:
		return map[string]any{
			"type": "object",
		}, nil
	case reflect.String:
		return map[string]any{"type": "string"}, nil
	case reflect.Bool:
		return map[string]any{"type": "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}, nil
	default:
		return nil, fmt.Errorf("unsupported kind %s", t.Kind())
	}
}
