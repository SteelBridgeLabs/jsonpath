/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

// Gets evaluates the given JsonPath expression on the input data and returns the result.
// The result is a single value if the JsonPath expression is definite, otherwise a list.
func Get(data any, expression string, options ...Option) (any, error) {
	// initial context
	ctx := &pathContext{
		definite: true,
	}
	// process options
	for _, option := range options {
		// check option
		if option.setup != nil {
			// update context
			option.setup(ctx)
		}
	}
	// create lexer
	lexer := lex(expression)
	// create Path
	path, err := createPath(ctx, lexer)
	if err != nil {
		return nil, err
	}
	// evaluate it
	it := path.expression(getOperation, data, data)
	// collect results
	result := it.ToSlice()
	// check we need to return a list
	if ctx.returnList {
		// return result
		return result, nil
	}
	// check execution is definite
	if ctx.definite {
		// check number of values in result
		switch len(result) {
		case 0:
			return nil, nil
		case 1:
			return result[0], nil
		default:
			return result, nil
		}
	}
	// return result
	return result, nil
}

// Sets evaluates the given JsonPath expression on the input data and sets the value to all matching paths.
func Set(data any, expression string, value any, options ...Option) error {
	// initial context
	ctx := &pathContext{
		definite: true,
	}
	// process options
	for _, option := range options {
		// check option
		if option.setup != nil {
			// update context
			option.setup(ctx)
		}
	}
	// create lexer
	lexer := lex(expression)
	// create Path
	path, err := createPath(ctx, lexer)
	if err != nil {
		return err
	}
	// evaluate it
	it := path.expression(setOperation, data, data)
	// loop iterator
	for r, ok := it(); ok; r, ok = it() {
		// current iterator value must be setExpression
		if f, ok := r.(setExpression); ok {
			// set value
			f(value)
		}
	}
	return nil
}
