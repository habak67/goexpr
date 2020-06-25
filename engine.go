package goexpr

type RequestContext interface {
	// Reference returns the value connected to the specified key. If key is not found an empty value
	// of the specified type signature is returned.
	Reference(key string, ts TypeSignature) Value
	// Assign assigns the specified value to the specified key.
	Assign(key string, value Value)
}
