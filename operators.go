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

// Compare Expression type
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

// Logical Expression type
type LogicalType string

const (
	LTAnd LogicalType = "and"
	LTOr  LogicalType = "or"
	LTNot LogicalType = "not"
)

// Search Expression type
type SearchType string

const (
	STExist   SearchType = "exist"
	STFind    SearchType = "find"
	STFindAll SearchType = "findAll"
)

// Expression represent a specific expression that calculates a result value from a set of sub-expressions.
type Expression interface {
	// Evaluate evaluates the expression using the specified request context. The value from the evaluation is returned.
	// If there was an error in the evaluation the error is returned.
	Evaluate(reqContext RequestContext) (Value, error)
	// Line returns the source line number in the rule specification for the start of the Expression definition
	Line() int
	// Col returns the source column number in the rule specification for the start of the Expression definition
	Col() int
	// ResultType returns the result type of the Expression
	ResultType() TypeSignature
	// nilResult returns a nil value of the value type of the Expression result type
	nilResult() Value
}

func newBaseExpression(rt TypeSignature, line, col int) baseExpression {
	return baseExpression{
		line:    line,
		col:     col,
		resType: rt,
	}
}

type baseExpression struct {
	line int
	col  int
	// Result type of the Expression
	resType TypeSignature
}

func (bo baseExpression) Line() int {
	return bo.line
}

func (bo baseExpression) Col() int {
	return bo.col
}

func (bo baseExpression) ResultType() TypeSignature {
	return bo.resType
}

func (bo baseExpression) nilResult() Value {
	return NewNilExprValue(bo.resType)
}

// exprAssign assigns a value to a reference in the request context reference heap or a referable value (struct).
// The result of the Expression is the value assigned (including nil).
// Source specifies the reference type (heap or referable variable). Index is the reference index to read. If source is
// a variable then the variable is the result of the source Expression.
// The value assigned to the reference is the result from the value Expression.
// If the result of the source Expression is nil then nothing is assigned.
// If the result of the value Expression is nil then nil is assigned to the reference.
type exprAssign struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The name of the reference as specified in the rule definition.
	name string
	// The reference key
	key interface{}
	// The value to be written "assigned" to the reference.
	valueOp Expression
	// The source for the reference operation (heap or variable).
	source ReferenceSource
	// If the source is "variable" the source op is used to get the variable.
	sourceOp Expression
}

func (op *exprAssign) Evaluate(recCtx RequestContext) (Value, error) {
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

func NewExprAssign(name string, key interface{}, valueOp, sourceOp Expression, source ReferenceSource, line, col int) Expression {
	return &exprAssign{
		baseExpression: newBaseExpression(valueOp.ResultType(), line, col),
		name:           name,
		key:            key,
		valueOp:        valueOp,
		sourceOp:       sourceOp,
		source:         source,
	}
}

// exprCompare compares the result from two sub-expressions using a specified compare Expression.
// The semantic is dependent of the compare Expression and follows normal semantics for those.
// The result is a boolean value.
// If one of the operands is a nil value the following applies.
// For equal and non-equal nil == nil and nil != non-nil.
// In other case the result is nil. That is the compare Expression propagates nil.
type exprCompare struct {
	baseExpression
	//	ruleCtx *ruleContext
	// Compare Expression
	ct      CompareType
	opLeft  Expression
	opRight Expression
}

func (op *exprCompare) Evaluate(recCtx RequestContext) (Value, error) {
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
		panic(fmt.Sprintf("unknown compare Expression type %v", op.ct))
	}
}

func NewExprCompare(ct CompareType, leftOp Expression, rightOp Expression, line, col int) (Expression, error) {
	// If match compare and matcher is a constant string convert to constant regexp
	if ct == CTMatch {
		constant, ok := rightOp.(*exprConstant)
		if ok && constant.ResultType().IsValueType(VTString) {
			str := constant.c.Value.(string)
			regexp, err := NewExprValueRegexp(str)
			if err != nil {
				return nil, fmt.Errorf("can't create regexp from %s: %v", str, err)
			}
			rightOp = NewExprConstant(regexp, line, col)
		}
	}

	return &exprCompare{
		baseExpression: newBaseExpression(NewScalarTypeSignature(VTBoolean), line, col),
		ct:             ct,
		opLeft:         leftOp,
		opRight:        rightOp,
	}, nil
}

