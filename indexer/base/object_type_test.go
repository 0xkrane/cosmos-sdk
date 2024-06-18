package indexerbase

import (
	"strings"
	"testing"
)

var object1Type = ObjectType{
	Name: "object1",
	KeyFields: []Field{
		{
			Name: "field1",
			Kind: StringKind,
		},
	},
}

var object2Type = ObjectType{
	KeyFields: []Field{
		{
			Name: "field1",
			Kind: StringKind,
		},
		{
			Name: "field2",
			Kind: Int32Kind,
		},
	},
}

var object3Type = ObjectType{
	Name: "object3",
	ValueFields: []Field{
		{
			Name: "field1",
			Kind: StringKind,
		},
		{
			Name: "field2",
			Kind: Int32Kind,
		},
	},
}

var object4Type = ObjectType{
	Name: "object4",
	KeyFields: []Field{
		{
			Name: "field1",
			Kind: Int32Kind,
		},
	},
	ValueFields: []Field{
		{
			Name: "field2",
			Kind: StringKind,
		},
	},
}

func TestObjectType_Validate(t *testing.T) {
	tests := []struct {
		name        string
		objectType  ObjectType
		errContains string
	}{
		{
			name:        "valid object type",
			objectType:  object1Type,
			errContains: "",
		},
		{
			name: "empty object type name",
			objectType: ObjectType{
				Name: "",
				KeyFields: []Field{
					{
						Name: "field1",
						Kind: StringKind,
					},
				},
			},
			errContains: "object type name cannot be empty",
		},
		{
			name: "invalid key field",
			objectType: ObjectType{
				Name: "object1",
				KeyFields: []Field{
					{
						Name: "",
						Kind: StringKind,
					},
				},
			},
			errContains: "field name cannot be empty",
		},
		{
			name: "invalid value field",
			objectType: ObjectType{
				Name: "object1",
				ValueFields: []Field{
					{
						Kind: StringKind,
					},
				},
			},
			errContains: "field name cannot be empty",
		},
		{
			name:        "no fields",
			objectType:  ObjectType{Name: "object0"},
			errContains: "has no key or value fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.objectType.Validate()
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestObjectType_ValidateKey(t *testing.T) {
	tests := []struct {
		name        string
		objectType  ObjectType
		key         interface{}
		errContains string
	}{
		{
			name: "no key fields",
			objectType: ObjectType{
				Name: "object1",
			},
			key: nil,
		},
		{
			name:        "single key field, valid",
			objectType:  object1Type,
			key:         "hello",
			errContains: "",
		},
		{
			name:        "single key field, invalid",
			objectType:  object1Type,
			key:         []interface{}{"value"},
			errContains: "invalid value",
		},
		{
			name:       "multiple key fields, valid",
			objectType: object2Type,
			key:        []interface{}{"hello", int32(42)},
		},
		{
			name:        "multiple key fields, not a slice",
			objectType:  object2Type,
			key:         map[string]interface{}{"field1": "hello", "field2": "42"},
			errContains: "expected slice of values",
		},
		{
			name:        "multiple key fields, wrong number of values",
			objectType:  object2Type,
			key:         []interface{}{"hello"},
			errContains: "expected 2 key fields",
		},
		{
			name:        "multiple key fields, invalid value",
			objectType:  object2Type,
			key:         []interface{}{"hello", "abc"},
			errContains: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.objectType.ValidateKey(tt.key)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestObjectType_ValidateValue(t *testing.T) {
	tests := []struct {
		name        string
		objectType  ObjectType
		value       interface{}
		errContains string
	}{
		{
			name:       "no value fields",
			objectType: ObjectType{Name: "object0"},
			value:      nil,
		},
		{
			name: "single value field, valid",
			objectType: ObjectType{
				ValueFields: []Field{
					{
						Name: "field1",
						Kind: StringKind,
					},
				},
			},
			value:       "hello",
			errContains: "",
		},
		{
			name:       "value updates, empty",
			objectType: object3Type,
			value:      MapValueUpdates(map[string]interface{}{}),
		},
		{
			name:       "value updates, 1 field valid",
			objectType: object3Type,
			value: MapValueUpdates(map[string]interface{}{
				"field1": "hello",
			}),
		},
		{
			name:       "value updates, 2 fields, 1 invalid",
			objectType: object3Type,
			value: MapValueUpdates(map[string]interface{}{
				"field1": "hello",
				"field2": "abc",
			}),
			errContains: "expected int32",
		},
		{
			name:       "value updates, extra value",
			objectType: object3Type,
			value: MapValueUpdates(map[string]interface{}{
				"field1": "hello",
				"field2": int32(42),
				"field3": "extra",
			}),
			errContains: "unexpected values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.objectType.ValidateValue(tt.value)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestObjectType_ValidateObjectUpdate(t *testing.T) {
	tests := []struct {
		name        string
		objectType  ObjectType
		object      ObjectUpdate
		errContains string
	}{
		{
			name:       "wrong name",
			objectType: object1Type,
			object: ObjectUpdate{
				TypeName: "object2",
				Key:      "hello",
			},
			errContains: "does not match update type name",
		},
		{
			name:       "invalid value",
			objectType: object1Type,
			object: ObjectUpdate{
				TypeName: "object1",
				Key:      123,
			},
			errContains: "invalid value",
		},
		{
			name:       "valid update",
			objectType: object4Type,
			object: ObjectUpdate{
				TypeName: "object4",
				Key:      int32(123),
				Value:    "hello",
			},
		},
		{
			name:       "valid deletion",
			objectType: object4Type,
			object: ObjectUpdate{
				TypeName: "object4",
				Key:      int32(123),
				Value:    "ignored!",
				Delete:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.objectType.ValidateObjectUpdate(tt.object)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}
