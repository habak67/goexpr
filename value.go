package goexpr

import (
	"encoding/json"
	"fmt"
	"github.com/habak67/go-utils"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Common constant type signatures
var TsNil TypeSignature
var TsDefault TypeSignature

// Common constant expression values
var EvBooleanTrue Value
var EvBooleanFalse Value
var EvStringEmpty Value
var EvNil Value
var EvNilBoolean Value
var EvNilInteger Value
var EvNilRegexp Value
var EvNilString Value

func init() {
	TsNil = NewScalarTypeSignature(VTNil)
	// If we don't have information about the type signature we assume a scalar string.
	TsDefault = NewScalarTypeSignature(VTString)
	EvBooleanTrue = NewExprValueBoolean(true)
	EvBooleanFalse = NewExprValueBoolean(false)
	EvStringEmpty = NewExprValueString("")
	EvNilBoolean = NewNilExprValue(NewScalarTypeSignature(VTBoolean))
	EvNilInteger = NewNilExprValue(NewScalarTypeSignature(VTInteger))
	EvNilRegexp = NewNilExprValue(NewScalarTypeSignature(VTRegexp))
	EvNilString = NewNilExprValue(NewScalarTypeSignature(VTString))
	EvNil = NewNilExprValue(TsNil)
}

// Typed value. The supported value type are represented as follows.
// Boolean
//   type = boolean
//	 value = bool
// Integer
//   type = integer
//   value = int
// List
//   type = list/<type of list values>
//   value = []Value
// Map
//   type = map/<type of map values>
//   value = map[string]Value
// Regexp
//   type = regexp
//   value = regexp definition (string)
//   regexp = compiled regexp
// String
//   type = string
//   value = string
// Struct
//   type = struct
//   value = []Value
type Value struct {
	// Type signature for the value
	Type TypeSignature `json:"type"`
	// The value itself. Note that the backend value (e.g. a string or a list) is depending of the type.
	// for more information read the code for UnmarshalJSON().
	Value interface{} `json:"value,omitempty"`
	// Compiled regular expression for a regexp
	Regexp *regexp.Regexp `json:"-"`
}

// Nil checks if the value is nil (no value).
func (ev Value) Nil() bool {
	return ev.Value == nil
}

// Equal return true if the REL value is equal to a specified REL value and false otherwise.
// If the value type doesn't support equality check a panic is raised.
func (ev Value) Equal(ev2 Value) bool {
	if !ev.Type.Equal(ev2.Type) {
		return false
	}
	// Nil is equal to nil but not equal to non-nil
	if ev.Nil() || ev2.Nil() {
		return ev.Nil() == ev2.Nil()
	}
	switch ev.Type.BaseType {
	case VTBoolean:
		return ev.Value.(bool) == ev2.Value.(bool)
	case VTInteger:
		return ev.Value.(int) == ev2.Value.(int)
	case VTList:
		l1 := ev.Value.([]Value)
		l2 := ev2.Value.([]Value)
		if len(l1) != len(l2) {
			return false
		}
		// Check each corresponding list value
		for i := 0; i < len(l1); i++ {
			if !l1[i].Equal(l2[i]) {
				return false
			}
		}
		return true
	case VTMap:
		m1 := ev.Value.(map[string]Value)
		m2 := ev2.Value.(map[string]Value)
		if len(m1) != len(m2) {
			return false
		}
		// Check each map entry
		for k, v1 := range m1 {
			v2, ok := m2[k]
			if !ok {
				return false
			}
			if !v1.Equal(v2) {
				return false
			}
		}
		return true
	case VTString, VTRegexp:
		return ev.Value.(string) == ev2.Value.(string)
	default:
		panic(fmt.Sprintf("value type %v doesn't support equality", ev.Type.BaseType))
	}
}

// Compare return a negative value if the REL value is less than a specified REL value.
// It return positive value if the REL value is greater than the specified REL value
// and it return 0 if the REL value is equal to the specified REL value.
// The result is returned as an integer rel value.
// If one of the rel values are nil then the result is nil.
// If the value type is not comparable a panic is raised.
func (ev Value) Compare(Ev2 Value) Value {
	if !ev.Type.Equal(Ev2.Type) {
		panic(fmt.Sprintf("incompatible values to compare (%v != %v)", ev.Type, Ev2.Type))
	}
	switch ev.Type.BaseType {
	case VTInteger:
		if ev.Nil() || Ev2.Nil() {
			return NewNilExprValue(ev.Type)
		}
		return NewExprValueInteger(ev.Value.(int) - Ev2.Value.(int))
	default:
		panic(fmt.Sprintf("value type %v is not comparable", ev.Type.BaseType))
	}
}

// String return a compact string representation of the REL value
func (ev Value) String() string {
	if ev.Nil() {
		return "<<nil>>"
	}
	switch ev.Type.BaseType {
	case VTBoolean:
		return strconv.FormatBool(ev.Value.(bool))
	case VTInteger:
		return strconv.FormatInt(int64(ev.Value.(int)), 10)
	case VTList:
		var sb strings.Builder
		sb.WriteString("[")
		first := true
		for _, val := range ev.Value.([]Value) {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(val.String())
			first = false
		}
		sb.WriteString("]")
		return sb.String()
	case VTMap:
		var sb strings.Builder
		sb.WriteString("{")

		// To get a consistent map string we "sort" the map by key
		valueMap := ev.Value.(map[string]Value)
		keys := make([]string, 0, len(valueMap))
		for key := range valueMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		first := true
		for _, key := range keys {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(key)
			sb.WriteString(":")
			sb.WriteString(valueMap[key].String())
			first = false
		}

		sb.WriteString("}")
		return sb.String()
	case VTString, VTRegexp:
		// A string value is represented as a string literar including string markers ("a string")
		var sb strings.Builder
		sb.WriteString(`"`)
		sb.WriteString(ev.Value.(string))
		sb.WriteString(`"`)
		return sb.String()
	default:
		panic(fmt.Sprintf("value type %v has no string representation", ev.Type.BaseType))
	}
}

// SearchAll return all values related to the search key.
// The key and the value to search for must have a natural string representation.
// If key exist searchAll return true otherwise searchAll return a nil slice and false.
// If the value type is not searchable a panic is raised.
func (ev Value) SearchAll(key Value) ([]Value, bool) {
	switch ev.Type.BaseType {
	case VTList:
		list := ev.Value.([]Value)
		res := make([]Value, 0)
		for _, v := range list {
			// We currently only return a single instance even if the searched for value occurs multiple times in the list.
			if v.String() == key.String() {
				res = append(res, v)
			}
		}
		return res, len(res) > 0
	case VTMap:
		// trim eventual string markers (")
		keyS := strings.Trim(key.String(), "\"")
		m := ev.Value.(map[string]Value)
		v, ok := m[keyS]
		return []Value{v}, ok
	default:
		panic(fmt.Sprintf("value type %v is not searchable", ev.Type.BaseType))
	}
}

// Assign assigns a specified sub-value to the specified key.
func (ev Value) Assign(_ interface{}, _ Value) {
	panic(fmt.Sprintf("value type %v is not assignable", ev.Type.BaseType))
}

// Reference returns the sub-value for the specified key
func (ev Value) Reference(key interface{}) Value {
	switch ev.Type.BaseType {
	case VTMap:
		// We only support string based map keys
		keyS := key.(string)
		m := ev.Value.(map[string]Value)
		return m[keyS]
	default:
		panic(fmt.Sprintf("value type %v is not referable", ev.Type.BaseType))
	}
}

func (ev *Value) UnmarshalJSON(data []byte) error {
	var ev1 struct {
		Type  TypeSignature   `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	err := json.Unmarshal(data, &ev1)
	if err != nil {
		return err
	}
	ev.Type = ev1.Type
	if len(ev1.Value) == 0 {
		return nil
	}
	// Do a typed value dependent unmarshalling depending on the value type
	switch ev.Type.BaseType {
	case VTBoolean:
		var v bool
		err := json.Unmarshal(ev1.Value, &v)
		if err != nil {
			return err
		}
		ev.Value = v
	case VTInteger:
		var v int
		err := json.Unmarshal(ev1.Value, &v)
		if err != nil {
			return err
		}
		ev.Value = v
	case VTList:
		var list []Value
		err := json.Unmarshal(ev1.Value, &list)
		if err != nil {
			return err
		}
		ev.Value = list
	case VTMap:
		var m map[string]Value
		err := json.Unmarshal(ev1.Value, &m)
		if err != nil {
			return err
		}
		ev.Value = m
	case VTRegexp:
		var v string
		err := json.Unmarshal(ev1.Value, &v)
		if err != nil {
			return err
		}
		ev.Value = v
		// Pre-compile regexp
		ev.Regexp, err = regexp.Compile(v)
		if err != nil {
			return fmt.Errorf("error pre-compiling regexp %s: %v", v, err)
		}
	case VTString:
		var v string
		err := json.Unmarshal(ev1.Value, &v)
		if err != nil {
			return err
		}
		ev.Value = v
	default:
		panic(fmt.Sprintf("unknown value type %v", ev1.Type))
	}
	return nil
}

// NewExprValue creates a new expression value from a type signature and a go value. The created expression value
// is depending of the type signature. Note that the go value must match the specified type signature otherwise an
// error is returned.
// VTBoolean => bool
// VTInteger => int (including intX and uintX)
// VTList => []Value - a slice of expression values where the type signature unit type specifies the type of the values.
// VTMap => map[string]Value - a map with string keys and where the type signature sub type specifies the type for the values.
// VTNil => nil
// VTRegexp => string - a string representing a valid go regular expression including a pre-compiled regexp.
// VTString => string
func NewExprValue(ts TypeSignature, value interface{}) (Value, error) {
	if value == nil {
		return NewNilExprValue(ts), nil
	}
	switch v := value.(type) {
	case bool:
		if !ts.IsValueType(VTBoolean) {
			return EvNil, fmt.Errorf("value %v is not a boolean", value)
		}
		return NewExprValueBoolean(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if !ts.IsValueType(VTInteger) {
			return EvNil, fmt.Errorf("value %v is not an integer", value)
		}
		return NewExprValueInteger(utils.NewIntMust(v)), nil
	case []Value:
		if !ts.IsValueType(VTList) {
			return EvNil, fmt.Errorf("value %v is not a list", value)
		}
		return NewExprValueList(*ts.UnitType, v), nil
	case map[string]Value:
		if !ts.IsValueType(VTMap) {
			return EvNil, fmt.Errorf("value %v is not a map", value)
		}
		return NewExprValueMap(*ts.UnitType, v), nil
	case string:
		switch ts.BaseType {
		case VTRegexp:
			re, err := NewExprValueRegexp(v)
			if err != nil {
				return EvNil, fmt.Errorf("error creating regexp from string %s: %v", v, err)
			}
			return re, nil
		case VTString:
			return NewExprValueString(v), nil
		default:
			return EvNil, fmt.Errorf("invalid type signature %v for string value %s", ts, v)
		}
	default:
		return EvNil, fmt.Errorf("unexpected type of value %v", value)
	}
}

// NewExprValueMust creates a new expression value from a type signature and a go value. The same rules as for NewExprValue()
// applies. If there is an error creating the expression value a panic is raised.
func NewExprValueMust(ts TypeSignature, value interface{}) Value {
	ev, err := NewExprValue(ts, value)
	if err != nil {
		panic(fmt.Sprintf("unexpected error creating expression value for type signature %v and go value %v: %v", ts, value, err))
	}
	return ev
}

// NewExprValueFromInterface creates a new expression value from a go value. The created expression value is
// depending on the type of go value. If the go value type is unsupported an error is returned.
// The supported go value types is specified in NewExprValue().
// Note that a list must be of type []interface{}.
// Note that a map must be of type map[string]interface{}.
func NewExprValueFromInterface(value interface{}) (Value, error) {
	if value == nil {
		return EvNil, nil
	}
	switch v := value.(type) {
	case bool:
		return NewExprValue(NewScalarTypeSignature(VTBoolean), v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return NewExprValue(NewScalarTypeSignature(VTInteger), v)
	case []interface{}:
		// Recursively create a slice of expression values from the go slice values
		slice := make([]Value, 0)
		var unitType TypeSignature
		for _, subValue := range v {
			exprValue, err := NewExprValueFromInterface(subValue)
			if err != nil {
				return EvNil, fmt.Errorf("error generating expression value from slice value %v: %v", subValue, err)
			}
			// Make sure the slice values are of the same type
			if !unitType.Empty() && !unitType.Equal(exprValue.Type) {
				return EvNil, fmt.Errorf("slice with different value types (%v(%v) != %v)", subValue, exprValue.Type, unitType)
			}
			unitType = exprValue.Type
			slice = append(slice, exprValue)
		}
		// If empty slice we assume a slice of default value type
		if unitType.Empty() {
			unitType = TsDefault
		}
		return NewExprValue(NewCompositeTypeSignature(VTList, unitType), slice)
	case map[string]interface{}: // We only support string keys
		// Recursively create map expression values from the go map values
		mp := make(map[string]Value)
		var unitType TypeSignature
		for key, subValue := range v {
			exprValue, err := NewExprValueFromInterface(subValue)
			if err != nil {
				return EvNil, fmt.Errorf("error generating expression value from map value %v: %v", subValue, err)
			}
			// Make sure the map values are of the same type
			if !unitType.Empty() && !unitType.Equal(exprValue.Type) {
				return EvNil, fmt.Errorf("map with different value types (%v(%v) != %v)", subValue, exprValue.Type, unitType)
			}
			unitType = exprValue.Type
			mp[key] = exprValue
		}
		// If empty map we assume a slice of default value type
		if unitType.Empty() {
			unitType = TsDefault
		}
		return NewExprValue(NewCompositeTypeSignature(VTMap, unitType), mp)
	case string:
		// As a string will be a VTString we can't create regexp
		return NewExprValue(NewScalarTypeSignature(VTString), v)
	default:
		return EvNil, fmt.Errorf("can't convert go value %v to an expression value", v)
	}
}

func NewExprValueFromString(ts TypeSignature, value string) (Value, error) {
	switch ts.BaseType {
	case VTBoolean:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return NewNilExprValue(ts), err
		}
		return NewExprValueBoolean(b), nil
	case VTInteger:
		i, err := strconv.Atoi(value)
		if err != nil {
			return NewNilExprValue(ts), err
		}
		return NewExprValueInteger(i), nil
	case VTNil:
		return EvNil, nil
	case VTRegexp:
		return NewExprValueRegexp(value)
	case VTString:
		return NewExprValueString(value), nil
	}
	return NewNilExprValue(ts), fmt.Errorf("can't create rel value of type %v from string", ts.BaseType)
}

func NewExprValueBoolean(value bool) Value {
	return Value{
		Type:  NewScalarTypeSignature(VTBoolean),
		Value: value,
	}
}

func NewExprValueInteger(value int) Value {
	return Value{
		Type:  NewScalarTypeSignature(VTInteger),
		Value: value,
	}
}

func NewExprValueList(ut TypeSignature, list []Value) Value {
	return Value{
		Type:  NewCompositeTypeSignature(VTList, ut),
		Value: list,
	}
}

func NewExprValueMap(ut TypeSignature, EvMap map[string]Value) Value {
	return Value{
		Type:  NewCompositeTypeSignature(VTMap, ut),
		Value: EvMap,
	}
}

func NewNilExprValue(ts TypeSignature) Value {
	return Value{
		Type: ts,
	}
}

func NewExprValueRegexp(value string) (Value, error) {
	re, err := regexp.Compile(value)
	if err != nil {
		return NewNilExprValue(NewScalarTypeSignature(VTRegexp)), fmt.Errorf("error pre-compiling regexp %s: %v", value, err)

	}
	return Value{
		Type:   NewScalarTypeSignature(VTRegexp),
		Value:  value,
		Regexp: re,
	}, nil
}

func NewExprValueRegexpMust(value string) Value {
	re, err := NewExprValueRegexp(value)
	if err != nil {
		panic(err)
	}
	return re
}

func NewExprValueString(value string) Value {
	v := Value{
		Type:  NewScalarTypeSignature(VTString),
		Value: value,
	}
	return v
}

// Value type
type ValueType string

const (
	VTBoolean ValueType = "boolean"
	VTInteger ValueType = "integer"
	VTList    ValueType = "list"
	VTMap     ValueType = "map"
	VTNil     ValueType = "nil"
	VTRegexp  ValueType = "regexp"
	VTString  ValueType = "string"
)

// valueTypeDefaultValue returns the default value (both go value and RelValue) for the specified value type.
// If the value type have no specific default value type nil and false is returned.
func (vt ValueType) DefaultValue() (interface{}, Value, bool) {
	switch vt {
	case VTBoolean:
		return false, NewExprValueMust(NewScalarTypeSignature(VTBoolean), false), true
	case VTInteger:
		return 0, NewExprValueMust(NewScalarTypeSignature(VTInteger), 0), true
	case VTString:
		return "", NewExprValueMust(NewScalarTypeSignature(VTString), ""), true
	}
	return nil, Value{}, false
}

// Metadata for value types
type ValueTypeMetadata map[ValueType]struct {
	// May you check for equality between two values of the type
	equality bool
	// May you compare "size" between two values of the type
	comparable bool
	// May you search, including check for existence, for a scalar value in an instance of the type
	searchable bool
	// May you assign a sub-value to the type
	assignable bool
	// May you reference a sub-value from the type
	reference bool
	// May you iterate over the type (e.g. using a foreach loop)
	iterable bool
	// May you use the type as a value in an iteration (e.g. as the value for a foreach loop variable)
	iterationValue bool
	// Is the type a scalar type
	scalar bool
}

// ValidValueType checks if the specified value type is a known value type
func (v ValueTypeMetadata) ValidValueType(vt ValueType) bool {
	_, ok := v[vt]
	return ok
}

// Equality returns true is you may check for equality for the value type
func (v ValueTypeMetadata) Equality(vt ValueType) bool {
	return v[vt].equality
}

// Comparable returns true if you may compare two values of the specified value type
func (v ValueTypeMetadata) Comparable(vt ValueType) bool {
	return v[vt].comparable
}

// Searchable returns true is you ay search for sub values in a value of the specified value type
func (v ValueTypeMetadata) Searchable(vt ValueType) bool {
	return v[vt].searchable
}

// Assignable returns true if you may assign a sub-value to the type
func (v ValueTypeMetadata) Assignable(vt ValueType) bool {
	return v[vt].assignable
}

// Reference returns true if you may reference a sub-value from the type
func (v ValueTypeMetadata) Reference(vt ValueType) bool {
	return v[vt].reference
}

// Iterable returns true if the type is itereable (you may iterate over it using e.g. a foreach loop)
func (v ValueTypeMetadata) Iterable(vt ValueType) bool {
	return v[vt].iterable
}

// IterationValue returns true if you may use the type as a value in an iteration (e.g. a foreach loop)
func (v ValueTypeMetadata) IterationValue(vt ValueType) bool {
	return v[vt].iterationValue
}

// Scalar returns true if the type is a scalar type
func (v ValueTypeMetadata) Scalar(vt ValueType) bool {
	return v[vt].iterationValue
}

var VTMetadata = ValueTypeMetadata{
	VTBoolean: {true, false, false, false, false, false,
		true, true},
	VTInteger: {true, true, false, false, false, false,
		true, true},
	VTList: {true, false, true, false, false, true,
		false, false},
	VTMap: {true, false, true, false, false, false,
		false, false},
	VTRegexp: {true, false, false, false, false, false,
		true, true},
	VTString: {true, false, false, false, false, false,
		true, true},
}

// TypeSignature holds type information for a typed value
type TypeSignature struct {
	// Base type for the signature
	BaseType ValueType `json:"base_type"`
	// If the base type is a composite type (list or map) the type for the values hold by the composite type (e.g. list of strings)
	// Note that a struct could have values of different types. UnitType is therefore nil for structs.
	UnitType *TypeSignature `json:"unit_type,omitempty"`
}

func (ts TypeSignature) String() string {
	if ts.UnitType == nil {
		return fmt.Sprintf("{%s}", ts.BaseType)
	}
	return fmt.Sprintf("{%s %s}", ts.BaseType, ts.UnitType.String())
}

func (ts TypeSignature) Empty() bool {
	return ts == TypeSignature{}
}

func (ts TypeSignature) IsValueType(vt ValueType) bool {
	return ts.BaseType == vt
}

func (ts TypeSignature) Equal(ts2 TypeSignature) bool {
	if ts.BaseType != ts2.BaseType {
		return false
	}
	if (ts.UnitType != nil && ts2.UnitType == nil) || (ts.UnitType == nil && ts2.UnitType != nil) {
		return false
	}
	if ts.UnitType != nil {
		return ts.UnitType.Equal(*ts2.UnitType)
	}
	return true
}

func (ts TypeSignature) Scalar() bool {
	return VTMetadata.Scalar(ts.BaseType)
}

// NewScalarTypeSignature creates a type signature for a scalar value type
func NewScalarTypeSignature(base ValueType) TypeSignature {
	return TypeSignature{BaseType: base}
}

// NewCompositeTypeSignature create a type signature for a composite value type
func NewCompositeTypeSignature(base ValueType, unit TypeSignature) TypeSignature {
	return TypeSignature{
		BaseType: base,
		UnitType: &unit,
	}
}

func NewEmptyTypeSignature() TypeSignature {
	return TypeSignature{}
}