func NewExprCompareMust(ct CompareType, leftOp Expression, rightOp Expression, line, col int) Expression {
	expr, err := NewExprCompare(ct, leftOp, rightOp, line, col)
	if err != nil {
		panic(fmt.Sprintf("error creating compare expression: %v", err))
	}
	return expr
}

// exprConstant returns a specified constant value.
type exprConstant struct {
	baseExpression
	//	ruleCtx *ruleContext
	c Value
}

func (op *exprConstant) Evaluate(_ RequestContext) (Value, error) {
	return op.c, nil
}

func NewExprConstant(c Value, line, col int) Expression {
	return &exprConstant{
		baseExpression: newBaseExpression(c.Type, line, col),
		c:              c,
	}
}

// exprError returns an error in evaluation.
// The Expression is used in unit tests to test failure scenarios.
type exprError struct {
	baseExpression
	//	ruleCtx *ruleContext
}

func (op *exprError) Evaluate(_ RequestContext) (Value, error) {
	return NewNilExprValue(op.resType), fmt.Errorf("exprError")
}

// exprFor represent a for loop.
// The loop Expression is applied one time for every value from the result of the list Expression (the result must be a list).
// The reference specified by the heap index is set to the value in each iteration so that the loop Expression may reference
// the current list value.
// The result of the for Expression is the result from the last execution of the loop Expression. That is the loop Expression
// applied to the last value in the list. This also implies that the result type is the same as the value type of the
// list values.
// If a break expression is specified and the loop expression returns the same result as the break expression (if exist)
// then break the loop and return the current return value.
// If the list is empty or a nil list then the loop Expression is never executed and nil is returned.
type exprFor struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The list of values to iterate over.
	opList Expression
	// The Expression to execute for each value in the value list.
	opLoop Expression
	// If the result of the loop Expression return the same result as the break Expression then break the for loop.
	opBreak Expression
	// The reference heap key where to store the current value of the value list.
	key string
}

func (op *exprFor) Evaluate(recCtx RequestContext) (Value, error) {
	list, err := op.opList.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	// If no value to loop on (empty or nil list) we return a nil value
	if list.Nil() || len(list.Value.([]Value)) == 0 {
		return op.nilResult(), nil
	}
	// Compute break value if break Expression exist
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

func NewExprFor(opList Expression, opLoop Expression, opBreak Expression, key string, line, col int) Expression {
	return &exprFor{
		baseExpression: newBaseExpression(*opList.ResultType().UnitType, line, col),
		opList:         opList,
		opLoop:         opLoop,
		opBreak:        opBreak,
		key:            key,
	}
}

// exprIf represent an if expression.
// The check Expression (must be a boolean) is evaluated.
// If the result is true then the then-Expression is evaluated otherwise the else-Expression is evaluated.
// The result of the if Expression is the result of the then- or else-Expression (the one evaluated). This also
// implies that the result type of the then- and else-operators must be the same.
// If the result of the check Expression is nil the neither then- nor else-Expression is evaluated and the
// result of the if Expression is nil. That is the if Expression propagates nil.
type exprIf struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The expression to check if the then or else sub-expression should be executed.
	checkOp Expression
	thenOp  Expression
	elseOp  Expression
}

func (op *exprIf) Evaluate(recCtx RequestContext) (Value, error) {
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
	// If no else Expression return a nil value.
	if op.elseOp == nil {
		return op.nilResult(), nil
	}
	elseRes, err := op.elseOp.Evaluate(recCtx)
	if err != nil {
		return op.nilResult(), err
	}
	return elseRes, nil
}

func NewExprIf(checkOp Expression, thenOp Expression, elseOp Expression, line, col int) Expression {
	return &exprIf{
		baseExpression: newBaseExpression(thenOp.ResultType(), line, col),
		checkOp:        checkOp,
		thenOp:         thenOp,
		elseOp:         elseOp,
	}
}

