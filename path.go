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
	"errors"
	"strings"
	"unicode/utf8"
)

type operation int

const (
	getOperation operation = iota
	setOperation
	deleteOperation
)

type pathExpression func(operation operation, value, root any) Iterator

type setExpression func(value any)

type deleteExpression func() error

// Path is a compiled JsonPath expression.
type Path struct {
	expression pathExpression
	terminal   bool
}

type pathContext struct {
	definite                 bool
	returnNullForMissingLeaf bool
	returnList               bool
}

// NewPath constructs a Path from a JsonPath expression.
func NewPath(path string) (*Path, error) {
	// create lexer
	lexer := lex(path)
	// create path context, use defaults
	ctx := &pathContext{}
	// create path instance
	return createPath(ctx, lexer)
}

// Evaluate evaluates the compiled JsonPath expression get operation on the given value.
func (p *Path) Evaluate(value any) []any {
	// evaluate path
	it := p.expression(getOperation, value, value)
	// to array, never return an error here! (panic if error is returned)
	return it.ToSlice()
}

func new(expression pathExpression) *Path {
	// create path
	return &Path{
		expression: expression,
		terminal:   false,
	}
}

func terminal(expression pathExpression) *Path {
	// create path
	return &Path{
		expression: expression,
		terminal:   true,
	}
}

func createPath(ctx *pathContext, lexer *lexer) (*Path, error) {
	// get next token from lexer
	token := lexer.nextLexeme()

	// process token
	switch token.typ {

	case lexemeError:
		return nil, errors.New(token.val)

	case lexemeIdentity, lexemeEOF:
		return terminal(identity), nil

	case lexemeRoot:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// create path expression
		exp := func(operation operation, value, root any) Iterator {
			// return iterator
			return compose(operation, FromValues(false, value), subPath, root)
		}
		// create path
		return new(exp), nil

	case lexemeRecursiveDescent:
		// expression is not definite
		ctx.definite = false
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// child name from lexer token
		childName := strings.TrimPrefix(token.val, "..")
		// process child name
		switch childName {

		case "*":
			// includes all values, not just mapping ones
			exp := func(operation operation, value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(operation, it, allChildrenThen(ctx, subPath), root)
			}
			return new(exp), nil

		case "":
			// include all values
			exp := func(operation operation, value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(operation, it, subPath, root)
			}
			return new(exp), nil

		default:
			// include all values
			exp := func(operation operation, value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(operation, it, childThen(ctx, childName, subPath, true), root)
			}
			return new(exp), nil
		}

	case lexemeDotChild:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// child name (remove '.')
		childName := strings.TrimPrefix(token.val, ".")
		// process child name
		return childThen(ctx, childName, subPath, false), nil

	case lexemeUndottedChild:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// process child name
		return childThen(ctx, token.val, subPath, false), nil

	case lexemeBracketChild:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// child name from lexer token
		childNames := strings.TrimSpace(token.val)
		childNames = strings.TrimSuffix(strings.TrimPrefix(childNames, "["), "]")
		childNames = strings.TrimSpace(childNames)
		// []
		return bracketChildThen(ctx, childNames, subPath, false), nil

	case lexemeArraySubscript:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// remove [] from token value
		subscript := strings.TrimSuffix(strings.TrimPrefix(token.val, "["), "]")
		// process subscript
		return arraySubscriptThen(ctx, subscript, subPath, false), nil

	case lexemeFilterBegin, lexemeRecursiveFilterBegin:
		// expression is not definite
		ctx.definite = false
		// recursive flag
		var recursive bool
		// update flag
		if token.typ == lexemeRecursiveFilterBegin {
			recursive = true
		}
		// initialize filters
		filterLexemes := []lexeme{}
		filterNestingLevel := 1
	f:
		for {
			// next lexer token
			lx := lexer.nextLexeme()
			// process token type
			switch lx.typ {

			case lexemeFilterBegin:
				filterNestingLevel++

			case lexemeFilterEnd:
				filterNestingLevel--
				if filterNestingLevel == 0 {
					break f
				}

			case lexemeError:
				return nil, errors.New(lx.val)

			case lexemeEOF:
				// should never happen as lexer should have detected an error
				return nil, errors.New("missing end of filter")
			}
			filterLexemes = append(filterLexemes, lx)
		}
		// create sub path expression
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// create recursive filter expression
		if recursive {
			return recursiveFilterThen(filterLexemes, subPath, false), nil
		}
		return filterThen(filterLexemes, subPath, false), nil

	case lexemePropertyName:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// remove '.' from lexer token
		childName := strings.TrimPrefix(token.val, ".")
		// remove '~' from child name
		childName = strings.TrimSuffix(childName, propertyName)
		// process property name
		return propertyNameChildThen(childName, subPath, false), nil

	case lexemeBracketPropertyName:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// trim token value
		childNames := strings.TrimSpace(token.val)
		// remove '~' from child name
		childNames = strings.TrimSuffix(childNames, propertyName)
		// remove brackets
		childNames = strings.TrimSuffix(strings.TrimPrefix(childNames, "["), "]")
		// trim
		childNames = strings.TrimSpace(childNames)
		// process property name
		return propertyNameBracketChildThen(ctx, childNames, subPath, false), nil

	case lexemeArraySubscriptPropertyName:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// trim '[' and ']~' from token value
		subscript := strings.TrimSuffix(strings.TrimPrefix(token.val, "["), "]~")
		// process property name
		return propertyNameArraySubscriptThen(ctx, subscript, subPath, false), nil
	}
	return nil, errors.New("invalid path expression")
}

