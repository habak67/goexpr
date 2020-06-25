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
				NewExprValue(NewScalarTypeSignature(VTString), "string 1"),
				NewExprValue(NewScalarTypeSignature(VTString), "string 2"),
			}), `
{"type":{"base_type":"list","unit_type":{"base_type":"string"}},"value":[
{"type":{"base_type":"string"},"value":"string 1"},
{"type":{"base_type":"string"},"value":"string 2"}]}`},
		{"map", NewExprValueMap(NewScalarTypeSignature(VTString),
			map[string][]Value{
				"one": {
					NewExprValueString("string 1"),
				},
				"two": {
					NewExprValueString("string 21"),
					NewExprValueString("string 22"),
				},
			}), `
{"type":{"base_type":"map","unit_type":{"base_type":"string"}},"value":{
"one":[{"type":{"base_type":"string"},"value":"string 1"}],
"two":[{"type":{"base_type":"string"},"value":"string 21"},{"type":{"base_type":"string"},"value":"string 22"}]}}`},
		{"regexp", NewExprValueRegexpSilent("[0-9]{3}"), `
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
				return
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
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}), true},
		{"mapFalseUnitType",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTInteger), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}), false},
		{"mapFalseMapLen",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")}}), false},
		{"mapFalseKey",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1":       {NewExprValueString("value1")},
				"anotherKey": {NewExprValueString("value1")}}), false},
		{"mapFalseEntryLen",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1"), NewExprValueString("value11")}}), false},
		{"mapFalseValueType",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueInteger(2)}}), false},
		{"mapFalseValueValue",
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("value1")}}),
			NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
				"key1": {NewExprValueString("value1")},
				"key2": {NewExprValueString("another value")}}), false},

		{"regexpTrue",
			NewExprValueRegexpSilent("[0-9]{3}"), NewExprValueRegexpSilent("[0-9]{3}"), true},
		{"regexpFalse",
			NewExprValueRegexpSilent("[0-9]{3}"), NewExprValueRegexpSilent("another regexp"), false},
		{"regexpFalseType",
			NewExprValueRegexpSilent("[0-9]{3}"), NewExprValueString("[0-9]{3}"), false},

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
				return
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
				return
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

		{"mapFoundSingle", NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
			"foo": {
				NewExprValueString("foo1"),
			},
			"bar": {
				NewExprValueString("bar11"),
				NewExprValueString("bar12"),
			},
		}), NewExprValueString("foo"), []Value{
			NewExprValueString("foo1"),
		}, true},
		{"mapFoundMulti", NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
			"foo": {
				NewExprValueString("foo1"),
			},
			"bar": {
				NewExprValueString("bar11"),
				NewExprValueString("bar12"),
			},
		}), NewExprValueString("bar"), []Value{
			NewExprValueString("bar11"),
			NewExprValueString("bar12"),
		}, true},
		{"mapNotFound", NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
			"foo": {
				NewExprValueString("foo1"),
			},
			"bar": {
				NewExprValueString("bar11"),
				NewExprValueString("bar12"),
			},
		}), NewExprValueString("not found"), []Value{}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, ok := test.rvTest.SearchAll(test.rvKey)

			if test.found {
				if !ok {
					t.Errorf("key %v unexpectedly not found in value %v", test.rvKey, test.rvTest)
				}
				if !reflect.DeepEqual(res, test.rvRes) {
					t.Errorf("result not equal.\nactual:   %v\nexpected: %v", res, test.rvRes)
					return
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
		{"mapFalse", NewExprValueMap(NewScalarTypeSignature(VTString), map[string][]Value{
			"key": {NewExprValueString("value")},
		}), false},
		{"regexpTrue", NewNilExprValue(NewScalarTypeSignature(VTRegexp)), true},
		{"regexpFalse", NewExprValueRegexpSilent("[0-9]{3}"), false},
		{"stringTrue", NewNilExprValue(NewScalarTypeSignature(VTString)), true},
		{"stringFalse", NewExprValueString("s"), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.rv.Nil() != test.isNil {
				t.Errorf("unexpected result (%t != %t)", test.rv.Nil(), test.isNil)
				return
			}
		})
	}
}

func TestNewExprValueFromString(t *testing.T) {
	tests := []struct {
		name  string
		ts    TypeSignature
		vs    string
		ok    bool
		value Value
	}{
		{"boolean/true", NewScalarTypeSignature(VTBoolean), "true", true, NewExprValueBoolean(true)},
		{"boolean/false", NewScalarTypeSignature(VTBoolean), "false", true, NewExprValueBoolean(false)},
		{"boolean/error", NewScalarTypeSignature(VTBoolean), "not boolean", false, NewNilExprValue(NewScalarTypeSignature(VTBoolean))},
		{"integer/ok", NewScalarTypeSignature(VTInteger), "5", true, NewExprValueInteger(5)},
		{"integer/error", NewScalarTypeSignature(VTInteger), "not integer", false, NewNilExprValue(NewScalarTypeSignature(VTInteger))},
		{"list/error", NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), "list", false,
			NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))},
		{"map/error", NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)), "map", false,
			NewNilExprValue(NewCompositeTypeSignature(VTMap, NewScalarTypeSignature(VTString)))},
		{"regexp/ok", NewScalarTypeSignature(VTRegexp), "[0-9]{3}", true, NewExprValueRegexpSilent("[0-9]{3}")},
		{"regexp/error", NewScalarTypeSignature(VTRegexp), "[invalid regexp}", false, NewNilExprValue(NewScalarTypeSignature(VTRegexp))},
		{"string/ok", NewScalarTypeSignature(VTString), "a string", true, NewExprValueString("a string")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := NewExprValueFromString(test.ts, test.vs)
			if !test.ok {
				if err == nil {
					t.Errorf("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error %v", err)
				return
			}
			if !value.Equal(test.value) {
				t.Errorf("unexpected created value (%v != %v)", value, test.value)
				return
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
		{"EvNil", EvNil, NewNilExprValue(NewScalarTypeSignature(VTBoolean))},
		{"EvNilBoolean", EvNilBoolean, NewNilExprValue(NewScalarTypeSignature(VTBoolean))},
		{"EvNilInteger", EvNilInteger, NewNilExprValue(NewScalarTypeSignature(VTInteger))},
		{"EvNilRegexp", EvNilRegexp, NewNilExprValue(NewScalarTypeSignature(VTRegexp))},
		{"EvNilString", EvNilString, NewNilExprValue(NewScalarTypeSignature(VTString))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.exp.Equal(test.crv) {
				t.Errorf("unexpected result (%v != %v)", test.crv, test.exp)
				return
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
				return
			}
		})
	}
}
