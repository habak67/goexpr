package goexpr

import (
	"testing"
)

func TestOperator_Evaluate_Ok(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()
	// opReference and opReturn manipulate contexts and therefore has separate unit tests
	tests := []struct {
		name   string
		op     operator
		result Value
	}{
		// compare ---------------------------------------
		{"OpCompareEqualTrue", newTestOpCompare(CTEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(true)},
		{"OpCompareEqualTrueNil", newTestOpCompare(CTEqual, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(true)},
		{"OpCompareEqualFalse", newTestOpCompare(CTEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(false)},
		{"OpCompareEqualFalseNonNilNil", newTestOpCompare(CTEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(false)},
		{"OpCompareEqualFalseNilNonNil", newTestOpCompare(CTEqual, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(false)},

		{"OpCompareNotEqualTrue", newTestOpCompare(CTNotEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(false)},
		{"OpCompareNotEqualTrueNonNilNil", newTestOpCompare(CTNotEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(true)},
		{"OpCompareNotEqualTrueNilNonNil", newTestOpCompare(CTNotEqual, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(true)},
		{"OpCompareNotEqualFalse", newTestOpCompare(CTNotEqual, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(true)},
		{"OpCompareNotEqualFalseNil", newTestOpCompare(CTNotEqual, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(false)},

		{"OpCompareNonNilNil", newTestOpCompare(CTLess, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(EvNilInteger)), EvNilBoolean},
		{"OpCompareNilNonNil", newTestOpCompare(CTLess, newTestOpConstant(EvNilInteger),
			newTestOpConstant(NewExprValueInteger(1))), EvNilBoolean},
		{"OpCompareNilNilLess", newTestOpCompare(CTLess, newTestOpConstant(EvNilInteger),
			newTestOpConstant(EvNilInteger)), EvNilBoolean},
		{"OpCompareNilNilLessEqual", newTestOpCompare(CTLessEqual, newTestOpConstant(EvNilInteger),
			newTestOpConstant(EvNilInteger)), EvNilBoolean},
		{"OpCompareNilNilGreater", newTestOpCompare(CTGreater, newTestOpConstant(EvNilInteger),
			newTestOpConstant(EvNilInteger)), EvNilBoolean},
		{"OpCompareNilNilGreaterEqual", newTestOpCompare(CTGreaterEqual, newTestOpConstant(EvNilInteger),
			newTestOpConstant(EvNilInteger)), EvNilBoolean},
		{"OpCompareNilNilMatch", newTestOpCompare(CTMatch, newTestOpConstant(EvNilString),
			newTestOpConstant(EvNilRegexp)), EvNilBoolean},

		{"OpCompareLessTrue", newTestOpCompare(CTLess, newTestOpConstant(NewExprValueInteger(1)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(true)},
		{"OpCompareLessFalse", newTestOpCompare(CTLess, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(1))), NewExprValueBoolean(false)},

		{"OpCompareLessEqualTrue", newTestOpCompare(CTLessEqual, newTestOpConstant(NewExprValueInteger(1)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(true)},
		{"OpCompareLessEqualTrue", newTestOpCompare(CTLessEqual, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(true)},
		{"OpCompareLessEqualFalse", newTestOpCompare(CTLessEqual, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(1))), NewExprValueBoolean(false)},

		{"OpCompareGreaterTrue", newTestOpCompare(CTGreater, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(1))), NewExprValueBoolean(true)},
		{"OpCompareGreaterFalse", newTestOpCompare(CTGreater, newTestOpConstant(NewExprValueInteger(1)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(false)},

		{"OpCompareGreaterEqualTrue", newTestOpCompare(CTGreaterEqual, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(1))), NewExprValueBoolean(true)},
		{"OpCompareGreaterEqualTrue", newTestOpCompare(CTGreaterEqual, newTestOpConstant(NewExprValueInteger(2)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(true)},
		{"OpCompareGreaterEqualFalse", newTestOpCompare(CTGreaterEqual, newTestOpConstant(NewExprValueInteger(1)),
			newTestOpConstant(NewExprValueInteger(2))), NewExprValueBoolean(false)},

		{"OpCompareMatchTrue", newTestOpCompare(CTMatch, newTestOpConstant(NewExprValueString("123")),
			newTestOpConstant(NewExprValueRegexpMust("[0-9]{3}"))), NewExprValueBoolean(true)},
		{"OpCompareMatchFalse", newTestOpCompare(CTMatch, newTestOpConstant(NewExprValueString("no match")),
			newTestOpConstant(NewExprValueRegexpMust("[0-9]{3}"))), NewExprValueBoolean(false)},
		// constant ---------------------------------------
		{"opConstant", newTestOpConstant(NewExprValueString("foo")), NewExprValueString("foo")},
		// for ---------------------------------------
		{"OpForNoBreak", newTestOpFor(
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			newTestOpReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString)),
			nil,
			"k1"),
			NewExprValueString("bar")},
		{"OpForBreakTrue", newTestOpFor(
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			newTestOpReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString)),
			newTestOpConstant(NewExprValueString("foo")),
			"k1"),
			NewExprValueString("foo")},
		{"OpForBreakFalse", newTestOpFor(
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			newTestOpReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString)),
			newTestOpConstant(NewExprValueString("not found")),
			"k1"),
			NewExprValueString("bar")},
		{"OpForEmptyList", newTestOpFor(
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{})),
			newTestOpReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString)),
			nil,
			"k1"),
			EvNilString},
		{"OpForNilList", newTestOpFor(
			newTestOpConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))),
			newTestOpReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString)),
			nil,
			"k1"),
			EvNilString},
		// if ---------------------------------------
		{"OpIfThen", newTestOpIf(newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueString("then")), nil),
			NewExprValueString("then")},
		{"OpIfElse", newTestOpIf(newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueString("then")),
			newTestOpConstant(NewExprValueString("else"))),
			NewExprValueString("else")},
		{"OpIfElseNoElse", newTestOpIf(newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueString("then")), nil),
			EvNilString},
		{"OpIfNil", newTestOpIf(newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueString("then")), newTestOpConstant(NewExprValueString("else"))),
			EvNilString},
		// logical ---------------------------------------
		{"OpLogicalAndTrueTrue", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(true)},
		{"OpLogicalAndTrueFalse", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(false)},
		{"OpLogicalAndTrueNil", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(EvNilBoolean)), EvNilBoolean},
		{"OpLogicalAndFalseTrue", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(false)},
		{"OpLogicalAndFalseFalse", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(false)},
		{"OpLogicalAndFalseNil", newTestOpLogical(LTAnd, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(false)},
		{"OpLogicalAndNilTrue", newTestOpLogical(LTAnd, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(true))), EvNilBoolean},
		{"OpLogicalAndNilFalse", newTestOpLogical(LTAnd, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(false))), EvNilBoolean},

		{"OpLogicalOrTrueTrue", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(true)},
		{"OpLogicalOrTrueFalse", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(true)},
		{"OpLogicalOrTrueNil", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(true)),
			newTestOpConstant(EvNilBoolean)), NewExprValueBoolean(true)},
		{"OpLogicalOrFalseTrue", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueBoolean(true))), NewExprValueBoolean(true)},
		{"OpLogicalOrFalseFalse", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(NewExprValueBoolean(false))), NewExprValueBoolean(false)},
		{"OpLogicalOrFalseNil", newTestOpLogical(LTOr, newTestOpConstant(NewExprValueBoolean(false)),
			newTestOpConstant(EvNilBoolean)), EvNilBoolean},
		{"OpLogicalOrNilTrue", newTestOpLogical(LTOr, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(true))), EvNilBoolean},
		{"OpLogicalOrNilFalse", newTestOpLogical(LTOr, newTestOpConstant(EvNilBoolean),
			newTestOpConstant(NewExprValueBoolean(false))), EvNilBoolean},

		{"OpLogicalNotTrue", newTestOpLogical(LTNot, newTestOpConstant(NewExprValueBoolean(true)),
			nil), NewExprValueBoolean(false)},
		{"OpLogicalNotFalse", newTestOpLogical(LTNot, newTestOpConstant(NewExprValueBoolean(false)),
			nil), NewExprValueBoolean(true)},
		{"OpLogicalNotNil", newTestOpLogical(LTNot, newTestOpConstant(EvNilBoolean),
			nil), EvNilBoolean},
		// search ---------------------------------------
		{"OpSearchNilKey", newTestOpSearch(
			newTestOpConstant(EvNilString),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), EvNilBoolean},
		{"OpSearchNilCollection", newTestOpSearch(
			newTestOpConstant(NewExprValueString("bar")),
			newTestOpConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), EvNilBoolean},

		{"OpSearchExistListFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("bar")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), NewExprValueBoolean(true)},
		{"OpSearchExistListNotFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			})),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), NewExprValueBoolean(false)},
		{"OpSearchExistMapFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), NewExprValueBoolean(true)},
		{"OpSearchExistMapNotFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			nil, STExist, NewScalarTypeSignature(VTBoolean)), NewExprValueBoolean(false)},

		{"OpSearchFindListFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})), newTestOpConstant(NewExprValueString("baz")), STFind, NewScalarTypeSignature(VTString)),
			NewExprValueString("foo")},
		{"OpSearchFindListNotFoundDefault", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			newTestOpConstant(NewExprValueString("default")), STFind, NewScalarTypeSignature(VTString)),
			NewExprValueString("default")},
		{"OpSearchFindListNotFoundDefaultNil", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			newTestOpConstant(EvNilString), STFind, NewScalarTypeSignature(VTString)),
			EvNilString},
		{"OpSearchFindListNotFoundNoDefault", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			nil, STFind, NewScalarTypeSignature(VTString)),
			EvNilString},

		{"OpSearchFindMapFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			newTestOpConstant(NewExprValueString("baz")), STFind, NewScalarTypeSignature(VTString)),
			NewExprValueString("foo1")},
		{"OpSearchFindMapNotFoundDefault", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			newTestOpConstant(NewExprValueString("default")), STFind, NewScalarTypeSignature(VTString)),
			NewExprValueString("default")},
		{"OpSearchFindMapNotFoundDefaultNil", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			newTestOpConstant(EvNilString), STFind, NewScalarTypeSignature(VTString)),
			EvNilString},
		{"OpSearchFindMapNotFoundNoDefault", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			nil, STFind, NewScalarTypeSignature(VTString)),
			EvNilString},
		{"OpSearchFindAllListFoundSingle", newTestOpSearch(
			newTestOpConstant(NewExprValueString("bar")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("bar"),
			})},
		{"OpSearchFindAllListFoundMulti", newTestOpSearch(
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("foo"),
			})},
		{"OpSearchFindAllListNotFoundDefault", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			})),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("default"),
			})), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("default"),
			})},
		{"OpSearchFindAllMapFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo1"),
			})},
		{"OpSearchFindAllMapNotFound", newTestOpSearch(
			newTestOpConstant(NewExprValueString("not found")),
			newTestOpConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			})),
			newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})},
		// sequence ---------------------------------------
		{"OpSequence", newTestOpSequence([]operator{
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(NewExprValueBoolean(true))}),
			NewExprValueBoolean(true)},
		{"OpSequenceLastNil", newTestOpSequence([]operator{
			newTestOpConstant(NewExprValueString("foo")),
			newTestOpConstant(EvNilBoolean)}),
			EvNilBoolean},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := test.op.Evaluate(reqCtx)
			if err != nil {
				t.Errorf("unexprected evaluation error: %v", err)
				return
			}
			if !res.Equal(test.result) {
				t.Errorf("wrong evaluation result.\nactual:   %v\nexpected: %v", res, test.result)
				return
			}
		})
	}
}