func identity(operation operation, value any, root any) Iterator {
	// return iterator
	return FromValues(false, value)
}

func empty(operation operation, value any, root any) Iterator {
	// emoty iterator
	return FromValues(false)
}

// evaluate path expression for all values in iterator
func compose(operation operation, it Iterator, path *Path, root any) Iterator {
	// iterator slice
	its := []Iterator{}
	// iterate
	for v, ok := it(); ok; v, ok = it() {
		// append
		its = append(its, path.expression(operation, v, root))
	}
	return FromIterators(its...)
}

func propertyNameChildThen(childName string, path *Path, recursive bool) *Path {
	// unescape child name
	childName = unescape(childName)
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// check value type (must be an object)
		switch o := value.(type) {

		case map[string]any:
			// find key in map
			if _, ok := o[childName]; ok {
				// return iterator
				return compose(operation, FromValues(false, childName), path, root)
			}

		case Map:
			// evaluate path expression on each key
			return compose(operation, o.Keys(childName), path, root)
		}
		return empty(operation, value, root)
	})
}

func propertyNameBracketChildThen(ctx *pathContext, childNames string, path *Path, recursive bool) *Path {
	// "[\"a\", \"b\", \"c\"]" => ["a", "b", "c"]
	unquotedChildren := bracketChildNames(childNames)
	// check more than one child
	if len(unquotedChildren) > 1 {
		// expression is not definite
		ctx.definite = false
	}
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// check value type (only objects are allowed)
		switch o := value.(type) {

		case map[string]any:
			// iterators
			its := []Iterator{}
			// loop children
			for _, childName := range unquotedChildren {
				// find key in map
				if _, ok := o[childName]; ok {
					// append key to iterators
					its = append(its, FromValues(false, childName))
				}
			}
			// evaluate path on keys
			return compose(operation, FromIterators(its...), path, root)

		case Map:
			// check we have keys to evaluate
			if len(unquotedChildren) > 0 {
				// evaluate path expression on keys
				return compose(operation, o.Keys(unquotedChildren...), path, root)
			}
			return empty(operation, value, root)
		}
		return empty(operation, value, root)
	})
}

