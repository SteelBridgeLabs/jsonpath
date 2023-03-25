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
	"strconv"
	"strings"
	"unicode/utf8"
)

type pathExpression func(value, root any) Iterator

// Path is a compiled JsonPath expression.
type Path struct {
	expression pathExpression
}

type pathContext struct {
	definite                 bool
	returnNullForMissingLeaf bool
	returnList               bool
}

// NewPath constructs a Path from a JsonPath expression.
func NewPath(path string) (*Path, error) {
	// create lexer
	lexer := lex("default", path)
	// create path context, use defaults
	ctx := &pathContext{}
	// create path instance
	return createPath(ctx, lexer)
}

func (p *Path) Evaluate(value any) []any {
	// evaluate path
	it := p.expression(value, value)
	// to array, never return an error here! (panic if error is returned)
	return it.ToSlice()
}

func new(expression pathExpression) *Path {
	// create path
	return &Path{
		expression: expression,
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
		return new(identity), nil

	case lexemeRoot:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// create path expression
		exp := func(value, root any) Iterator {
			// return iterator
			return compose(FromValues(false, value), subPath, root)
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
			exp := func(value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(it, allChildrenThen(subPath), root)
			}
			return new(exp), nil

		case "":
			// include all values
			exp := func(value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(it, subPath, root)
			}
			return new(exp), nil

		default:
			// include all values
			exp := func(value, root any) Iterator {
				// recursive iterator
				it := FromValues(false, value).RecurseValues()
				// compose iterator
				return compose(it, childThen(ctx, childName, subPath), root)
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
		return childThen(ctx, childName, subPath), nil

	case lexemeUndottedChild:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// process child name
		return childThen(ctx, token.val, subPath), nil

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
		return bracketChildThen(childNames, subPath), nil

	case lexemeArraySubscript:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// remove [] from token value
		subscript := strings.TrimSuffix(strings.TrimPrefix(token.val, "["), "]")
		// process subscript
		return arraySubscriptThen(subscript, subPath), nil

	case lexemeFilterBegin, lexemeRecursiveFilterBegin:
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
			return recursiveFilterThen(filterLexemes, subPath), nil
		}
		return filterThen(filterLexemes, subPath), nil

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
		return propertyNameChildThen(childName, subPath), nil

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
		return propertyNameBracketChildThen(childNames, subPath), nil

	case lexemeArraySubscriptPropertyName:
		// create sub path
		subPath, err := createPath(ctx, lexer)
		if err != nil {
			return nil, err
		}
		// trim '[' and ']~' from token value
		subscript := strings.TrimSuffix(strings.TrimPrefix(token.val, "["), "]~")
		// process property name
		return propertyNameArraySubscriptThen(subscript, subPath), nil
	}
	return nil, errors.New("invalid path expression")
}

func identity(value any, root any) Iterator {
	// return iterator
	return FromValues(false, value)
}

func empty(node, root any) Iterator {
	// emoty iterator
	return FromValues(false)
}

// evaluate path expression for all values in iterator
func compose(it Iterator, path *Path, root any) Iterator {
	// iterator slice
	its := []Iterator{}
	// iterate
	for v, ok := it(); ok; v, ok = it() {
		// append
		its = append(its, path.expression(v, root))
	}
	return FromIterators(its...)
}

func propertyNameChildThen(childName string, path *Path) *Path {
	// unescape child name
	childName = unescape(childName)
	// create path expression
	return new(func(value, root any) Iterator {
		// check value type
		switch o := value.(type) {

		case map[string]any:
			// find key in map
			if _, ok := o[childName]; ok {
				// return iterator
				return compose(FromValues(false, childName), path, root)
			}

		case MapIterator:
			// evaluate path expression on each key
			return compose(o.Keys(childName), path, root)
		}
		return empty(value, root)
	})
}

func propertyNameBracketChildThen(childNames string, path *Path) *Path {
	// "[\"a\", \"b\", \"c\"]" => ["a", "b", "c"]
	unquotedChildren := bracketChildNames(childNames)
	// create path expression
	return new(func(value, root any) Iterator {
		// check value type
		switch o := value.(type) {

		case []any:
			// iterators
			its := []Iterator{}
			// loop children
			for _, childName := range unquotedChildren {
				// convert child name to int
				i, err := strconv.Atoi(childName)
				if err != nil {
					continue
				}
				// check if index is in range
				if i >= 0 && i < len(o) {
					// append key to iterators
					its = append(its, FromValues(false, childName))
				}
			}
			// evaluate path on keys
			return compose(FromIterators(its...), path, root)

		case ArrayIterator:
			// iterators
			its := []Iterator{}
			// loop children
			for _, childName := range unquotedChildren {
				// convert child name to int
				i, err := strconv.Atoi(childName)
				if err != nil {
					continue
				}
				// check if index is in range
				if i >= 0 && i < o.Len() {
					// append key to iterators
					its = append(its, FromValues(false, childName))
				}
			}
			// evaluate path on keys
			return compose(FromIterators(its...), path, root)

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
			return compose(FromIterators(its...), path, root)

		case MapIterator:
			// evaluate path expression on keys
			return compose(o.Keys(unquotedChildren...), path, root)
		}
		return empty(value, root)
	})
}

func bracketChildThen(childNames string, path *Path) *Path {
	// "[\"a\", \"b\", \"c\"]" => ["a", "b", "c"]
	unquotedChildren := bracketChildNames(childNames)
	// iterator
	return new(func(value, root any) Iterator {
		// process value type
		switch v := value.(type) {

		case []any:
			// iterators
			its := make([]Iterator, 0, len(unquotedChildren))
			// iterate children
			for _, childName := range unquotedChildren {
				// get index
				i, err := strconv.Atoi(childName)
				if err != nil {
					continue
				}
				// check bounds
				if i >= 0 && i < len(v) {
					// append
					its = append(its, FromValues(false, v[i]))
				}
			}
			return compose(FromIterators(its...), path, root)

		case map[string]any:
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
			return compose(FromIterators(its...), path, root)

		case ArrayIterator:
			// convert unquoted children to int slice
			indices := make([]int, 0, len(unquotedChildren))
			// loop children
			for _, childName := range unquotedChildren {
				// get index
				i, err := strconv.Atoi(childName)
				if err != nil {
					// next
					continue
				}
				indices = append(indices, i)
			}
			// evaluate path expression on values @ indices
			return compose(v.Values(false, indices...), path, root)

		case MapIterator:
			// evaluate path expression on values @ keys
			return compose(v.Values(unquotedChildren...), path, root)
		}
		// empty iterator
		return empty(value, root)
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

func allChildrenThen(path *Path) *Path {
	// create path expression
	return new(func(value, root any) Iterator {
		// process value type
		switch v := value.(type) {

		case map[string]any:
			// iterators
			its := make([]Iterator, 0, len(v))
			// iterate map
			loopMap(v, func(_ string, mv any) {
				// append iterator
				its = append(its, compose(FromValues(false, mv), path, root))
			})
			return FromIterators(its...)

		case []any:
			// iterators
			its := make([]Iterator, 0, len(v))
			// loop over array
			for _, av := range v {
				// append iterator
				its = append(its, compose(FromValues(false, av), path, root))
			}
			return FromIterators(its...)

		case MapIterator:
			// evaluate path expression on each value
			return compose(v.Values(), path, root)

		case ArrayIterator:
			// evaluate path expression on each value
			return compose(v.Values(false), path, root)

		default:
			// empty
			return empty(value, root)
		}
	})
}

func arraySubscriptThen(subscript string, path *Path) *Path {
	// create path expression
	return new(func(value, root any) Iterator {
		// check wildcard
		if subscript == "*" {
			// process value type
			switch v := value.(type) {

			case []any:
				// iterators
				its := make([]Iterator, 0, len(v))
				// loop over array
				for _, av := range v {
					// append iterator
					its = append(its, compose(FromValues(false, av), path, root))
				}
				return FromIterators(its...)

			case map[string]any:
				// iterators
				its := make([]Iterator, 0, len(v))
				// iterate map
				loopMap(v, func(_ string, mv any) {
					// append iterator
					its = append(its, compose(FromValues(false, mv), path, root))
				})
				return FromIterators(its...)

			case ArrayIterator:
				// evaluate path expression on each value
				return compose(v.Values(false), path, root)

			case MapIterator:
				// evaluate path expression on each value
				return compose(v.Values(), path, root)
			}
			// empty
			return empty(value, root)
		}
		// process value type
		switch v := value.(type) {

		case []any:
			// process subscript, returns possible array indexes
			slice, err := slice(subscript, len(v))
			if err != nil {
				panic(err) // should not happen, lexer should have detected errors
			}
			// iterators
			its := make([]Iterator, 0, len(slice))
			// iterate indexes
			for _, i := range slice {
				// check index
				if i >= 0 && i < len(v) {
					// evaluate path expression on value
					its = append(its, compose(FromValues(false, v[i]), path, root))
				}
			}
			return FromIterators(its...)

		case ArrayIterator:
			// process subscript, returns possible indexes
			slice, err := slice(subscript, v.Len())
			if err != nil {
				panic(err) // should not happen, lexer should have detected errors
			}
			// evaluate path expression on values @ indexes
			return compose(v.Values(false, slice...), path, root)
		}
		// empty
		return empty(value, root)
	})
}

func filterThen(filterLexemes []lexeme, path *Path) *Path {
	// create filter from lexer tokens
	filter := newFilter(newFilterNode(filterLexemes))
	// create path expression
	return new(func(value, root any) Iterator {

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
					its = append(its, compose(FromValues(false, av), path, root))
				}
			}
			return FromIterators(its...)

		case ArrayIterator:
			// iterators
			its := make([]Iterator, 0, v.Len())
			// iterator
			it := v.Values(false)
			// loop over iterator
			for av, ok := it(); ok; av, ok = it() {
				// evaluate filter on value
				if filter(av, root) {
					// evaluate path expression on value
					its = append(its, compose(FromValues(false, av), path, root))
				}
			}
			return FromIterators(its...)

		default:
			// evaluate filter on value
			if filter(value, root) {
				// evaluate path expression on value
				return compose(FromValues(false, value), path, root)
			}
		}
		return empty(value, root)
	})
}

func propertyNameArraySubscriptThen(subscript string, path *Path) *Path {
	// create path expression
	return new(func(value, root any) Iterator {
		// check wildcard
		if subscript == "*" {
			// process value type
			switch v := value.(type) {

			case []any:
				// iterators
				its := []Iterator{}
				// loop over array indices
				for i := range v {
					// evaluate path expression on index
					its = append(its, compose(FromValues(false, i), path, root))
				}
				return FromIterators(its...)

			case ArrayIterator:
				// iterators
				its := []Iterator{}
				// loop over array indices
				for i := 0; i < v.Len(); i++ {
					// evaluate path expression on index
					its = append(its, compose(FromValues(false, i), path, root))
				}
				return FromIterators(its...)

			case map[string]any:
				// iterators
				its := []Iterator{}
				// loop over map keys
				loopMap(v, func(k string, _ any) {
					// append iterator
					its = append(its, compose(FromValues(false, k), path, root))
				})
				return FromIterators(its...)

			case MapIterator:
				// evaluate path expression on each key
				return compose(v.Keys(), path, root)
			}
		}
		return empty(value, root)
	})
}

func childThen(ctx *pathContext, childName string, path *Path) *Path {
	// check child name
	if childName == "*" {
		// all
		return allChildrenThen(path)
	}
	// process child name
	childName = unescape(childName)
	// return path
	return new(func(value, root any) Iterator {
		// check value type
		switch o := value.(type) {

		case []any:
			// convert child name to int
			i, err := strconv.Atoi(childName)
			if err != nil {
				// not an int
				return empty(value, root)
			}
			// check index
			if i >= 0 && i < len(o) {
				// evaluate path expression on value @ index
				return compose(FromValues(false, o[i]), path, root)
			}

		case map[string]any:
			// find key in map
			if v, ok := o[childName]; ok {
				// return iterator
				return compose(FromValues(false, v), path, root)
			}
			// check we need to return null for missing leaf
			if ctx.returnNullForMissingLeaf {
				// null value
				return FromValues(false, nil)
			}

		case ArrayIterator:
			// convert child name to int
			i, err := strconv.Atoi(childName)
			if err != nil {
				// not an int
				return empty(value, root)
			}
			// check index
			if i >= 0 && i < o.Len() {
				// evaluate path expression on value @ index
				return compose(o.Values(false, i), path, root)
			}

		case MapIterator:
			// check we need to return null for missing leaf
			if ctx.returnNullForMissingLeaf {
				// check key exists in map
				if _, ok := o.Keys(childName)(); !ok {
					// null value
					return FromValues(false, nil)
				}
			}
			// evaluate path expression on each value
			return compose(o.Values(childName), path, root)
		}
		return empty(value, root)
	})
}

func recursiveFilterThen(filterLexemes []lexeme, path *Path) *Path {
	// create filter
	filter := newFilter(newFilterNode(filterLexemes))
	// create path expression
	return new(func(value, root any) Iterator {
		// apply filter on value
		if filter(value, root) {
			// evaluate path expression on value
			return compose(FromValues(false, value), path, root)
		}
		return empty(value, root)
	})
}
