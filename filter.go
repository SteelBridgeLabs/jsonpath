/*
 * Copyright 2020 VMware, Inc.
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Changes:
 *   - Changed package name from github.com/vmware-labs/yamlpath to github.com/SteelBridgeLabs/jsonpath
 *   - Removed YAML implementation and added JSON implementation
 */

package jsonpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type filter func(value, root any) bool

func newFilter(node *filterNode) filter {
	// check node
	if node == nil {
		return never
	}
	// process lexer token type
	switch node.lexeme.typ {

	case lexemeFilterAt, lexemeRoot:
		// create filter scanner
		path := pathFilterScanner(node)
		// return filter
		return func(value, root any) bool {
			// check path
			return len(path(value, root)) > 0
		}

	case lexemeFilterEquality, lexemeFilterInequality, lexemeFilterGreaterThan, lexemeFilterGreaterThanOrEqual, lexemeFilterLessThan, lexemeFilterLessThanOrEqual:
		// comparison filter
		return comparisonFilter(node)

	case lexemeFilterMatchesRegularExpression:
		return matchRegularExpression(node)

	case lexemeFilterNot:
		// create filter
		f := newFilter(node.children[0])
		// return filter
		return func(value, root any) bool {
			// evaluate not filter
			return !f(value, root)
		}

	case lexemeFilterOr:
		// left filter
		f1 := newFilter(node.children[0])
		// right filter
		f2 := newFilter(node.children[1])
		// return filter
		return func(value, root any) bool {
			// evaluate or filter
			return f1(value, root) || f2(value, root)
		}

	case lexemeFilterAnd:
		// left filter
		f1 := newFilter(node.children[0])
		// right filter
		f2 := newFilter(node.children[1])
		// return filter
		return func(value, root any) bool {
			// evaluate and filter
			return f1(value, root) && f2(value, root)
		}

	case lexemeFilterBooleanLiteral:
		// parse boolean literal
		b, err := strconv.ParseBool(node.lexeme.val)
		if err != nil {
			panic(err) // should not happen
		}
		// return filter
		return func(value, root any) bool {
			return b
		}

	default:
		return never
	}
}

func never(value, root any) bool {
	return false
}

func comparisonFilter(node *filterNode) filter {
	// create comparison function
	compare := func(b bool) bool {
		if b {
			// use comparator from lexer token
			return node.lexeme.comparator()(compareEqual)
		}
		// use comparator from lexer token
		return node.lexeme.comparator()(compareIncomparable)
	}
	// return filter
	return nodeToFilter(node, func(l, r typedValue) bool {
		if !l.typ.compatibleWith(r.typ) {
			return compare(false)
		}
		switch l.typ {
		case booleanValueType:
			return compare(equalBooleans(l.val, r.val))

		case nullValueType:
			return compare(equalNulls(l.val, r.val))

		default:
			return node.lexeme.comparator()(compareNodeValues(l, r))
		}
	})
}

// var x, y typedValue

// func init() {
// 	x = typedValue{stringValueType, "x"}
// 	y = typedValue{stringValueType, "y"}
// }

func nodeToFilter(node *filterNode, accept func(typedValue, typedValue) bool) filter {
	// left filter scanner
	lhsPath := newFilterScanner(node.children[0])
	// right filter scanner
	rhsPath := newFilterScanner(node.children[1])
	// create filter
	return func(value, root any) (result bool) {
		// perform a set-wise comparison of the values in each path
		match := false
		for _, l := range lhsPath(value, root) {
			for _, r := range rhsPath(value, root) {
				if !accept(l, r) {
					return false
				}
				match = true
			}
		}
		return match
	}
}

func equalBooleans(l, r string) bool {
	// Note: the YAML parser and our JSONPath lexer both rule out invalid boolean literals such as tRue.
	return strings.EqualFold(l, r)
}

func equalNulls(l, r string) bool {
	// Note: the YAML parser and our JSONPath lexer both rule out invalid null literals such as nUll.
	return true
}

// filterScanner is a function that returns a slice of typed values from either a filter literal or a path expression
// which refers to either the current node or the root node. It is used in filter comparisons.
type filterScanner func(value, root any) []typedValue

func emptyScanner(any, any) []typedValue {
	return []typedValue{}
}

func newFilterScanner(node *filterNode) filterScanner {
	switch {
	case node == nil:
		return emptyScanner

	case node.isItemFilter():
		return pathFilterScanner(node)

	case node.isLiteral():
		return literalFilterScanner(node)

	default:
		return emptyScanner
	}
}