func bracketChildThen(ctx *pathContext, childNames string, path *Path, recursive bool) *Path {
	// "[\"a\", \"b\", \"c\"]" => ["a", "b", "c"]
	unquotedChildren := bracketChildNames(childNames)
	// check more than one child
	if len(unquotedChildren) > 1 {
		// expression is not definite
		ctx.definite = false
	}
	// iterator
	return new(func(operation operation, value, root any) Iterator {
		// process value type (it must be an object)
		switch v := value.(type) {

		case map[string]any:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := make([]any, 0, len(unquotedChildren))
					// iterate children
					for _, childName := range unquotedChildren {
						// capture key
						key := childName
						// set
						var f setExpression = func(value any) {
							// set value
							v[key] = value
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := make([]any, 0, len(unquotedChildren))
					// iterate children
					for _, childName := range unquotedChildren {
						// capture key
						key := childName
						// delete
						var f deleteExpression = func() error {
							// delete key
							delete(v, key)
							// exit
							return nil
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)
				}
			}
			// iterators
			its := make([]Iterator, 0, len(unquotedChildren))
			// iterate children
			for _, childName := range unquotedChildren {
				// find child in map
				if mv, ok := v[childName]; ok {
					// append
					its = append(its, FromValues(false, mv))
				}
			}
			return compose(operation, FromIterators(its...), path, root)

		case Map:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := make([]any, 0, len(unquotedChildren))
					// iterate children
					for _, childName := range unquotedChildren {
						// capture key
						key := childName
						// set
						var f setExpression = func(value any) {
							// set value
							v.Set(key, value)
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := make([]any, 0, len(unquotedChildren))
					// iterate children
					for _, childName := range unquotedChildren {
						// capture key
						key := childName
						// delete
						var f deleteExpression = func() error {
							// delete key
							v.Delete(key)
							// exit
							return nil
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)
				}
			}
			// check we have keys to evaluate
			if len(unquotedChildren) > 0 {
				// evaluate path expression on values @ keys
				return compose(operation, v.Values(unquotedChildren...), path, root)
			}
			return empty(operation, value, root)
		}
		// empty iterator
		return empty(operation, value, root)
	})
}

func bracketChildNames(childNames string) []string {
	// split names "[\"a\", \"b\", \"c\"]"
	tokens := strings.Split(childNames, ",")
	// reconstitute child names with embedded commas
	children := []string{}
	accum := ""
	// loop tokens
	for _, token := range tokens {
		// check for balanced quotes "' ... '" or `" ... "`
		if balanced(token, '\'') && balanced(token, '"') {
			// check we are accumulating
			if accum != "" {
				// append current
				accum += "," + token
			} else {
				// append token to result
				children = append(children, token)
				// reset accumulator
				accum = ""
			}
		} else {
			// accumulate
			if accum == "" {
				// initialize accumulator
				accum = token
			} else {
				// append to accumulator
				accum += "," + token
				// append accumulated value to result
				children = append(children, accum)
				// reset accumulator
				accum = ""
			}
		}
	}
	// check for accumulated value
	if accum != "" {
		// append last accumulated value
		children = append(children, accum)
	}
	// unquote children
	result := []string{}
	for _, token := range children {
		// trim
		token = strings.TrimSpace(token)
		// check for single or double quotes
		if strings.HasPrefix(token, "'") {
			// remove outer quotes
			token = strings.TrimSuffix(strings.TrimPrefix(token, "'"), "'")
		} else {
			// remove outer quotes
			token = strings.TrimSuffix(strings.TrimPrefix(token, `"`), `"`)
		}
		// process scaped characters
		token = unescape(token)
		// append to result
		result = append(result, token)
	}
	return result
}

// checks whether a string is balanced with respect to a given quote character
func balanced(token string, q rune) bool {
	// flags
	balanced := true
	prev := eof
	// loop over bytes
	for i := 0; i < len(token); {
		// rune @ i
		rune, width := utf8.DecodeRuneInString(token[i:])
		// advance []byte index by rune width
		i += width
		// check rune is the quote character
		if rune == q {
			// verify it is escaped "a\"b"
			if i > 0 && prev == '\\' {
				// reset prev
				prev = rune
				// not the final quote
				continue
			}
			// toggle balanced
			balanced = !balanced
		}
		prev = rune
	}
	return balanced
}

func unescape(raw string) string {
	// escaped characters flags
	esc := ""
	escaped := false
	// loop over runes
	for i := 0; i < len(raw); {
		// run @ i
		rune, width := utf8.DecodeRuneInString(raw[i:])
		// advance index
		i += width
		// check rune
		if rune == '\\' {
			// check current text is escaped
			if escaped {
				// append rune
				esc += string(rune)
			}
			// toggle escaped
			escaped = !escaped
			// next
			continue
		}
		// reset
		escaped = false
		// append escaped rune
		esc += string(rune)
	}
	return esc
}

func allChildrenThen(ctx *pathContext, path *Path) *Path {
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// process value type
		switch v := value.(type) {

		case map[string]any:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := make([]any, 0, len(v))
					// iterate map
					loopMap(v, func(k string, _ any) {
						// set
						var f setExpression = func(value any) {
							// set value
							v[k] = value
						}
						// append iterator
						expressions = append(expressions, f)
					})
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := make([]any, 0, len(v))
					// iterate map
					loopMap(v, func(k string, _ any) {
						// delete
						var f deleteExpression = func() error {
							// delete key
							delete(v, k)
							// exit
							return nil
						}
						// append iterator
						expressions = append(expressions, f)
					})
					return FromValues(false, expressions...)
				}
			}
			// iterators
			its := make([]Iterator, 0, len(v))
			// iterate map
			loopMap(v, func(_ string, mv any) {
				// append iterator
				its = append(its, compose(operation, FromValues(false, mv), path, root))
			})
			return FromIterators(its...)

		case []any:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// length
					length := len(v)
					// expressions
					expressions := make([]any, 0, length)
					// loop over array indexes
					for i := 0; i < length; i++ {
						// capture index
						index := i
						// setter
						var f setExpression = func(value any) {
							// set value
							v[index] = value
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// length
					length := len(v)
					// expressions
					expressions := make([]any, 0, length)
					// loop over array indexes (backwards)
					for i := 0; i < length; i++ {
						// delete
						var f deleteExpression = func() error {
							// delete is not supported on arrays
							return errors.New("delete is not supported on slices")
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)
				}
			}
			// evaluate path on array items
			return compose(operation, FromValues(false, v...), path, root)

		case Map:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := []any{}
					// map keys
					it := v.Keys()
					// iterate map keys
					for k, ok := it(); ok; k, ok = it() {
						// capture key
						key := k.(string)
						// set
						var f setExpression = func(value any) {
							// set value
							v.Set(key, value)
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := []any{}
					// map keys
					it := v.Keys()
					// iterate map keys
					for k, ok := it(); ok; k, ok = it() {
						// capture key
						key := k.(string)
						// delete
						var f deleteExpression = func() error {
							// delete key
							v.Delete(key)
							// exit
							return nil
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)
				}
			}
			// evaluate path expression on each value
			return compose(operation, v.Values(), path, root)

		case Array:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// length
					length := v.Len()
					// expressions
					expressions := make([]any, 0, length)
					// loop over array indexes
					for i := 0; i < length; i++ {
						// capture index
						index := i
						// setter
						var f setExpression = func(value any) {
							// set value
							v.Set(index, value)
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// length
					length := v.Len()
					// expressions
					expressions := make([]any, 0, length)
					// loop over array indexes
					for i := 0; i < length; i++ {
						// delete
						var f deleteExpression = func() error {
							// delete is not supported on arrays
							return errors.New("delete is not supported on arrays")
						}
						// append iterator
						expressions = append(expressions, f)
					}
					return FromValues(false, expressions...)
				}
			}
			// evaluate path on array items
			return compose(operation, v.Values(false), path, root)

		default:
			// empty
			return empty(operation, value, root)
		}
	})
}

