package ezcapsolver

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// This file holds the two things encoding/json cannot express on its own, and
// which every task and solution model needs:
//
//   - flattening a pass-through map into the same object level as the declared
//     fields, on the way out;
//   - collecting the keys a model does not declare into that same map, on the
//     way in.
//
// Both are driven by reflection over the json struct tags, so a model only has
// to declare its fields and a three-line MarshalJSON or UnmarshalJSON.
//
// Tag conventions used throughout the models:
//
//   - `json:"name"`          always serialized, even at its zero value;
//   - `json:"name,omitzero"` optional, omitted when unset;
//   - `json:"-"`             the Extra map itself, which is flattened instead;
//   - `ezcapsolver:"required"` decoding fails if the key is absent.
//
// omitzero rather than omitempty is deliberate: omitempty would also drop an
// explicitly empty map or slice, and the service distinguishes "field omitted"
// from "field present and empty" for several parameters.

// Ptr returns a pointer to v, for the optional task fields whose service-side
// default is not the Go zero value.
//
// Those fields have to be pointers: they carry omitzero so that leaving them
// unset lets the service apply its own default, and omitzero drops a false or a
// zero just as readily as an unset field. A pointer is what separates "not set"
// from "set to the zero value". Go has no address-of for a literal, hence this.
//
//	IsInvisible: ezcapsolver.Ptr(false)
func Ptr[T any](v T) *T { return &v }

// requiredTag marks a solution field the worker is known to always return.
const requiredTag = "required"

// typeFields caches what reflection found out about one struct type.
type typeFields struct {
	// declared holds every JSON key the type serializes.
	declared map[string]bool
	// required holds the JSON keys whose absence is a decoding failure.
	required []string
}

// fieldCache maps a reflect.Type to its typeFields. Models are a fixed set, so
// this fills once and is read-only afterwards.
var fieldCache sync.Map

// fieldsOf reflects over a struct type's json tags, memoising the result.
func fieldsOf(structType reflect.Type) *typeFields {
	if cached, ok := fieldCache.Load(structType); ok {
		return cached.(*typeFields)
	}
	fields := &typeFields{declared: make(map[string]bool)}
	collectFields(structType, fields)
	actual, _ := fieldCache.LoadOrStore(structType, fields)
	return actual.(*typeFields)
}

// collectFields walks one struct level, descending into embedded structs.
func collectFields(structType reflect.Type, into *typeFields) {
	if structType.Kind() != reflect.Struct {
		return
	}
	for index := range structType.NumField() {
		field := structType.Field(index)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == "-" && tag != "-," {
			// `json:"-"` opts the field out entirely; `json:"-,"` names it "-".
			continue
		}
		if field.Anonymous && name == "" {
			collectFields(field.Type, into)
			continue
		}
		if name == "" {
			name = field.Name
		}
		into.declared[name] = true
		if field.Tag.Get("ezcapsolver") == requiredTag {
			into.required = append(into.required, name)
		}
	}
}

// marshalWithExtra serializes value and merges extra into the same object level.
//
// A declared field wins over a pass-through entry of the same name: an
// accidental duplicate in caller-supplied data must not silently replace a
// modelled field. The losing entry is dropped without a diagnostic, which is
// what keeps this layer free of logging and the models pure data.
func marshalWithExtra(value any, extra map[string]any) ([]byte, error) {
	declared, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(extra) == 0 {
		return declared, nil
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(declared, &merged); err != nil {
		return nil, err
	}
	for key, item := range extra {
		if _, taken := merged[key]; taken {
			continue
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("extra field %q is not serializable: %w", key, err)
		}
		merged[key] = encoded
	}
	return json.Marshal(merged)
}

// unmarshalWithExtra decodes data into target and routes every key target does
// not declare into extra.
//
// Dropping an unknown key would lose it: workers ship faster than the SDK, and a
// field added on their side has to stay reachable without an SDK upgrade.
//
// target must be a pointer to a struct whose type does not itself implement
// json.Unmarshaler, otherwise decoding recurses; models pass a local alias type
// for exactly this reason.
func unmarshalWithExtra(data []byte, target any, extra *map[string]any) error {
	// The service always sends a solution as a JSON object. Anything else is a
	// contract break, and this rejects it with the offending value in hand
	// rather than silently producing a zero-valued model.
	var present map[string]json.RawMessage
	if err := json.Unmarshal(data, &present); err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}

	fields := fieldsOf(reflect.TypeOf(target).Elem())
	for _, name := range fields.required {
		if _, ok := present[name]; !ok {
			return fmt.Errorf("missing required field %q", name)
		}
	}

	// max: a model can declare more fields than the response carried, and the
	// spec says make with a negative size panics — gc clamps it today, but that
	// is not something to rely on for a hint that costs nothing to get right.
	rest := make(map[string]any, max(0, len(present)-len(fields.declared)))
	for name, raw := range present {
		if fields.declared[name] {
			continue
		}
		var item any
		if err := json.Unmarshal(raw, &item); err != nil {
			return fmt.Errorf("extra field %q is not decodable: %w", name, err)
		}
		rest[name] = item
	}
	*extra = rest
	return nil
}
