package log

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
