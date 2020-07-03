package goexpr

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Tests NewExprValueX(), NewNilExprValue(),
func TestValue_Json(t *testing.T) {
	tests := []struct {
		name string
		rv   Value
		json string
	}{
		{"nilBoolean", NewNilExprValue(NewScalarTypeSignature(VTBoolean)), `
{"type":{"base_type":"boolean"}}`},
		{"nilInteger", NewNilExprValue(NewScalarTypeSignature(VTInteger)), `
{"type":{"base_type":"integer"}}`},
		{"nilList", NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))), `
{"type":{"base_type":"list","unit_type":{"base_type":"string"}}}`},
		{"nilMap", NewNilExprValue(NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString))), `
{"type":{"base_type":"map","unit_type":{"base_type":"string"}}}`},
		{"nilRegexp", NewNilExprValue(NewScalarTypeSignature(VTRegexp)), `
{"type":{"base_type":"regexp"}}`},
		{"nilString", NewNilExprValue(NewScalarTypeSignature(VTString)), `
{"type":{"base_type":"string"}}`},

		{"boolean", NewExprValueBoolean(true), `
{"type":{"base_type":"boolean"},"value":true}`},
		{"integer", NewExprValueInteger(3), `
{"type":{"base_type":"integer"},"value":3}`},
		{"list", NewExprValueList(NewScalarTypeSignature(VTString),
			[]Value{
				NewExprValueMust(NewScalarTypeSignature(VTString), "string 1"),
				NewExprValueMust(NewScalarTypeSignature(VTString), "string 2"),
			}), `
{"type":{"base_type":"list","unit_type":{"base_type":"string"}},"value":[
{"type":{"base_type":"string"},"value":"string 1"},
{"type":{"base_type":"string"},"value":"string 2"}]}`},
		{"map", NewExprValueMap(NewScalarTypeSignature(VTString),
			map[string]Value{
				"one": NewExprValueString("string 1"),
				"two": NewExprValueString("string 2"),
			}), `
{"type":{"base_type":"map","unit_type":{"base_type":"string"}},"value":{
"one":{"type":{"base_type":"string"},"value":"string 1"},
"two":{"type":{"base_type":"string"},"value":"string 2"}}}`},
		{"regexp", NewExprValueRegexpMust("[0-9]{3}"), `
{"type":{"base_type":"regexp"},"value":"[0-9]{3}"}`},
		{"string", NewExprValueString("a string"), `
{"type":{"base_type":"string"},"value":"a string"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			js, err := json.Marshal(test.rv)
			if err != nil {
				t.Errorf("unexpected error json marshal value %v: %v", test.rv, err)
				return
			}
			// remove newlines from expected
			expected := strings.Replace(test.json, "\n", "", -1)
			if string(js) != expected {
				t.Errorf("invalid json\nactual:   %s\nexpected: %s", string(js), expected)
				return
			}

			var rv Value
			err = json.Unmarshal(js, &rv)
			if err != nil {
				t.Errorf("unexpected error json unmarshal json %v: %v", string(js), err)
				return
			}
			if !rv.Equal(test.rv) {
				t.Errorf("invalid values\nactual:   %v\nexpected: %v", rv, test.rv)
			}
		})
	}
}

func TestValue_Equal(t *testing.T) {
	tests := []struct {
		name   string
		rv1    Value
		rv2    Value
		result bool
	}{
		{"nilNil",
			NewNilExprValue(NewScalarTypeSignature(VTBoolean)), NewNilExprValue(NewScalarTypeSignature(VTBoolean)), true},
		{"nilNonNil",
			NewNilExprValue(NewScalarTypeSignature(VTBoolean)), NewExprValueBoolean(true), false},
		{"nonNilNil",
			NewExprValueBoolean(true), NewNilExprValue(NewScalarTypeSignature(VTBoolean)), false},

		{"boolTrue",
			NewExprValueBoolean(true), NewExprValueBoolean(true), true},
		{"boolFalse",
			NewExprValueBoolean(true), NewExprValueBoolean(false), false},
		{"boolFalseType",
			NewExprValueBoolean(true), NewExprValueString("false"), false},

		{"integerTrue",
			NewExprValueInteger(2), NewExprValueInteger(2), true},
		{"integerFalse",
			NewExprValueInteger(2), NewExprValueInteger(4), false},
		{"integerFalseType",
			NewExprValueInteger(2), NewExprValueString("2"), false},

		{"listTrue",
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}), true},
		{"listFalseUnitType",
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}),
			NewExprValueList(NewScalarTypeSignature(VTInteger), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}), false},
		{"listFalseLen",
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1")}), false},
		{"listValueType",
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueInteger(5)}), false},
		{"listFalseValueValue",
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("value2")}),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("value1"), NewExprValueString("another value")}), false},

		{"mapTrue",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}), true},
		{"mapFalseUnitType",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTInteger), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}), false},
		{"mapFalseMapLen",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1")}), false},
		{"mapFalseKey",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1":        NewExprValueString("value1"),
				"another key": NewExprValueString("value2")}), false},
		{"mapFalseValueType",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueInteger(2)}), false},
		{"mapFalseValueValue",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("value2")}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"key1": NewExprValueString("value1"),
				"key2": NewExprValueString("another value")}), false},

		{"regexpTrue",
			NewExprValueRegexpMust("[0-9]{3}"), NewExprValueRegexpMust("[0-9]{3}"), true},
		{"regexpFalse",
			NewExprValueRegexpMust("[0-9]{3}"), NewExprValueRegexpMust("another regexp"), false},
		{"regexpFalseType",
			NewExprValueRegexpMust("[0-9]{3}"), NewExprValueString("[0-9]{3}"), false},

		{"stringTrue",
			NewExprValueString("a string"), NewExprValueString("a string"), true},
		{"stringFalse",
			NewExprValueString("a string"), NewExprValueString("another string"), false},
		{"stringFalseType",
			NewExprValueString("a string"), NewExprValueInteger(5), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.rv1.Equal(test.rv2)
			if res != test.result {
				t.Errorf("invalid result (%v != %v)\nvalue1: %v\nvalue2: %v", res, test.result, test.rv1, test.rv2)
			}
		})
	}
}

func TestValue_Compare(t *testing.T) {
	tests := []struct {
		name   string
		rv1    Value
		rv2    Value
		pnic   bool
		result Value
	}{
		// Non-equal value types
		{"integerNonInteger",
			NewExprValueInteger(2), NewExprValueBoolean(true), true, EvNilInteger},
		{"integerNonIntegerNil",
			NewExprValueInteger(2), EvNilBoolean, true, EvNilInteger},
		{"nonIntegerInteger",
			NewExprValueBoolean(true), NewExprValueInteger(2), true, EvNilInteger},

		// Nil
		{"nilNil",
			EvNilInteger, EvNilInteger, false, EvNilInteger},
		{"nilNonNil",
			EvNilInteger, NewExprValueInteger(5), false, EvNilInteger},
		{"nonNilNil",
			NewExprValueInteger(5), EvNilInteger, false, EvNilInteger},

		{"integerEqual",
			NewExprValueInteger(2), NewExprValueInteger(2), false, NewExprValueInteger(0)},
		{"integerLess",
			NewExprValueInteger(2), NewExprValueInteger(3), false, NewExprValueInteger(-1)},
		{"integerGreater",
			NewExprValueInteger(3), NewExprValueInteger(2), false, NewExprValueInteger(1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				rec := recover()
				if rec != nil {
					if !test.pnic {
						t.Errorf("unexpected panic: %v", rec)
					}
				}
			}()
			res := test.rv1.Compare(test.rv2)
			if res != test.result {
				t.Errorf("invalid result (%v != %v)\nvalue1: %v\nvalue2: %v", res, test.result, test.rv1, test.rv2)
			}
		})
	}
}

func TestValue_String(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		str   string ``
	}{
		{"nil", EvNil, `<<nil>>`},
		{"boolean", NewExprValueBoolean(true), `true`},
		{"integer", NewExprValueInteger(5), `5`},
		{"list", NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("value1"), NewExprValueString("value2")}),
			`["value1","value2"]`},
		{"map", NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
			"key1": NewExprValueString("value1"),
			"key2": NewExprValueString("value2")}),
			`{key1:"value1",key2:"value2"}`},
		{"regexp", NewExprValueRegexpMust("[0-9]{3}"), `"[0-9]{3}"`},
		{"string", NewExprValueString("value"), `"value"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str := test.value.String()
			if str != test.str {
				t.Errorf("invalid string result.\nactual:   %v\nexpected: %v", str, test.str)
			}
		})
	}
}

