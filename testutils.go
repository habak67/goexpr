package goexpr

import "fmt"

type testRequestContext struct {
	Values map[string]Value
}

func (rc testRequestContext) Reference(key interface{}, ts TypeSignature) (Value, error) {
	// We only support string based keys in our test context
	keyS, ok := key.(string)
	if !ok {
		return NewNilExprValue(ts), fmt.Errorf("key %v is not a string", key)
	}
	v, found := rc.Values[keyS]
	if !found {
		return NewNilExprValue(ts), nil
	}
	return v, nil
}

func (rc testRequestContext) Assign(key interface{}, value Value) error {
	// We only support string based keys in our test context
	keyS, ok := key.(string)
	if !ok {
		return fmt.Errorf("key %v is not a string", key)
	}
	rc.Values[keyS] = value
	return nil
}

func newEmptyTestRequestContext() RequestContext {
	values := make(map[string]Value, 0)
	return newTestRequestContext(values)
}

func newTestRequestContext(values map[string]Value) RequestContext {
	return testRequestContext{Values: values}
}
