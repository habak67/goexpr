package goexpr

type RequestContext interface {
	// Reference returns the value connected to the specified key. If key is not found an empty value
	// of the specified type signature is returned. The concrete key datatype is dependent on the
	// implementation of the RequestContext interface.
	Reference(key interface{}, ts TypeSignature) (Value, error)
	// Assign assigns the specified value to the specified key. The concrete key datatype is dependent on the
	// implementation of the RequestContext interface.
	Assign(key interface{}, value Value) error
}

type Expression interface {
	// Evaluate evaluates the expression using the specified request context. The value from the evaluation is returned.
	// If there was an error in the evaluation the error is returned.
	Evaluate(reqContext RequestContext) (Value, error)
}
