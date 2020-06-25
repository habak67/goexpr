package goexpr

func testBaseOp(rt TypeSignature) baseOperator {
	return baseOperator{
		line:    5,
		col:     10,
		resType: rt,
	}
}

func newTestOpAssign(name string, key interface{}, valueOp, sourceOp operator, source ReferenceSource) *opAssign {
	return &opAssign{
		baseOperator: testBaseOp(valueOp.ResType()),
		name:         name,
		key:          key,
		valueOp:      valueOp,
		sourceOp:     sourceOp,
		source:       source,
	}
}

func newTestOpCompare(ct CompareType, leftOp operator, rightOp operator) *opCompare {
	return &opCompare{
		baseOperator: testBaseOp(NewScalarTypeSignature(VTBoolean)),
		ct:           ct,
		opLeft:       leftOp,
		opRight:      rightOp,
	}
}

func newTestOpConstant(c Value) *opConstant {
	return &opConstant{
		baseOperator: testBaseOp(c.Type),
		c:            c,
	}
}

func newTestOpFor(opList operator, opLoop operator, opBreak operator, key string) *opFor {
	return &opFor{
		baseOperator: testBaseOp(*opList.ResType().UnitType),
		opList:       opList,
		opLoop:       opLoop,
		opBreak:      opBreak,
		key:          key,
	}
}

func newTestOpIf(checkOp operator, thenOp operator, elseOp operator) *opIf {
	return &opIf{
		baseOperator: testBaseOp(thenOp.ResType()),
		checkOp:      checkOp,
		thenOp:       thenOp,
		elseOp:       elseOp,
	}
}

func newTestOpLogical(lt LogicalType, leftOp operator, rightOp operator) *opLogical {
	return &opLogical{
		baseOperator: testBaseOp(NewScalarTypeSignature(VTBoolean)),
		lt:           lt,
		opLeft:       leftOp,
		opRight:      rightOp,
	}
}

func newTestOpReference(name string, key interface{}, sourceOp operator, source ReferenceSource, resType TypeSignature) *opReference {
	return &opReference{
		baseOperator: testBaseOp(resType),
		name:         name,
		key:          key,
		sourceOp:     sourceOp,
		source:       source,
	}
}

func newTestOpSearch(opKey, opColl, opDef operator, searchType SearchType, resType TypeSignature) *opSearch {
	return &opSearch{
		baseOperator: testBaseOp(resType),
		opKey:        opKey,
		opColl:       opColl,
		opDef:        opDef,
		searchType:   searchType,
	}
}

func newTestOpSequence(ops []operator) *opSequence {
	return &opSequence{
		baseOperator: testBaseOp(ops[len(ops)-1].ResType()), // Empty sequences are not supported
		ops:          ops,
	}
}

type testRequestContext struct {
	Values map[string]Value
}

func (rc testRequestContext) Reference(key interface{}, ts TypeSignature) Value {
	// We only support string based keys in our test context
	keyS := key.(string)
	v, found := rc.Values[keyS]
	if !found {
		return NewNilExprValue(ts)
	}
	return v
}

func (rc testRequestContext) Assign(key interface{}, value Value) {
	// We only support string based keys in our test context
	keyS := key.(string)
	rc.Values[keyS] = value
}

func newEmptyTestRequestContext() RequestContext {
	values := make(map[string]Value, 0)
	return newTestRequestContext(values)
}

func newTestRequestContext(values map[string]Value) RequestContext {
	return testRequestContext{Values: values}
}
