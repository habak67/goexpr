package goexpr

import (
	"testing"
)

func TestOperator_Evaluate_Ok(t *testing.T) {
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()
	// exprReference and opReturn manipulate contexts and therefore has separate unit tests
	tests := []struct {
		name   string
		op     Expression
		result Value
	}{
		// compare ---------------------------------------
		{"OpCompareEqualTrue", NewExprCompareMust(CTEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareEqualTrueNil", NewExprCompareMust(CTEqual, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareEqualFalse", NewExprCompareMust(CTEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(false)},
		{"OpCompareEqualFalseNonNilNil", NewExprCompareMust(CTEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(false)},
		{"OpCompareEqualFalseNilNonNil", NewExprCompareMust(CTEqual, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareNotEqualTrue", NewExprCompareMust(CTNotEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(false)},
		{"OpCompareNotEqualTrueNonNilNil", NewExprCompareMust(CTNotEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareNotEqualTrueNilNonNil", NewExprCompareMust(CTNotEqual, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareNotEqualFalse", NewExprCompareMust(CTNotEqual, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareNotEqualFalseNil", NewExprCompareMust(CTNotEqual, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareNonNilNil", NewExprCompareMust(CTLess, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(EvNilInteger, l, c), l, c), EvNilBoolean},
		{"OpCompareNilNonNil", NewExprCompareMust(CTLess, NewExprConstant(EvNilInteger, l, c),
			NewExprConstant(NewExprValueInteger(1), l, c), l, c), EvNilBoolean},
		{"OpCompareNilNilLess", NewExprCompareMust(CTLess, NewExprConstant(EvNilInteger, l, c),
			NewExprConstant(EvNilInteger, l, c), l, c), EvNilBoolean},
		{"OpCompareNilNilLessEqual", NewExprCompareMust(CTLessEqual, NewExprConstant(EvNilInteger, l, c),
			NewExprConstant(EvNilInteger, l, c), l, c), EvNilBoolean},
		{"OpCompareNilNilGreater", NewExprCompareMust(CTGreater, NewExprConstant(EvNilInteger, l, c),
			NewExprConstant(EvNilInteger, l, c), l, c), EvNilBoolean},
		{"OpCompareNilNilGreaterEqual", NewExprCompareMust(CTGreaterEqual, NewExprConstant(EvNilInteger, l, c),
			NewExprConstant(EvNilInteger, l, c), l, c), EvNilBoolean},
		{"OpCompareNilNilMatch", NewExprCompareMust(CTMatch, NewExprConstant(EvNilString, l, c),
			NewExprConstant(EvNilRegexp, l, c), l, c), EvNilBoolean},

		{"OpCompareLessTrue", NewExprCompareMust(CTLess, NewExprConstant(NewExprValueInteger(1), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareLessFalse", NewExprCompareMust(CTLess, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(1), l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareLessEqualTrue", NewExprCompareMust(CTLessEqual, NewExprConstant(NewExprValueInteger(1), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareLessEqualTrue", NewExprCompareMust(CTLessEqual, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareLessEqualFalse", NewExprCompareMust(CTLessEqual, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(1), l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareGreaterTrue", NewExprCompareMust(CTGreater, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(1), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareGreaterFalse", NewExprCompareMust(CTGreater, NewExprConstant(NewExprValueInteger(1), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareGreaterEqualTrue", NewExprCompareMust(CTGreaterEqual, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(1), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareGreaterEqualTrue", NewExprCompareMust(CTGreaterEqual, NewExprConstant(NewExprValueInteger(2), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareGreaterEqualFalse", NewExprCompareMust(CTGreaterEqual, NewExprConstant(NewExprValueInteger(1), l, c),
			NewExprConstant(NewExprValueInteger(2), l, c), l, c), NewExprValueBoolean(false)},

		{"OpCompareMatchTrue", NewExprCompareMust(CTMatch, NewExprConstant(NewExprValueString("123"), l, c),
			NewExprConstant(NewExprValueRegexpMust("[0-9]{3}"), l, c), l, c), NewExprValueBoolean(true)},
		{"OpCompareMatchFalse", NewExprCompareMust(CTMatch, NewExprConstant(NewExprValueString("no match"), l, c),
			NewExprConstant(NewExprValueRegexpMust("[0-9]{3}"), l, c), l, c), NewExprValueBoolean(false)},
		{"OpCompareMatchFromConstantString", NewExprCompareMust(CTMatch, NewExprConstant(NewExprValueString("123"), l, c),
			NewExprConstant(NewExprValueString("[0-9]{3}"), l, c), l, c), NewExprValueBoolean(true)},
		// constant ---------------------------------------
		{"exprConstant", NewExprConstant(NewExprValueString("foo"), l, c), NewExprValueString("foo")},
		// for ---------------------------------------
		{"OpForNoBreak", NewExprFor(
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			NewExprReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString), l, c),
			nil,
			"k1", l, c),
			NewExprValueString("bar")},
		{"OpForBreakTrue", NewExprFor(
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			NewExprReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString), l, c),
			NewExprConstant(NewExprValueString("foo"), l, c),
			"k1", l, c),
			NewExprValueString("foo")},
		{"OpForBreakFalse", NewExprFor(
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			NewExprReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString), l, c),
			NewExprConstant(NewExprValueString("not found"), l, c),
			"k1", l, c),
			NewExprValueString("bar")},
		{"OpForEmptyList", NewExprFor(
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{}), l, c),
			NewExprReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString), l, c),
			nil,
			"k1", l, c),
			EvNilString},
		{"OpForNilList", NewExprFor(
			NewExprConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))), l, c),
			NewExprReference("loop", "k1", nil, RSHeap, NewScalarTypeSignature(VTString), l, c),
			nil,
			"k1", l, c),
			EvNilString},
		// if ---------------------------------------
		{"OpIfThen", NewExprIf(NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueString("then"), l, c), nil, l, c),
			NewExprValueString("then")},
		{"OpIfElse", NewExprIf(NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueString("then"), l, c),
			NewExprConstant(NewExprValueString("else"), l, c), l, c),
			NewExprValueString("else")},
		{"OpIfElseNoElse", NewExprIf(NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueString("then"), l, c), nil, l, c),
			EvNilString},
		{"OpIfNil", NewExprIf(NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueString("then"), l, c), NewExprConstant(NewExprValueString("else"), l, c), l, c),
			EvNilString},
		// logical ---------------------------------------
		{"OpLogicalAndTrueTrue", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(true)},
		{"OpLogicalAndTrueFalse", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(false)},
		{"OpLogicalAndTrueNil", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), EvNilBoolean},
		{"OpLogicalAndFalseTrue", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(false)},
		{"OpLogicalAndFalseFalse", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(false)},
		{"OpLogicalAndFalseNil", NewExprLogical(LTAnd, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(false)},
		{"OpLogicalAndNilTrue", NewExprLogical(LTAnd, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), EvNilBoolean},
		{"OpLogicalAndNilFalse", NewExprLogical(LTAnd, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), EvNilBoolean},

		{"OpLogicalOrTrueTrue", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(true)},
		{"OpLogicalOrTrueFalse", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(true)},
		{"OpLogicalOrTrueNil", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(true), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), NewExprValueBoolean(true)},
		{"OpLogicalOrFalseTrue", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), NewExprValueBoolean(true)},
		{"OpLogicalOrFalseFalse", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), NewExprValueBoolean(false)},
		{"OpLogicalOrFalseNil", NewExprLogical(LTOr, NewExprConstant(NewExprValueBoolean(false), l, c),
			NewExprConstant(EvNilBoolean, l, c), l, c), EvNilBoolean},
		{"OpLogicalOrNilTrue", NewExprLogical(LTOr, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c), l, c), EvNilBoolean},
		{"OpLogicalOrNilFalse", NewExprLogical(LTOr, NewExprConstant(EvNilBoolean, l, c),
			NewExprConstant(NewExprValueBoolean(false), l, c), l, c), EvNilBoolean},

		{"OpLogicalNotTrue", NewExprLogicalUnary(LTNot, NewExprConstant(NewExprValueBoolean(true), l, c),
			l, c), NewExprValueBoolean(false)},
		{"OpLogicalNotFalse", NewExprLogicalUnary(LTNot, NewExprConstant(NewExprValueBoolean(false), l, c),
			l, c), NewExprValueBoolean(true)},
		{"OpLogicalNotNil", NewExprLogicalUnary(LTNot, NewExprConstant(EvNilBoolean, l, c),
			l, c), EvNilBoolean},
		// search ---------------------------------------
		{"OpSearchNilKey", NewExprSearch(
			NewExprConstant(EvNilString, l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), EvNilBoolean},
		{"OpSearchNilCollection", NewExprSearch(
			NewExprConstant(NewExprValueString("bar"), l, c),
			NewExprConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), EvNilBoolean},

		{"OpSearchExistListFound", NewExprSearch(
			NewExprConstant(NewExprValueString("bar"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), NewExprValueBoolean(true)},
		{"OpSearchExistListNotFound", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
			}), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), NewExprValueBoolean(false)},
		{"OpSearchExistMapFound", NewExprSearch(
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), NewExprValueBoolean(true)},
		{"OpSearchExistMapNotFound", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			nil, STExist, NewScalarTypeSignature(VTBoolean), l, c), NewExprValueBoolean(false)},

		{"OpSearchFindListFound", NewExprSearch(
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c), NewExprConstant(NewExprValueString("baz"), l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			NewExprValueString("foo")},
		{"OpSearchFindListNotFoundDefault", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			NewExprConstant(NewExprValueString("default"), l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			NewExprValueString("default")},
		{"OpSearchFindListNotFoundDefaultNil", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			NewExprConstant(EvNilString, l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			EvNilString},
		{"OpSearchFindListNotFoundNoDefault", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			nil, STFind, NewScalarTypeSignature(VTString), l, c),
			EvNilString},

		{"OpSearchFindMapFound", NewExprSearch(
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			NewExprConstant(NewExprValueString("baz"), l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			NewExprValueString("foo1")},
		{"OpSearchFindMapNotFoundDefault", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			NewExprConstant(NewExprValueString("default"), l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			NewExprValueString("default")},
		{"OpSearchFindMapNotFoundDefaultNil", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			NewExprConstant(EvNilString, l, c), STFind, NewScalarTypeSignature(VTString), l, c),
			EvNilString},
		{"OpSearchFindMapNotFoundNoDefault", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			nil, STFind, NewScalarTypeSignature(VTString), l, c),
			EvNilString},
		{"OpSearchFindAllListFoundSingle", NewExprSearch(
			NewExprConstant(NewExprValueString("bar"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			}), l, c), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("bar"),
			})},
		{"OpSearchFindAllListFoundMulti", NewExprSearch(
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			}), l, c), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("foo"),
			})},
		{"OpSearchFindAllListNotFoundDefault", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo"),
				NewExprValueString("bar"),
				NewExprValueString("foo"),
			}), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("default"),
			}), l, c), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("default"),
			})},
		{"OpSearchFindAllMapFound", NewExprSearch(
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			}), l, c), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("foo1"),
			})},
		{"OpSearchFindAllMapNotFound", NewExprSearch(
			NewExprConstant(NewExprValueString("not found"), l, c),
			NewExprConstant(NewExprValueMap(NewScalarTypeSignature(VTString), map[string]Value{
				"foo": NewExprValueString("foo1"),
				"bar": NewExprValueString("bar1"),
			}), l, c),
			NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			}), l, c), STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c),
			NewExprValueList(NewScalarTypeSignature(VTString), []Value{
				NewExprValueString("baz"),
			})},
		// sequence ---------------------------------------
		{"OpSequence", NewExprSequence([]Expression{
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(NewExprValueBoolean(true), l, c)}, l, c),
			NewExprValueBoolean(true)},
		{"OpSequenceLastNil", NewExprSequence([]Expression{
			NewExprConstant(NewExprValueString("foo"), l, c),
			NewExprConstant(EvNilBoolean, l, c)}, l, c),
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
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()

	op := NewExprSearch(
		NewExprConstant(NewExprValueString("not found"), l, c),
		NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		}), l, c),
		NewExprConstant(NewNilExprValue(NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString))), l, c),
		STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c)
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
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()

	op := NewExprSearch(
		NewExprConstant(NewExprValueString("not found"), l, c),
		NewExprConstant(NewExprValueList(NewScalarTypeSignature(VTString), []Value{
			NewExprValueString("foo"),
			NewExprValueString("bar"),
			NewExprValueString("foo"),
		}), l, c),
		nil,
		STFindAll, NewCompositeTypeSignature(VTList, NewScalarTypeSignature(VTString)), l, c)
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