func TestValue_SearchAll(t *testing.T) {
	tests := []struct {
		name   string
		rvTest Value
		rvKey  Value
		rvRes  []Value
		found  bool
	}{
		{"listFoundSingle", NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		}), NewExprValueString("bar"), []Value{
			NewExprValueString("bar"),
		}, true},
		{"listFoundMulti", NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		}), NewExprValueString("foo"), []Value{
			NewExprValueString("foo"), NewExprValueString("foo"),
		}, true},
		{"listNotFound", NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
		}), NewExprValueString("not found"), []Value{}, false},

		{"mapFound", NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
			"foo": NewExprValueString("foo1"),
			"bar": NewExprValueString("bar1"),
		}), NewExprValueString("foo"), []Value{
			NewExprValueString("foo1"),
		}, true},
		{"mapNotFound", NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
			"foo": NewExprValueString("foo1"),
			"bar": NewExprValueString("bar1"),
		}), NewExprValueString("not found"), []Value{}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, ok := test.rvTest.SearchAll(test.rvKey)

			if test.found {
				if !ok {
					t.Errorf("key %v unexpectedly not found in value %v", test.rvKey, test.rvTest)
					return
				}
				if !reflect.DeepEqual(res, test.rvRes) {
					t.Errorf("result not equal.\nactual:   %v\nexpected: %v", res, test.rvRes)
				}
			} else {
				if ok {
					t.Errorf("key %v unexpectedly found in value %v", test.rvKey, test.rvTest)
				}
			}
		})
	}
}

