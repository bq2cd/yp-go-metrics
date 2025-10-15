package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestField_compare(t *testing.T) {
	type fields struct {
		Key   string
		Type  FieldType
		Value any
	}
	type args struct {
		b Field
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "key, type different -> less",
			fields: fields{Key: "k1", Type: FieldTypeFloat, Value: 1.23},
			args:   args{b: Int("k2", 123)},
			want:   -1, // k1 < k2
		},
		{
			name:   "key, type different -> greater",
			fields: fields{Key: "k3", Type: FieldTypeFloat, Value: 1.23},
			args:   args{b: Int("k2", 123)},
			want:   1, // k3 > k2
		},
		{
			name:   "key equal, type different -> less",
			fields: fields{Key: "k1", Type: FieldTypeFloat, Value: 1.23},
			args:   args{b: Str("k1", "bigger")},
			want:   -1, // k1 == k2, float < str
		},
		{
			name:   "key equal, type different -> greater",
			fields: fields{Key: "k1", Type: FieldTypeFloat, Value: 1.23},
			args:   args{b: Int("k1", 123)},
			want:   1, // k1 == k2, float > int
		},
		{
			name:   "key, type equal; value different",
			fields: fields{Key: "k1", Type: FieldTypeInt, Value: 456},
			args:   args{b: Int("k1", 123)},
			want:   0, // k1 == k2, int == int
		},
		{
			name:   "key, type, value equal",
			fields: fields{Key: "k1", Type: FieldTypeInt, Value: 123},
			args:   args{b: Int("k1", 123)},
			want:   0, // k1 == k2, int == int
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Field{
				Key:   tt.fields.Key,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
			}
			assert.Equal(t, tt.want, a.compare(tt.args.b))
		})
	}
}

func TestField_equals(t *testing.T) {
	type fields struct {
		Key   string
		Type  FieldType
		Value any
	}
	type args struct {
		b Field
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "different key -> false",
			fields: fields{Key: "k1", Type: FieldTypeInt, Value: 123},
			args:   args{b: Int("k2", 213)},
			want:   false,
		},
		{
			name:   "same key, different type -> false",
			fields: fields{Key: "k1", Type: FieldTypeInt, Value: 123},
			args:   args{b: Float("k1", 2.13)},
			want:   false,
		},
		{
			name:   "same key, type; different value -> false",
			fields: fields{Key: "k1", Type: FieldTypeInt, Value: 123},
			args:   args{b: Int("k1", 213)},
			want:   false,
		},
		{
			name:   "same key, type; different compound value -> false",
			fields: fields{Key: "k1", Type: FieldTypeAny, Value: map[string]string{"a1": "b1"}},
			args:   args{b: Any("k1", []string{"a1", "b1"})},
			want:   false,
		},
		{
			name:   "same key, type; different compound value 2 -> false",
			fields: fields{Key: "k1", Type: FieldTypeAny, Value: []string{"b1", "a1"}},
			args:   args{b: Any("k1", []string{"a1", "b1"})},
			want:   false,
		},
		{
			name:   "same key, type, simple value -> true",
			fields: fields{Key: "k1", Type: FieldTypeStr, Value: "yes please"},
			args:   args{b: Str("k1", "yes please")},
			want:   true,
		},
		{
			name:   "same key, type, compound value -> true",
			fields: fields{Key: "k1", Type: FieldTypeAny, Value: map[string]string{"a1": "b1"}},
			args:   args{b: Any("k1", map[string]string{"a1": "b1"})},
			want:   true,
		},
		{
			name:   "same key, type, compound value 2 -> true",
			fields: fields{Key: "k1", Type: FieldTypeAny, Value: []string{"a1", "b1"}},
			args:   args{b: Any("k1", []string{"a1", "b1"})},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Field{
				Key:   tt.fields.Key,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
			}
			assert.Equal(t, tt.want, a.equals(tt.args.b))
		})
	}
}

func TestFieldSet_sorted(t *testing.T) {
	tests := []struct {
		name string
		fs   FieldSet
		want FieldSet
	}{
		{
			name: "simple sort",
			fs: FieldSet{
				Int("k3", 333),
				Float("k2", 22.22),
				Str("k1", "a value"),
				Float("k5", 5.55),
				Int("k2", 2222),
			},
			want: FieldSet{
				Str("k1", "a value"),
				Int("k2", 2222),
				Float("k2", 22.22),
				Int("k3", 333),
				Float("k5", 5.55),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fs.sorted())
		})
	}
}