// TODO test for exprAssign value

func TestEvaluate_OpAssign_Heap(t *testing.T) {
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()

	key := "key"
	newValue := NewExprValueString("value")

	op := NewExprAssign("ref1", key, NewExprConstant(newValue, l, c), nil, RSHeap, l, c)
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != newValue {
		t.Errorf("wrong evaluation result.\nactual:   %v\nexprected: %v", res, newValue)
		return
	}
	actual, err := reqCtx.Reference(key, op.ResultType())
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
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()

	key := "key"
	newValue := EvNilString

	op := NewExprAssign("ref1", key, NewExprConstant(newValue, l, c), nil, RSHeap, l, c)
	res, err := op.Evaluate(reqCtx)
	if err != nil {
		t.Errorf("unexprected evaluation error: %v", err)
		return
	}
	if res != newValue {
		t.Errorf("wrong evaluation result.\nactual:   %v\nexprected: %v", res, newValue)
		return
	}
	actual, err := reqCtx.Reference(key, op.ResultType())
	if err != nil {
		t.Errorf("unexprected reference error: %v", err)
		return
	}
	if actual != newValue {
		t.Errorf("wrong request context reference heap (key %s).\nactual:   %v\nexprected: %v", key, actual, newValue)
		return
	}
}

// TODO test for exprReference value

func TestEvaluate_OpReference_Heap(t *testing.T) {
	l, c := 1, 2
	reqCtx := newEmptyTestRequestContext()

	// Non nil value found
	key := "k1"
	value := NewExprValueString("value")
	err := reqCtx.Assign(key, value)
	if err != nil {
		t.Errorf("unexprected assignment error: %v", err)
		return
	}

	op := NewExprReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString), l, c)
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

	op = NewExprReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString), l, c)
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

	op = NewExprReference("ref1", key, nil, RSHeap, NewScalarTypeSignature(VTString), l, c)
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