// TODO test reference and assign for map

func TestValue_Nil(t *testing.T) {
	tests := []struct {
		name  string
		rv    Value
		isNil bool
	}{
		{"booleanTrue", NewNilExprValue(NewScalarTypeSignature(VTBoolean)), true},
		{"booleanFalse", NewExprValueBoolean(true), false},
		{"integerTrue", NewNilExprValue(NewScalarTypeSignature(VTInteger)), true},
		{"integerFalse", NewExprValueInteger(5), false},
		{"listTrue", NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean))), true},
		{"listFalse", NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("value"),
		}), false},
		{"mapTrue", NewNilExprValue(NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTBoolean))), true},
		{"mapFalse", NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
			"key": NewExprValueString("value"),
		}), false},
		{"regexpTrue", NewNilExprValue(NewScalarTypeSignature(VTRegexp)), true},
		{"regexpFalse", NewExprValueRegexpMust("[0-9]{3}"), false},
		{"stringTrue", NewNilExprValue(NewScalarTypeSignature(VTString)), true},
		{"stringFalse", NewExprValueString("s"), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.rv.Nil() != test.isNil {
				t.Errorf("unexpected result (%t != %t)", test.rv.Nil(), test.isNil)
			}
		})
	}
}

func TestNewExprValue(t *testing.T) {
	tests := []struct {
		name  string
		ts    TypeSignature
		value interface{}
		ev    Value
	}{
		{"boolean/true", NewScalarTypeSignature(VTBoolean), true, NewExprValueBoolean(true)},
		{"boolean/false", NewScalarTypeSignature(VTBoolean), false, NewExprValueBoolean(false)},
		{"integer/int", NewScalarTypeSignature(VTInteger), 5, NewExprValueInteger(5)},
		{"integer/int8", NewScalarTypeSignature(VTInteger), int8(5), NewExprValueInteger(5)},
		{"integer/int16", NewScalarTypeSignature(VTInteger), int16(5), NewExprValueInteger(5)},
		{"integer/int32", NewScalarTypeSignature(VTInteger), int32(5), NewExprValueInteger(5)},
		{"integer/int64", NewScalarTypeSignature(VTInteger), int64(5), NewExprValueInteger(5)},
		{"integer/uint", NewScalarTypeSignature(VTInteger), uint(5), NewExprValueInteger(5)},
		{"integer/uint8", NewScalarTypeSignature(VTInteger), uint8(5), NewExprValueInteger(5)},
		{"integer/uint16", NewScalarTypeSignature(VTInteger), uint16(5), NewExprValueInteger(5)},
		{"integer/uint32", NewScalarTypeSignature(VTInteger), uint32(5), NewExprValueInteger(5)},
		{"integer/uint64", NewScalarTypeSignature(VTInteger), uint64(5), NewExprValueInteger(5)},
		{"list", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)),
			[]Value{NewExprValueMust(NewScalarTypeSignature(VTString), "v1")},
			NewExprValueMust(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)),
				[]Value{NewExprValueMust(NewScalarTypeSignature(VTString), "v1")})},
		{"map", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)),
			map[string]Value{"k1": NewExprValueMust(NewScalarTypeSignature(VTString), "v1")},
			NewExprValueMust(NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)),
				map[string]Value{"k1": NewExprValueMust(NewScalarTypeSignature(VTString), "v1")})},
		{"regexp", NewScalarTypeSignature(VTRegexp), "[0-9]{3}", NewExprValueRegexpMust("[0-9]{3}")},
		{"string", NewScalarTypeSignature(VTString), "a string", NewExprValueString("a string")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ev, err := NewExprValue(test.ts, test.value)
			if err != nil {
				t.Errorf("unexpected error %v", err)
				return
			}
			if !ev.Equal(test.ev) {
				t.Errorf("unexpected created value (%v != %v)", ev, test.value)
			}
		})
	}
}