func TestFieldSet_containsSubset(t *testing.T) {
	type args struct {
		subset FieldSet
	}
	tests := []struct {
		name string
		fs   FieldSet
		args args
		want bool
	}{
		{
			name: "empty set, subset",
			fs:   FieldSet{},
			args: args{
				subset: FieldSet{},
			},
			want: true,
		},
		{
			name: "empty set",
			fs:   FieldSet{},
			args: args{
				subset: FieldSet{Int("i1", 123)},
			},
			want: false,
		},
		{
			name: "empty subset",
			fs: FieldSet{
				Int("i1", 123),
			},
			args: args{
				subset: FieldSet{},
			},
			want: true,
		},
		{
			name: "no match",
			fs: FieldSet{
				Int("i1", 123),
			},
			args: args{
				subset: FieldSet{
					Int("i2", 456),
				},
			},
			want: false,
		},
		{
			name: "all fields, same order",
			fs: FieldSet{
				Int("i1", 123),
				Float("f2", 4.56),
			},
			args: args{
				subset: FieldSet{
					Int("i1", 123),
					Float("f2", 4.56),
				},
			},
			want: true,
		},
		{
			name: "all fields, different order",
			fs: FieldSet{
				Int("i1", 123),
				Str("x5", "hellow"),
				Float("f2", 4.56),
			},
			args: args{
				subset: FieldSet{
					Float("f2", 4.56),
					Int("i1", 123),
					Str("x5", "hellow"),
				},
			},
			want: true,
		},
		{
			name: "some fields 1",
			fs: FieldSet{
				Int("i1", 123),
				Str("x5", "hellow"),
				Float("f2", 4.56),
				Int("a2", -89),
			},
			args: args{
				subset: FieldSet{
					Float("f2", 4.56),
					Str("x5", "hellow"),
				},
			},
			want: true,
		},
		{
			name: "some fields 2",
			fs: FieldSet{
				Int("i1", 123),
				Str("x5", "hellow"),
				Float("f2", 4.56),
				Int("a2", -89),
				Any("g9", []string{"a1", "b1", "c1"}),
			},
			args: args{
				subset: FieldSet{
					Any("g9", []string{"a1", "b1", "c1"}),
					Str("x5", "hellow"),
				},
			},
			want: true,
		},
		{
			name: "some fields, slice mismatch",
			fs: FieldSet{
				Int("i1", 123),
				Str("x5", "hellow"),
				Float("f2", 4.56),
				Int("a2", -89),
				Any("g9", []string{"a1", "b1", "c1"}),
			},
			args: args{
				subset: FieldSet{
					Any("g9", []string{"c1", "a1", "b1"}),
					Str("x5", "hellow"),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fs.containsSubset(tt.args.subset))
		})
	}
}

func TestFieldSet_GetFieldByKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		fs   FieldSet
		args args
		want *Field
	}{
		{
			name: "empty set",
			fs:   FieldSet{},
			args: args{key: "k1"},
			want: nil,
		},
		{
			name: "key mismatch",
			fs: FieldSet{
				Int("i1", 123),
				Float("f2", 1.23),
			},
			args: args{key: "k1"},
			want: nil,
		},
		{
			name: "found",
			fs: FieldSet{
				Int("i1", 123),
				Float("f2", 1.23),
				Str("k1", "a needle"),
				Err("e1", errors.New("hiding here")),
			},
			args: args{key: "k1"},
			want: &Field{Key: "k1", Type: FieldTypeStr, Value: "a needle"},
		},
		{
			name: "last key wins",
			fs: FieldSet{
				Int("i1", 123),
				Float("f2", 1.23),
				Str("k1", "a needle"),
				Float("f3", 4.56),
				Str("k1", "another one"),
				Float("f5", -3.21),
				Int("k1", -19),
				Str("x3", "surprise"),
			},
			args: args{key: "k1"},
			want: &Field{Key: "k1", Type: FieldTypeInt, Value: -19},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fs.GetFieldByKey(tt.args.key))
		})
	}
}

func TestFieldSet_ToMap(t *testing.T) {
	tests := []struct {
		name string
		fs   FieldSet
		want map[string]Field
	}{
		{
			name: "empty set",
			fs:   FieldSet{},
			want: map[string]Field{},
		},
		{
			name: "unique fields",
			fs: FieldSet{
				Int("i1", 123),
				Str("k3", "why not"),
				Bool("b4", true),
			},
			want: map[string]Field{
				"i1": Int("i1", 123),
				"k3": Str("k3", "why not"),
				"b4": Bool("b4", true),
			},
		},
		{
			name: "duplicate fields",
			fs: FieldSet{
				Int("i1", 123),
				Str("k3", "why not"),
				Bool("b4", true),
				Int("i5", 543),
				Str("k10", "exactly"),
				Bool("b4", false),
				Float("ff7", -79),
			},
			want: map[string]Field{
				"i1":  Int("i1", 123),
				"k3":  Str("k3", "why not"),
				"i5":  Int("i5", 543),
				"k10": Str("k10", "exactly"),
				"b4":  Bool("b4", false),
				"ff7": Float("ff7", -79),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fs.ToMap())
		})
	}
}
