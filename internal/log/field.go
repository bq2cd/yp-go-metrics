package log

import (
	"cmp"
	"reflect"
	"slices"
	"sort"
)

// FieldType encodes value type inside the Field.
type FieldType int8

// Field encapsulates a key-value pair of a log event.
type Field struct {
	Key   string
	Type  FieldType
	Value any
}

// FieldSet represents a collection of fields.
type FieldSet []Field

// compare mimicks `cmp.Compare` but only takes into
// account field key and type.
// This funcionn is primarily used for sorting.
func (a Field) compare(b Field) int {
	r := cmp.Compare(a.Key, b.Key)
	if r == 0 {
		r = cmp.Compare(a.Type, b.Type)
	}
	return r
}

// equals invokes `reflect.DeepEqual` to compare field values.
// Beware of the performance implications.
func (a Field) equals(b Field) bool {
	if a.Key != b.Key {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	return reflect.DeepEqual(a.Value, b.Value)
}

// sorted returns a sorted copy of the fields.
func (fs FieldSet) sorted() FieldSet {
	return slices.SortedStableFunc(slices.Values(fs), func(a Field, b Field) int {
		return a.compare(b)
	})

}

// containsSubset returns true if all fields from a subset are
// present in the current field set.
// This function heavily relies on sorting and reflection, use it in
// tests only.
func (fs FieldSet) containsSubset(subset FieldSet) bool {
	if len(fs) == 0 {
		return len(subset) == 0
	}
	haystack := fs.sorted()
	for _, field := range subset {
		j, ok := sort.Find(len(haystack), func(i int) int {
			return field.compare(haystack[i])
		})
		if !ok {
			return false
		}
		if j >= len(haystack) || !field.equals(haystack[j]) {
			return false
		}
	}
	return true
}

// GetFieldByKey returns a pointer to a field by its key if current field set contains the key,
// and `nil` otherwise.
// In case there are multiple fields with the same key, the last field wins.
func (fs FieldSet) GetFieldByKey(key string) *Field {
	var found *Field
	for _, f := range fs {
		if f.Key == key {
			found = &f
		}
	}
	return found
}

// ToMap creates a map keyed by a field key, with value being the field itself.
// If multiple fields with the same key exist, the last added wins.
func (fs FieldSet) ToMap() map[string]Field {
	fieldMap := make(map[string]Field, len(fs))
	for _, f := range fs {
		fieldMap[f.Key] = f
	}
	return fieldMap
}