func TestNewExprValueError(t *testing.T) {
	tests := []struct {
		name  string
		ts    TypeSignature
		value interface{}
	}{
		{"boolean", NewScalarTypeSignature(VTString), true},
		{"integer/int", NewScalarTypeSignature(VTString), 5},
		{"integer/int8", NewScalarTypeSignature(VTString), int8(5)},
		{"integer/int16", NewScalarTypeSignature(VTString), int16(5)},
		{"integer/int32", NewScalarTypeSignature(VTString), int32(5)},
		{"integer/int64", NewScalarTypeSignature(VTString), int64(5)},
		{"integer/uint", NewScalarTypeSignature(VTString), uint(5)},
		{"integer/uint8", NewScalarTypeSignature(VTString), uint8(5)},
		{"integer/uint16", NewScalarTypeSignature(VTString), uint16(5)},
		{"integer/uint32", NewScalarTypeSignature(VTString), uint32(5)},
		{"integer/uint64", NewScalarTypeSignature(VTString), uint64(5)},
		{"list", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)),
			[]Value{NewExprValueMust(NewScalarTypeSignature(VTString), "v1")}},
		{"map", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)),
			map[string]Value{"k1": NewExprValueMust(NewScalarTypeSignature(VTString), "v1")}},
		{"regexp", NewScalarTypeSignature(VTInteger), "[0-9]{3}"},
		{"string", NewScalarTypeSignature(VTInteger), "a string"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ev, err := NewExprValue(test.ts, test.value)
			if err == nil {
				t.Errorf("expected error %v", err)
			}
			if !ev.Equal(EvNil) {
				t.Errorf("expected EvNil: %v", ev)
			}
		})
	}
}

func TestNewExprValueMustPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	NewExprValueMust(NewScalarTypeSignature(VTString), 5)
}

func TestNewExprValueFromInterface(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		ev    Value
	}{
		{"boolean/true", true, NewExprValueBoolean(true)},
		{"boolean/false", false, NewExprValueBoolean(false)},
		{"integer/int", 5, NewExprValueInteger(5)},
		{"integer/int8", int8(5), NewExprValueInteger(5)},
		{"integer/int16", int16(5), NewExprValueInteger(5)},
		{"integer/int32", int32(5), NewExprValueInteger(5)},
		{"integer/int64", int64(5), NewExprValueInteger(5)},
		{"integer/uint", uint(5), NewExprValueInteger(5)},
		{"integer/uint8", uint8(5), NewExprValueInteger(5)},
		{"integer/uint16", uint16(5), NewExprValueInteger(5)},
		{"integer/uint32", uint32(5), NewExprValueInteger(5)},
		{"integer/uint64", uint64(5), NewExprValueInteger(5)},
		{"list", []interface{}{"v1"},
			NewExprValueList(NewScalarTypeSignature(VTString),
				[]Value{NewExprValueMust(NewScalarTypeSignature(VTString), "v1")})},
		{"map", map[string]interface{}{"k1": "v1"},
			NewExprValueMap(NewScalarTypeSignature(VTString),
				map[string]Value{"k1": NewExprValueMust(NewScalarTypeSignature(VTString), "v1")})},
		{"nil", nil, EvNil},
		{"string", "a string", NewExprValueString("a string")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ev, err := NewExprValueFromInterface(test.value)
			if err != nil {
				t.Errorf("unexpected error %v", err)
				return
			}
			if !ev.Equal(test.ev) {
				t.Errorf("unexpected created value (%v != %v)", ev, test.value)
			}
		})
	}
}

