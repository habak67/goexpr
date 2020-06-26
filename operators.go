package goexpr

import (
	"fmt"
)

// Reference source type
type ReferenceSource string

const (
	RSHeap  ReferenceSource = "heap"
	RSValue ReferenceSource = "value"
)

// Compare operator type
type CompareType string

const (
	CTEqual        CompareType = "EQ"
	CTNotEqual     CompareType = "NE"
	CTLess         CompareType = "LS"
	CTLessEqual    CompareType = "LE"
	CTGreater      CompareType = "GT"
	CTGreaterEqual CompareType = "GE"
	CTMatch        CompareType = "match"
)

// Logical operator type
type LogicalType string

const (
	LTAnd LogicalType = "and"
	LTOr  LogicalType = "or"
	LTNot LogicalType = "not"
)

// Search operator type
type SearchType string

const (
	STExist   SearchType = "exist"
	STFind    SearchType = "find"
	STFindAll SearchType = "findAll"
)

type baseOperator struct {
	line int
	col  int
	// Result type of the operator
	resType TypeSignature
}

func (bo baseOperator) Line() int {
	return bo.line
}

func (bo baseOperator) Col() int {
	return bo.col
}

func (bo baseOperator) ResType() TypeSignature {
	return bo.resType
}

func (bo baseOperator) nilResult() Value {
	return NewNilExprValue(bo.resType)
}

// operator represent a specific expression that calculates a result value from a set of sub-expressions.
type operator interface {
	// An operator is an expression
	Expression
	// Line returns the source line number in the rule specification for the start of the operator definition
	Line() int
	// Col returns the source column number in the rule specification for the start of the operator definition
	Col() int
	// ResType returns the result type of the operator
	ResType() TypeSignature
	// nilResult returns a nil value of the value type of the operator result type
	nilResult() Value
}

// opAssign assigns a value to a reference in the request context reference heap or a referable value (struct).
// The result of the operator is the value assigned (including nil).
// Source specifies the reference type (heap or referable variable). Index is the reference index to read. If source is
// a variable then the variable is the result of the source operator.
// The value assigned to the reference is the result from the value operator.
// If the result of the source operator is nil then nothing is assigned.
// If the result of the value operator is nil then nil is assigned to the reference.
type opAssign struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The name of the reference as specified in the rule definition.
	name string
	// The reference key
	key interface{}
	// The value to be written "assigned" to the reference.
	valueOp operator
	// The source for the reference operation (heap or variable).
	source ReferenceSource
	// If the source is "variable" the source op is used to get the variable.
	sourceOp operator
}

func (op *opAssign) Evaluate(recCtx RequestContext) (Value, error) {
	value, err := op.valueOp.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	switch op.source {
	case RSHeap:
		err := recCtx.Assign(op.key, value)
		if err != nil {
			return op.nilResult(), err
		}
	case RSValue:
		// The source of the reference is a referable value. Currently struct is supported
		source, err := op.sourceOp.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		// We can't assign to nil.
		if !source.Nil() {
			source.Assign(op.key, value)
		}
	default:
		panic(fmt.Sprintf("unknown reference source %v", op.source))
	}
	return value, nil
}

// opCompare compares the result from two sub-expressions using a specified compare operator.
// The semantic is dependent of the compare operator and follows normal semantics for those.
// The result is a boolean value.
// If one of the operands is a nil value the following applies.
// For equal and non-equal nil == nil and nil != non-nil.
// In other case the result is nil. That is the compare operator propagates nil.
type opCompare struct {
	baseOperator
	//	ruleCtx *ruleContext
	// Compare operator
	ct      CompareType
	opLeft  operator
	opRight operator
}

func (op *opCompare) Evaluate(recCtx RequestContext) (Value, error) {
	resLeft, err := op.opLeft.Evaluate(recCtx)
	if err != nil {
		return EvNilBoolean, err
	}
	resRight, err := op.opRight.Evaluate(recCtx)
	if err != nil {
		return EvNilBoolean, err
	}

	switch op.ct {
	case CTEqual:
		return NewExprValueBoolean(resLeft.Equal(resRight)), nil
	case CTNotEqual:
		return NewExprValueBoolean(!resLeft.Equal(resRight)), nil
	}

	// Compare and match propagates nil. That is if one of the operands are nil the the result is nil
	if resLeft.Nil() || resRight.Nil() {
		return EvNilBoolean, nil
	}

	switch op.ct {
	case CTLess:
		return NewExprValueBoolean(resLeft.Compare(resRight).Value.(int) < 0), nil
	case CTLessEqual:
		return NewExprValueBoolean(resLeft.Compare(resRight).Value.(int) <= 0), nil
	case CTGreater:
		return NewExprValueBoolean(resLeft.Compare(resRight).Value.(int) > 0), nil
	case CTGreaterEqual:
		return NewExprValueBoolean(resLeft.Compare(resRight).Value.(int) >= 0), nil
	case CTMatch:
		// <string> match <regexp>
		check, err := op.opLeft.Evaluate(recCtx)
		if err != nil {
			return EvNilBoolean, err
		}
		matcher, err := op.opRight.Evaluate(recCtx)
		if err != nil {
			return EvNilBoolean, err
		}
		return NewExprValueBoolean(matcher.Regexp.MatchString(check.Value.(string))), nil
	default:
		// Should not happen...
		panic(fmt.Sprintf("unknown compare operator type %v", op.ct))
	}
}