func TestEvaluate_OpSearchAll_NotFoundDefaultNil(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()

	op := newTestOpSearch(
		newTestOpConstant(NewExprValueString("not found")),
		newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		})),
		newTestOpConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))),
		STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if !res.Nil() {
		t.Errorf("expected nil list as result (got %v)", res)
		return
	}
}

func TestEvaluate_OpSearchAll_NotFoundNoDefault(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()

	op := newTestOpSearch(
		newTestOpConstant(NewExprValueString("not found")),
		newTestOpConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		})),
		nil,
		STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)))
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if !res.Nil() {
		t.Errorf("expected nil list as result (got %v)", res)
		return
	}
}

// TODO test for opAssign value

func TestEvaluate_OpAssign_Heap(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()

	key := "key"
	newValue := NewExprValueString("value")

	op := newTestOpAssign("ref1", key, newTestOpConstant(newValue), nil, RSHeap)
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != newValue {
		t.Errorf("wrong evaluation result.\nactual:   %v\nexprected: %v", res, newValue)
		return
	}
	actual, err := reqCtx.Reference(key, op.resType)
	if err != nil {
		t.Errorf("unexprected reference error: %v", err)
		return
	}
	if actual != newValue {
		t.Errorf("wrong request context reference heap (key %s).\nactual:   %v\nexprected: %v", key, actual, newValue)
		return
	}
}