func TestNewExprValueFromString(t *testing.T) {
	tests := []struct {
		name string
		ts   TypeSignature
		str  string
		ev   Value
	}{
		{"boolean/true", NewScalarTypeSignature(VTBoolean), "true", NewExprValueBoolean(true)},
		{"boolean/false", NewScalarTypeSignature(VTBoolean), "false", NewExprValueBoolean(false)},
		{"integer", NewScalarTypeSignature(VTInteger), "5", NewExprValueInteger(5)},
		// List not supported
		// Map not supported
		{"nil", TsNil, "nil", EvNil},
		{"regexp", NewScalarTypeSignature(VTRegexp), "[0-9]{3}", NewExprValueRegexpMust("[0-9]{3}")},
		{"string", NewScalarTypeSignature(VTString), "a string", NewExprValueString("a string")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := NewExprValueFromString(test.ts, test.str)
			if err != nil {
				t.Errorf("unexpected error %v", err)
				return
			}
			if !value.Equal(test.ev) {
				t.Errorf("unexpected created value (%v != %v)", value, test.ev)
			}
		})
	}
}

func TestNewExprValueFromStringError(t *testing.T) {
	tests := []struct {
		name string
		ts   TypeSignature
		str  string
		ev   Value
	}{
		{"boolean", NewScalarTypeSignature(VTBoolean), "not a boolean", NewNilExprValue(NewScalarTypeSignature(VTBoolean))},
		{"integer", NewScalarTypeSignature(VTInteger), "not an integer", NewNilExprValue(NewScalarTypeSignature(VTInteger))},
		{"list", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), "list unsupported",
			NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))},
		{"map", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)), "map unsupported",
			NewNilExprValue(NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)))},
		// A nil expression value may always be created
		{"regexp", NewScalarTypeSignature(VTRegexp), "[invalid regexp}",
			NewNilExprValue(NewScalarTypeSignature(VTRegexp))},
		// An expression string may always be created from a go string
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := NewExprValueFromString(test.ts, test.str)
			if err == nil {
				t.Errorf("expected error")
				return
			}
			if !value.Equal(test.ev) {
				t.Errorf("expected nil expression value: %v", value)
			}
		})
	}
}

func TestNewNilExprValue(t *testing.T) {
	tests := []struct {
		name string
		ts   TypeSignature
	}{
		{"boolean", NewScalarTypeSignature(VTBoolean)},
		{"integer", NewScalarTypeSignature(VTInteger)},
		{"list", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))},
		{"map", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString))},
		{"regexp", NewScalarTypeSignature(VTRegexp)},
		{"string", NewScalarTypeSignature(VTString)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := NewNilExprValue(test.ts)
			if !value.Nil() {
				t.Errorf("expected nil value. Got %v", value)
			}
			if !value.Type.Equal(test.ts) {
				t.Errorf("wrong nil value type (%v != %v)", value.Type, test.ts)
			}
		})
	}
}

func TestConstantRelValues(t *testing.T) {
	tests := []struct {
		name string
		crv  Value
		exp  Value
	}{
		{"RvBooleanFalse", EvBooleanFalse, NewExprValueBoolean(false)},
		{"RvBooleanTrue", EvBooleanTrue, NewExprValueBoolean(true)},
		{"RvStringEmpty", EvStringEmpty, NewExprValueString("")},
		{"EvNil", EvNil, NewNilExprValue(TsNil)},
		{"EvNilBoolean", EvNilBoolean, NewNilExprValue(NewScalarTypeSignature(VTBoolean))},
		{"EvNilInteger", EvNilInteger, NewNilExprValue(NewScalarTypeSignature(VTInteger))},
		{"EvNilRegexp", EvNilRegexp, NewNilExprValue(NewScalarTypeSignature(VTRegexp))},
		{"EvNilString", EvNilString, NewNilExprValue(NewScalarTypeSignature(VTString))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.exp.Equal(test.crv) {
				t.Errorf("unexpected result (%v != %v)", test.crv, test.exp)
			}
		})
	}
}