// opConstant returns a specified constant value.
type opConstant struct {
	baseOperator
	//	ruleCtx *ruleContext
	c Value
}

func (op *opConstant) Evaluate(_ RequestContext) (Value, error) {
	return op.c, nil
}

// opError returns an error in evaluation.
// The operator is used in unit tests to test failure scenarios.
type opError struct {
	baseOperator
	//	ruleCtx *ruleContext
}

func (op *opError) Evaluate(_ RequestContext) (Value, error) {
	return NewNilExprValue(op.resType), fmt.Errorf("opError")
}

// opFor represent a for loop.
// The loop operator is applied one time for every value from the result of the list operator (the result must be a list).
// The reference specified by the heap index is set to the value in each iteration so that the loop operator may reference
// the current list value.
// The result of the for operator is the result from the last execution of the loop operator. That is the loop operator
// applied to the last value in the list. This also implies that the result type is the same as the value type of the
// list values.
// If a break expression is specified and the loop expression returns the same result as the break expression (if exist)
// then break the loop and return the current return value.
// If the list is empty or a nil list then the loop operator is never executed and nil is returned.
type opFor struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The list of values to iterate over.
	opList operator
	// The operator to execute for each value in the value list.
	opLoop operator
	// If the result of the loop operator return the same result as the break operator then break the for loop.
	opBreak operator
	// The reference heap key where to store the current value of the value list.
	key string
}

func (op *opFor) Evaluate(recCtx RequestContext) (Value, error) {
	list, err := op.opList.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	// If no value to loop on (empty or nil list) we return a nil value
	if list.Nil() || len(list.Value.([]Value)) == 0 {
		return op.nilResult(), nil
	}
	// Compute break value if break operator exist
	breakValue := NewNilExprValue(*list.Type.UnitType)
	if op.opBreak != nil {
		breakValue, err = op.opBreak.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
	}
	var res Value
	// The compiler should have checked that the list is a list.
	for _, value := range list.Value.([]Value) {
		err := recCtx.Assign(op.key, value)
		if err != nil {
			return op.nilResult(), err
		}
		res, err = op.opLoop.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		// Check for break if break value exist
		if !breakValue.Nil() && res.Equal(breakValue) {
			return res, nil
		}
	}
	// The result from the last loop iteration is the result of the for expression
	return res, nil
}

// opIf represent an if expression.
// The check operator (must be a boolean) is evaluated.
// If the result is true then the then-operator is evaluated otherwise the else-operator is evaluated.
// The result of the if operator is the result of the then- or else-operator (the one evaluated). This also
// implies that the result type of the then- and else-operators must be the same.
// If the result of the check operator is nil the neither then- nor else-operator is evaluated and the
// result of the if operator is nil. That is the if operator propagates nil.
type opIf struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The expression to check if the then or else sub-expression should be executed.
	checkOp operator
	thenOp  operator
	elseOp  operator
}

func (op *opIf) Evaluate(recCtx RequestContext) (Value, error) {
	res, err := op.checkOp.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	// If propagates nil.
	if res.Nil() {
		return op.nilResult(), nil
	}
	// The compiler should have checked that the check expression returns a boolean.
	if res.Value.(bool) {
		thenRes, err := op.thenOp.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		return thenRes, nil
	}
	// If no else operator return a nil value.
	if op.elseOp == nil {
		return op.nilResult(), nil
	}
	elseRes, err := op.elseOp.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	return elseRes, nil
}

// opLogical represent a logical expression.
// Note that lazy evaluation is used. That is if opLeft returns false for "and" then opRight will not be
// evaluated and if opLeft evaluates to true for "or" then opRight will not be evaluated.
// If left or right operand evaluates to nil the the result of the logical operator is nil. That is
// nil is propagated.
type opLogical struct {
	baseOperator
	//	ruleCtx *ruleContext
	lt      LogicalType
	opLeft  operator
	opRight operator
}