func arraySubscriptThen(ctx *pathContext, subscript string, path *Path, recursive bool) *Path {
	// check for wildcard, union or range
	if subscript == "*" || strings.Contains(subscript, ",") || strings.Contains(subscript, ":") {
		// path is not definite
		ctx.definite = false
	}
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// check wildcard
		if subscript == "*" {
			// process value type
			switch v := value.(type) {

			case []any, Array:
				// process array below
				break

			case map[string]any:
				// check path is terminal
				if path.terminal {
					// process operation
					switch operation {

					case setOperation:
						// expressions
						expressions := make([]any, 0, len(v))
						// iterate map
						loopMap(v, func(k string, _ any) {
							// set
							var f setExpression = func(value any) {
								// set value
								v[k] = value
							}
							// append iterator
							expressions = append(expressions, f)
						})
						return FromValues(false, expressions...)

					case deleteOperation:
						// expressions
						expressions := make([]any, 0, len(v))
						// iterate map
						loopMap(v, func(k string, _ any) {
							// delete
							var f deleteExpression = func() error {
								// delete key
								delete(v, k)
								// exit
								return nil
							}
							// append iterator
							expressions = append(expressions, f)
						})
						return FromValues(false, expressions...)
					}
				}
				// iterators
				its := make([]Iterator, 0, len(v))
				// iterate map
				loopMap(v, func(_ string, mv any) {
					// append iterator
					its = append(its, compose(operation, FromValues(false, mv), path, root))
				})
				return FromIterators(its...)

			case Map:
				// check path is terminal
				if path.terminal {
					// process operation
					switch operation {

					case setOperation:
						// expressions
						expressions := []any{}
						// key iterator
						it := v.Keys()
						// iterate map
						for k, ok := it(); ok; k, ok = it() {
							// capture key
							key := k.(string)
							// set
							var f setExpression = func(value any) {
								// set value
								v.Set(key, value)
							}
							// append iterator
							expressions = append(expressions, f)
						}
						return FromValues(false, expressions...)

					case deleteOperation:
						// expressions
						expressions := []any{}
						// key iterator
						it := v.Keys()
						// iterate map
						for k, ok := it(); ok; k, ok = it() {
							// capture key
							key := k.(string)
							// delete
							var f deleteExpression = func() error {
								// delete key
								v.Delete(key)
								// exit
								return nil
							}
							// append iterator
							expressions = append(expressions, f)
						}
						return FromValues(false, expressions...)
					}
				}
				// evaluate path expression on each value
				return compose(operation, v.Values(), path, root)

			default:
				// empty
				return empty(operation, value, root)
			}
		}
		// process value type (at this moment we process only arrays)
		switch v := value.(type) {

		case []any:
			// process subscript, returns possible array indexes
			slice, err := slice(subscript, len(v))
			if err != nil {
				panic(err) // should not happen, lexer should have detected errors
			}
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := make([]any, 0, len(slice))
					// iterate indexes
					for _, i := range slice {
						// check index
						if i >= 0 && i < len(v) {
							// capture index
							index := i
							// setter
							var f setExpression = func(value any) {
								// set value
								v[index] = value
							}
							// append index setter
							expressions = append(expressions, f)
						}
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := make([]any, 0, len(slice))
					// iterate indexes
					for _, i := range slice {
						// check index
						if i >= 0 && i < len(v) {
							// delete
							var f deleteExpression = func() error {
								// delete is not supported on slices
								return errors.New("delete is not supported on slices")
							}
							// append index setter
							expressions = append(expressions, f)
						}
					}
					return FromValues(false, expressions...)
				}
			}
			// iterators
			its := make([]Iterator, 0, len(slice))
			// iterate indexes
			for _, i := range slice {
				// check index
				if i >= 0 && i < len(v) {
					// evaluate path expression on value
					its = append(its, compose(operation, FromValues(false, v[i]), path, root))
				}
			}
			return FromIterators(its...)

		case Array:
			// process subscript, returns possible indexes
			slice, err := slice(subscript, v.Len())
			if err != nil {
				panic(err) // should not happen, lexer should have detected errors
			}
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// expressions
					expressions := make([]any, 0, len(slice))
					// iterate indexes
					for _, i := range slice {
						// check index
						if i >= 0 && i < v.Len() {
							// capture index
							index := i
							// setter
							var f setExpression = func(value any) {
								// set value
								v.Set(index, value)
							}
							// append index setter
							expressions = append(expressions, f)
						}
					}
					return FromValues(false, expressions...)

				case deleteOperation:
					// expressions
					expressions := make([]any, 0, len(slice))
					// iterate indexes
					for _, i := range slice {
						// check index
						if i >= 0 && i < v.Len() {
							// delete
							var f deleteExpression = func() error {
								// delete is not supported on slices
								return errors.New("delete is not supported on arrays")
							}
							// append index setter
							expressions = append(expressions, f)
						}
					}
					return FromValues(false, expressions...)
				}
			}
			// check slice contain indexes
			if len(slice) > 0 {
				// evaluate path expression on values @ indexes
				return compose(operation, v.Values(false, slice...), path, root)
			}
			// empty
			return empty(operation, value, root)
		}
		// empty
		return empty(operation, value, root)
	})
}