// exprLogical represent a logical expression.
// Note that lazy evaluation is used. That is if opLeft returns false for "and" then opRight will not be
// evaluated and if opLeft evaluates to true for "or" then opRight will not be evaluated.
// If left or right operand evaluates to nil the the result of the logical Expression is nil. That is
// nil is propagated.
type exprLogical struct {
	baseExpression
	//	ruleCtx *ruleContext
	lt      LogicalType
	opLeft  Expression
	opRight Expression
}

func (op *exprLogical) Evaluate(recCtx RequestContext) (Value, error) {
	resLeft, err := op.opLeft.Evaluate(recCtx)
	if err != nil {
		return EvNilBoolean, err
	}
	if resLeft.Nil() {
		return op.nilResult(), nil
	}
	// we use "lazy evaluation" in the sense that we return as soon as we know the result of the logical Expression
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
		panic(fmt.Sprintf("unknown logial Expression type %v", op.lt))
	}
}

func NewExprLogical(lt LogicalType, leftOp Expression, rightOp Expression, line, col int) Expression {
	return &exprLogical{
		baseExpression: newBaseExpression(NewScalarTypeSignature(VTBoolean), line, col),
		lt:             lt,
		opLeft:         leftOp,
		opRight:        rightOp,
	}
}

func NewExprLogicalUnary(lt LogicalType, leftOp Expression, line, col int) Expression {
	return NewExprLogical(lt, leftOp, nil, line, col)
}

// exprReference reads a reference from the request context reference heap or a referable value (struct).
// The result of the Expression is the value read.
// Source specifies the reference type (heap or referable variable). Index is the reference index to read. If source is
// a variable then the variable is the result of the source Expression.
// If the result of the source Expression is nil then the result of the reference Expression is nil. That is the reference of
// a nil value is nil.
type exprReference struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The name of the reference as specified in the rule definition.
	name string
	// The reference key
	key interface{}
	// The source for the reference operation (heap or variable).
	source ReferenceSource
	// If the source is "variable" the source op is used to get the variable.
	sourceOp Expression
}

func (op *exprReference) Evaluate(recCtx RequestContext) (Value, error) {
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

func NewExprReference(name string, key interface{}, sourceOp Expression, source ReferenceSource, resType TypeSignature, line, col int) Expression {
	return &exprReference{
		baseExpression: newBaseExpression(resType, line, col),
		name:           name,
		key:            key,
		sourceOp:       sourceOp,
		source:         source,
	}
}

// exprSearch applies a specified search operation on a specified searchable value. The result of the Expression is specific
// to the search operation.
// Exist
//   Boolean indicating if the value to search for exist or not.
// Find
//   The found value.
//   If value is not found and a default Expression is specified the result of the default Expression is returned.
//   If value is not found and there is no default Expression (or result is nil) then nil is returned.
// FindAll
//   A list of all values found.
//   If no values are found and a default Expression is specified the result of the default Expression is returned (must be a list).
//   If no values are found and there is no default Expression (or result is nil) then nil is returned.
// If key or collection operators evaluates to nil then nil is returned.
type exprSearch struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The key to search for
	opKey Expression
	// The collection to search in
	opColl Expression
	// Default value to return if key not found in collection
	opDef Expression
	// Type of search operation
	searchType SearchType
}

func (op *exprSearch) Evaluate(recCtx RequestContext) (Value, error) {
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
			return NewExprValueList(*op.ResultType().UnitType, list), nil
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

func NewExprSearch(opKey, opColl, opDef Expression, searchType SearchType, resType TypeSignature, line, col int) Expression {
	return &exprSearch{
		baseExpression: newBaseExpression(resType, line, col),
		opKey:          opKey,
		opColl:         opColl,
		opDef:          opDef,
		searchType:     searchType,
	}
}

// exprSequence represent a sequence of sub-expressions.
// The value from the last sub-Expression is returned (including nil).
// Note that empty sequences (no sub-operators) are not supported.
type exprSequence struct {
	baseExpression
	//	ruleCtx *ruleContext
	// The sequence of sub-expressions to evaluate
	ops []Expression
}

func (op *exprSequence) Evaluate(recCtx RequestContext) (Value, error) {
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

func NewExprSequence(ops []Expression, line, col int) Expression {
	return &exprSequence{
		baseExpression: newBaseExpression(ops[len(ops)-1].ResultType(), line, col), // Empty sequences are not supported
		ops:            ops,
	}
}