func TestEvaluate_OpAssign_HeapNil(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()

	key := "key"
	newValue := EvNilString

	op := newTestOpAssign("ref1", key, newTestOpConstant(newValue), nil, RSHeap)
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != newValue {
		t.Errorf("wrong evaluation result.\nactual:   %v\nexprected: %v", res, newValue)
		return
	}
	actual, err := reqCtx.Reference(key, op.resType)
	if err != nil {
		t.Errorf("unexprected reference error: %v", err)
		return
	}
	if actual != newValue {
		t.Errorf("wrong request context reference heap (key %s).\nactual:   %v\nexprected: %v", key, actual, newValue)
		return
	}
}

// TODO test for opReference value

func TestEvaluate_OpReference_Heap(t *testing.T) {
	reqCtx := newEmptyTestRequestContext()

	// Non nil value found
	key := "k1"
	value := NewExprValueString("value")
	err := reqCtx.Assign(key, value)
	if err != nil {
		t.Errorf("unexprected assignment error: %v", err)
		return
	}

	op := newTestOpReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString))
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != value {
		t.Errorf("wrong request context reference heap (key %s).\nactual:   %v\nexprected: %v", key, res, value)
		return
	}

	// Nil value found
	key = "k2"
	value = NewNilExprValue(NewScalarTypeSignature(VTString))
	err = reqCtx.Assign(key, value)
	if err != nil {
		t.Errorf("unexprected assignment error: %v", err)
		return
	}

	op = newTestOpReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString))
	res, err = op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != value {
		t.Errorf("wrong request context reference heap (key %s).\nactual:   %v\nexprected: %v", key, res, value)
		return
	}

	// Reference not found
	key = "k3"

	op = newTestOpReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString))
	res, err = op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if !res.Nil() {
		t.Errorf("Expected nil value (key %s). Got %v", key, res)
		return
	}

}