func filterThen(filterLexemes []lexeme, path *Path, recursive bool) *Path {
	// create filter from lexer tokens
	filter := newFilter(newFilterNode(filterLexemes))
	// create path expression
	return new(func(operation operation, value, root any) Iterator {

		// process value type
		switch v := value.(type) {

		case []any:
			// iterators
			its := make([]Iterator, 0, len(v))
			// loop over array
			for _, av := range v {
				// evaluate filter on value
				if filter(av, root) {
					// evaluate path expression on value
					its = append(its, compose(operation, FromValues(false, av), path, root))
				}
			}
			return FromIterators(its...)

		case Array:
			// iterators
			its := make([]Iterator, 0, v.Len())
			// iterator
			it := v.Values(false)
			// loop over iterator
			for av, ok := it(); ok; av, ok = it() {
				// evaluate filter on value
				if filter(av, root) {
					// evaluate path expression on value
					its = append(its, compose(operation, FromValues(false, av), path, root))
				}
			}
			return FromIterators(its...)

		default:
			// evaluate filter on value
			if filter(value, root) {
				// evaluate path expression on value
				return compose(operation, FromValues(false, value), path, root)
			}
		}
		return empty(operation, value, root)
	})
}

func propertyNameArraySubscriptThen(ctx *pathContext, subscript string, path *Path, recursive bool) *Path {
	// check wildcard
	if subscript == "*" {
		// expression is not definite
		ctx.definite = false
	}
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// check wildcard
		if subscript == "*" {
			// process value type (only objects)
			switch v := value.(type) {

			case map[string]any:
				// iterators
				its := []Iterator{}
				// loop over map keys
				loopMap(v, func(k string, _ any) {
					// append iterator
					its = append(its, compose(operation, FromValues(false, k), path, root))
				})
				return FromIterators(its...)

			case Map:
				// evaluate path expression on each key
				return compose(operation, v.Keys(), path, root)
			}
		}
		return empty(operation, value, root)
	})
}

