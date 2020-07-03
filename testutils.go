package goexpr

import (
	"fmt"
	"strings"
)

type testRequestContext struct {
	// We store all values as strings
	Values map[string]string
}

func (rc testRequestContext) Reference(key interface{}, ts TypeSignature) (Value, error) {
	// We only support string based keys in our test context
	keyS, ok := key.(string)
	if !ok {
		return NewNilExprValue(ts), fmt.Errorf("key %v is not a string", key)
	}
	vS, found := rc.Values[keyS]
	if !found {
		return NewNilExprValue(ts), nil
	}
	v, err := NewExprValueFromString(ts, vS)
	if err != nil {
		return NewNilExprValue(ts), fmt.Errorf("can't convert value %v to type %v: %v", v, ts, err)
	}
	return v, nil
}

func (rc testRequestContext) Assign(key interface{}, value Value) error {
	// We only support string based keys in our test context
	keyS, ok := key.(string)
	if !ok {
		return fmt.Errorf("key %v is not a string", key)
	}
	if value.Nil() {
		// We treat nil value as "no value"
		return nil
	}
	// Store value as a string removing eventual string markers (") that exist for expression string values
	rc.Values[keyS] = strings.Trim(value.String(), "\"")
	return nil
}

func newEmptyTestRequestContext() RequestContext {
	values := make(map[string]string, 0)
	return newTestRequestContext(values)
}

func newTestRequestContext(values map[string]string) RequestContext {
	return testRequestContext{Values: values}
}