func TestTypeSignature_IsValueType(t *testing.T) {
	tests := []struct {
		name string
		ts   TypeSignature
		vt   ValueType
	}{
		{"boolean", NewScalarTypeSignature(VTBoolean), VTBoolean},
		{"integer", NewScalarTypeSignature(VTInteger), VTInteger},
		{"list", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), VTList},
		{"map", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)), VTMap},
		{"string", NewScalarTypeSignature(VTString), VTString},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.ts.IsValueType(test.vt) {
				t.Errorf("expected IsValueType(%v) to be true for type signature %v", test.vt, test.ts)
			}
		})
	}
}

func TestTypeSignature_String(t *testing.T) {
	tests := []struct {
		name string
		ts   TypeSignature
		str  string
	}{
		{"scalar", NewScalarTypeSignature(VTBoolean),
			"{boolean}"},
		{"composite1", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)),
			"{list {string}}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ts.String() != test.str {
				t.Errorf("invalid String() result\nactual:   %s\nexpected: %s", test.ts.String(), test.str)
			}
		})
	}
}

// Also tests NewScalarTypeSignature(), NewCompositeTypeSignature()
func TestTypeSignature_Equal(t *testing.T) {
	tests := []struct {
		name string
		ts1  TypeSignature
		ts2  TypeSignature
		res  bool
	}{
		{"booleanTrue", NewScalarTypeSignature(VTBoolean), NewScalarTypeSignature(VTBoolean), true},
		{"booleanFalse1", NewScalarTypeSignature(VTBoolean), NewScalarTypeSignature(VTString), false},
		{"booleanFalse2", NewScalarTypeSignature(VTBoolean), NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean)), false},

		{"integerTrue", NewScalarTypeSignature(VTInteger), NewScalarTypeSignature(VTInteger), true},
		{"integerFalse1", NewScalarTypeSignature(VTInteger), NewScalarTypeSignature(VTBoolean), false},
		{"integerFalse2", NewScalarTypeSignature(VTInteger), NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTInteger)), false},

		{"listTrue", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean)),
			NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean)), true},
		{"listFalse1", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean)),
			NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTInteger)), false},
		{"listFalse2", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTBoolean)), NewScalarTypeSignature(VTBoolean), false},

		{"mapTrue", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTBoolean)),
			NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTBoolean)), true},
		{"mapFalse1", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTBoolean)),
			NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTInteger)), false},
		{"mapFalse2", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTBoolean)), NewScalarTypeSignature(VTBoolean), false},

		{"stringTrue", NewScalarTypeSignature(VTString), NewScalarTypeSignature(VTString), true},
		{"stringFalse1", NewScalarTypeSignature(VTString), NewScalarTypeSignature(VTBoolean), false},
		{"stringFalse2", NewScalarTypeSignature(VTString), NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ts1.Equal(test.ts2) != test.res {
				t.Errorf("wrong result for %v equal %v\nactual:   %v\nexpected: %v", test.ts1, test.ts2, test.ts1.Equal(test.ts2), test.res)
			}
		})
	}
}

func TestTypeSignature_Scalar(t *testing.T) {
	tests := []struct {
		name   string
		ts     TypeSignature
		scalar bool
	}{
		{"boolean", NewScalarTypeSignature(VTBoolean), true},
		{"integer", NewScalarTypeSignature(VTInteger), true},
		{"list", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), false},
		{"map", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)), false},
		{"string", NewScalarTypeSignature(VTRegexp), true},
		{"string", NewScalarTypeSignature(VTString), true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.ts.Scalar() == test.scalar {
				t.Errorf("wrong scalar result (%v != %v", test.ts.Scalar(), test.scalar)
			}
		})
	}
}

// Also tests NewEmptyTypeSignature()
func TestTypeSignature_Empty(t *testing.T) {
	tests := []struct {
		name string
		rv   TypeSignature
		res  bool
	}{
		{"true", NewEmptyTypeSignature(), true},
		{"false", NewScalarTypeSignature(VTString), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.rv.Empty()
			if res != test.res {
				t.Errorf("unexpected result (%t != %t)", res, test.res)
			}
		})
	}
}