func childThen(ctx *pathContext, childName string, path *Path, recursive bool) *Path {
	// check child name
	if childName == "*" {
		// all
		return allChildrenThen(ctx, path)
	}
	// process child name
	childName = unescape(childName)
	// return path
	return new(func(operation operation, value, root any) Iterator {

		// evaluate array items
		evaluateArrayItems := func(mv any) Iterator {
			// process array items
			switch v := mv.(type) {

			case []any:
				// iterators
				its := make([]Iterator, 0, len(v)+1)
				// evaluate path expression on array
				its = append(its, compose(operation, FromValues(false, v), path, root))
				// evaluate path on slice items
				its = append(its, compose(operation, FromValues(false, v...), path, root))
				// combine iterators
				return FromIterators(its...)

			case Array:
				// iterators
				its := make([]Iterator, 0, v.Len()+1)
				// evaluate path expression on array
				its = append(its, compose(operation, FromValues(false, v), path, root))
				// evaluate path on array items
				its = append(its, compose(operation, v.Values(false), path, root))
				// combine iterators
				return FromIterators(its...)

			default:
				// return iterator
				return compose(operation, FromValues(false, mv), path, root)
			}
		}

		// check value type (it must be an object)
		switch o := value.(type) {

		case map[string]any:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// set
					var f setExpression = func(value any) {
						// set value
						o[childName] = value
					}
					// set
					return FromValues(false, f)

				case deleteOperation:
					// delete
					var f deleteExpression = func() error {
						// delete key
						delete(o, childName)
						// exit
						return nil
					}
					// set
					return FromValues(false, f)
				}
			}
			// find key in map
			if mv, ok := o[childName]; ok {
				// check we are in recursive mode and path is not terminal
				if recursive && !path.terminal {
					// evaluate array items
					return evaluateArrayItems(mv)
				}
				// return iterator
				return compose(operation, FromValues(false, mv), path, root)
			}
			// check we need to return null for missing leaf (this is a terminal path)
			if ctx.returnNullForMissingLeaf && path.terminal {
				// null value
				return FromValues(false, nil)
			}

		case Map:
			// check path is terminal
			if path.terminal {
				// process operation
				switch operation {

				case setOperation:
					// set
					var f setExpression = func(value any) {
						// set value
						o.Set(childName, value)
					}
					return FromValues(false, f)

				case deleteOperation:
					// delete
					var f deleteExpression = func() error {
						// delete key
						o.Delete(childName)
						// exit
						return nil
					}
					return FromValues(false, f)
				}
			}
			// iterator
			it := o.Values(childName)
			// find value in map
			if mv, ok := it(); ok {
				// check we are in recursive mode and path is not terminal
				if recursive && !path.terminal {
					// evaluate array items
					return evaluateArrayItems(mv)
				}
				// return iterator
				return compose(operation, FromValues(false, mv), path, root)
			}
			// check we need to return null for missing leaf (this is a terminal path)
			if ctx.returnNullForMissingLeaf && path.terminal {
				// null value
				return FromValues(false, nil)
			}
		}
		return empty(operation, value, root)
	})
}

func recursiveFilterThen(filterLexemes []lexeme, path *Path, recursive bool) *Path {
	// create filter
	filter := newFilter(newFilterNode(filterLexemes))
	// create path expression
	return new(func(operation operation, value, root any) Iterator {
		// apply filter on value
		if filter(value, root) {
			// evaluate path expression on value
			return compose(operation, FromValues(false, value), path, root)
		}
		return empty(operation, value, root)
	})
}