func (op *opLogical) Evaluate(recCtx RequestContext) (Value, error) {
	resLeft, err := op.opLeft.Evaluate(recCtx)
	if err != nil {
		return EvNilBoolean, err
	}
	if resLeft.Nil() {
		return op.nilResult(), nil
	}
	// we use "lazy evaluation" in the sense that we return as soon as we know the result of the logical operator
	switch op.lt {
	case LTAnd:
		if !resLeft.Value.(bool) {
			return EvBooleanFalse, nil
		}
		resRight, err := op.opRight.Evaluate(recCtx)
		if err != nil {
			return EvNilBoolean, err
		}
		if resRight.Nil() {
			return op.nilResult(), nil
		}
		if !resRight.Value.(bool) {
			return EvBooleanFalse, nil
		}
		return EvBooleanTrue, nil
	case LTOr:
		if resLeft.Value.(bool) {
			return EvBooleanTrue, nil
		}
		resRight, err := op.opRight.Evaluate(recCtx)
		if err != nil {
			return EvNilBoolean, err
		}
		if resRight.Nil() {
			return op.nilResult(), nil
		}
		if resRight.Value.(bool) {
			return EvBooleanTrue, nil
		}
		return EvBooleanFalse, nil
	case LTNot:
		if resLeft.Value.(bool) {
			return EvBooleanFalse, nil
		}
		return EvBooleanTrue, nil
	default:
		panic(fmt.Sprintf("unknown logial operator type %v", op.lt))
	}
}

// opReference reads a reference from the request context reference heap or a referable value (struct).
// The result of the operator is the value read.
// Source specifies the reference type (heap or referable variable). Index is the reference index to read. If source is
// a variable then the variable is the result of the source operator.
// If the result of the source operator is nil then the result of the reference operator is nil. That is the reference of
// a nil value is nil.
type opReference struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The name of the reference as specified in the rule definition.
	name string
	// The reference key
	key interface{}
	// The source for the reference operation (heap or variable).
	source ReferenceSource
	// If the source is "variable" the source op is used to get the variable.
	sourceOp operator
}

func (op *opReference) Evaluate(recCtx RequestContext) (Value, error) {
	switch op.source {
	case RSHeap:
		// The source of the reference is the request context heap
		return recCtx.Reference(op.key, op.resType)
	case RSValue:
		// The source of the reference is a referable value. Currently struct is supported.
		source, err := op.sourceOp.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		if source.Nil() {
			return op.nilResult(), nil
		}
		return source.Reference(op.key), nil
	default:
		panic(fmt.Sprintf("unknown reference source %v", op.source))
	}
}

// opSearch applies a specified search operation on a specified searchable value. The result of the operator is specific
// to the search operation.
// Exist
//   Boolean indicating if the value to search for exist or not.
// Find
//   The found value.
//   If value is not found and a default operator is specified the result of the default operator is returned.
//   If value is not found and there is no default operator (or result is nil) then nil is returned.
// FindAll
//   A list of all values found.
//   If no values are found and a default operator is specified the result of the default operator is returned (must be a list).
//   If no values are found and there is no default operator (or result is nil) then nil is returned.
// If key or collection operators evaluates to nil then nil is returned.
type opSearch struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The key to search for
	opKey operator
	// The collection to search in
	opColl operator
	// Default value to return if key not found in collection
	opDef operator
	// Type of search operation
	searchType SearchType
}

func (op *opSearch) Evaluate(recCtx RequestContext) (Value, error) {
	key, err := op.opKey.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	if key.Nil() {
		return op.nilResult(), nil
	}
	coll, err := op.opColl.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	if coll.Nil() {
		return op.nilResult(), nil
	}

	list, ok := coll.SearchAll(key)
	switch op.searchType {
	case STExist:
		if ok {
			return EvBooleanTrue, nil
		}
		return EvBooleanFalse, nil
	case STFind:
		if ok {
			// Return first value. There should be at least one value as ok == true
			return list[0], nil
		}
		// If key not found return default value if specified. Otherwise return nil.
		if op.opDef == nil {
			return op.nilResult(), nil
		}
		def, err := op.opDef.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		return def, nil
	case STFindAll:
		if ok {
			return NewExprValueList(*op.ResType().UnitType, list), nil
		}
		// If key not found return default value list if specified. Otherwise return a nil list.
		if op.opDef == nil {
			return op.nilResult(), nil
		}
		def, err := op.opDef.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
		return def, nil
	default:
		panic(fmt.Sprintf("unknown search type %v", op.searchType))
	}
}

// opSequence represent a sequence of sub-expressions.
// The value from the last sub-operator is returned (including nil).
// Note that empty sequences (no sub-operators) are not supported.
type opSequence struct {
	baseOperator
	//	ruleCtx *ruleContext
	// The sequence of sub-expressions to evaluate
	ops []operator
}

func (op *opSequence) Evaluate(recCtx RequestContext) (Value, error) {
	var res Value
	for _, subOp := range op.ops {
		var err error
		res, err = subOp.Evaluate(recCtx)
		if err != nil {
			return op.nilResult(), err
		}
	}
	return res, nil
}
