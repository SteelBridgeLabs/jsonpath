/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

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
	lexer := lex("default", expression)
	// create Path
	path, err := createPath(ctx, lexer)
	if err != nil {
		return nil, err
	}
	// evaluate it
	it := path.expression(data, data)
	// collect results
	result := it.ToSlice()
	// check we need to return a list
	if ctx.returnList {
		// return result
		return result, nil
	}
	// check execution is definite and we have a single result
	if ctx.definite && len(result) == 1 {
		// return single result
		return result[0], nil
	}
	// return result
	return result, nil
}

func Set(data any, expression string, value any, options ...Option) error {
	return nil
}