func pathFilterScanner(node *filterNode) filterScanner {
	// should we evaluate on actual value?
	var at bool
	// process node token type
	switch node.lexeme.typ {

	case lexemeFilterAt:
		at = true

	case lexemeRoot:
		at = false

	default:
		panic("false precondition")
	}
	// all subpaths concatenated
	subpath := ""
	// loop subpaths
	for _, lexeme := range node.subpath {
		subpath += lexeme.val
	}
	// create path expression
	path, err := NewPath(subpath)
	if err != nil {
		// empty path expression
		return emptyScanner
	}
	// return path expression
	return func(value, root any) []typedValue {
		// check we need to evaluate (value)
		if at {
			return values(path.expression(getOperation, value, value))
		}
		// evaluate on root
		return values(path.expression(getOperation, root, root))
	}
}

type valueType int

const (
	unknownValueType valueType = iota
	stringValueType
	intValueType
	floatValueType
	booleanValueType
	nullValueType
	regularExpressionValueType
)

func (vt valueType) isNumeric() bool {
	return vt == intValueType || vt == floatValueType
}

func (vt valueType) compatibleWith(vt2 valueType) bool {
	return vt.isNumeric() && vt2.isNumeric() || vt == vt2 || vt == stringValueType && vt2 == regularExpressionValueType
}

type typedValue struct {
	typ valueType
	val string
}

func typedValueOfNode(value any) typedValue {
	// process value type
	switch v := value.(type) {
	case nil:
		return typedValueOfNull()
	case bool:
		return typedValueOfBool(v)
	case string:
		return typedValueOfString(v)
	case int:
		return typedValueOfInt(v)
	case int8:
		return typedValueOfInt8(v)
	case int16:
		return typedValueOfInt16(v)
	case int32:
		return typedValueOfInt32(v)
	case int64:
		return typedValueOfInt64(v)
	case float32:
		return typedValueOfFloat32(v)
	case float64:
		return typedValueOfFloat64(v)
	default:
		// unknown
		return typedValue{
			typ: unknownValueType,
			val: fmt.Sprint(value),
		}
	}
}

func newTypedValue(t valueType, v string) typedValue {
	return typedValue{
		typ: t,
		val: v,
	}
}

func typedValueOfNull() typedValue {
	return newTypedValue(nullValueType, "null")
}

func typedValueOfBool(v bool) typedValue {
	// true
	if v {
		return newTypedValue(booleanValueType, "true")
	}
	// false
	return newTypedValue(booleanValueType, "false")
}

func typedValueOfString(s string) typedValue {
	return newTypedValue(stringValueType, s)
}

func typedValueOfInt(i int) typedValue {
	return newTypedValue(intValueType, strconv.FormatInt(int64(i), 10))
}

func typedValueOfInt8(i int8) typedValue {
	return newTypedValue(intValueType, strconv.FormatInt(int64(i), 10))
}

func typedValueOfInt16(i int16) typedValue {
	return newTypedValue(intValueType, strconv.FormatInt(int64(i), 10))
}

func typedValueOfInt32(i int32) typedValue {
	return newTypedValue(intValueType, strconv.FormatInt(int64(i), 10))
}

func typedValueOfInt64(i int64) typedValue {
	return newTypedValue(intValueType, strconv.FormatInt(i, 10))
}

func typedValueOfFloat32(f float32) typedValue {
	return newTypedValue(floatValueType, strconv.FormatFloat(float64(f), 'f', -1, 32))
}

func typedValueOfFloat64(f float64) typedValue {
	return newTypedValue(floatValueType, strconv.FormatFloat(f, 'f', -1, 64))
}

func values(it Iterator) []typedValue {
	// result
	result := []typedValue{}
	// loop iterator
	for v, ok := it(); ok; v, ok = it() {
		// append typed for v
		result = append(result, typedValueOfNode(v))
	}
	return result
}

func literalFilterScanner(n *filterNode) filterScanner {
	// literal value from lexer token
	v := n.lexeme.literalValue()
	// create filter
	return func(value, root any) []typedValue {
		return []typedValue{v}
	}
}

func matchRegularExpression(parseTree *filterNode) filter {
	return nodeToFilter(parseTree, stringMatchesRegularExpression)
}

func stringMatchesRegularExpression(s, expr typedValue) bool {
	if s.typ != stringValueType || expr.typ != regularExpressionValueType {
		return false // can't compare types so return false
	}
	re, _ := regexp.Compile(expr.val) // regex already compiled during lexing
	return re.Match([]byte(s.val))
}
