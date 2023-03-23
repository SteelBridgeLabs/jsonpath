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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFilter(t *testing.T) {
	cases := []struct {
		name      string
		filter    string
		parseTree *filterNode
		jsonDoc   string
		rootDoc   string
		match     bool
		focus     bool // if true, run only tests with focus set to true
	}{
		{
			name:      "no lexemes",
			filter:    "",
			parseTree: nil,
			jsonDoc:   "",
			rootDoc:   "",
			match:     false,
		},
		{
			name:    "existence filter, match",
			filter:  "@.category",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "existence filter, no match",
			filter:  "@.nosuch",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "@.price>8.90",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "@.price>9",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "@.price>=8.95",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "@.price>=9",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "@.price<8.96",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "@.price<8",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "@.price<=8.95",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "@.price<=8",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "8.90<@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "9<@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "8.95<=@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "9<=@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "8.96>@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "8>@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, match",
			filter:  "8.95>=@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, no match",
			filter:  "8>=@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to path, match",
			filter:  "@.x<@.y",
			jsonDoc: `{"x": 1, "y": 2}`,
			match:   true,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:    "numeric comparison filter, not found path to literal, no match",
			filter:  "@.x>=9",
			jsonDoc: `{ "category": "reference" }`,
			match:   false,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:    "numeric comparison filter, literal to not found path, no match",
			filter:  "1<@.x",
			jsonDoc: `{ "category": "reference" }`,
			match:   false,
		},
		{
			// When a filter path does not match, it produces an empty set of nodes.
			// Comparison against an empty set does not match even it matches every element of the set.
			name:    "numeric comparison filter, path to not found path, match",
			filter:  "@.x<@.y",
			jsonDoc: `{"x": 1}`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to path, mixed values, match",
			filter:  "@.x<@.y && @.y==@.z && @.y==@.w",
			jsonDoc: `{ "x": 1.1, "y": 2, "z": 2.0, "w": 2}`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, path to path, no match",
			filter:  "@.x>@.y",
			jsonDoc: `{"x": 1, "y": 2}`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, literal to literal, match",
			filter:  "8>=7",
			jsonDoc: "",
			match:   true,
		},
		{
			name:    "numeric comparison filter, literal to literal, no match",
			filter:  "8<7",
			jsonDoc: "",
			match:   false,
		},
		{
			name:    "numeric comparison filter, multiple, match",
			filter:  "@.price[*]>8.90",
			jsonDoc: `{ "price": [9, 9.5] }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, multiple, no match",
			filter:  "@.price[*]>8.90",
			jsonDoc: `{ "price": [8, 9, 9.5] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to path, single to multiple, match",
			filter:  "@.x<@.y[*]",
			jsonDoc: `{ "x": 1, "y": [2, 3] }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, path to path, single to empty set, match",
			filter:  "@.x<@.y[*]",
			jsonDoc: `{ "x": 1, "y": [] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to path, single to multiple, no match",
			filter:  "@.x<@.y[*]",
			jsonDoc: `{ "x": 4, "y": [2, 3] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to path, multiple to multiple, match",
			filter:  "@.x[*]<@.y[*]",
			jsonDoc: `{ "x": [0, 1], "y": [2, 3] }`,
			match:   true,
		},
		{
			name:    "numeric comparison filter, path to path, multiple to multiple, no match",
			filter:  "@.x[*]<@.y[*]",
			jsonDoc: `{ "x": [0, 2], "y": [2, 3] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, path to invalid path, no match",
			filter:  "@.x<@.y",
			jsonDoc: `{ "x": 4, "y": [2, 3] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, literal to invalid path, no match",
			filter:  "1<@.y",
			jsonDoc: `{ "y": [2, 3] }`,
			match:   false,
		},
		{
			name:    "numeric comparison filter, invalid path to literal, no match",
			filter:  "@.y>1",
			jsonDoc: `{ "y": [2, 3] }`,
			match:   false,
		},
		{
			// this testcase relies on an artifice of the test framework to test an edge case
			// which would normally not be reached because the lexer returns an error
			name:    "numeric comparison filter, integer to string, no match",
			filter:  "1>'x'", // produces filter parse tree with nil child
			jsonDoc: `{ "y": [2, 3] }`,
			match:   false,
		},
		{
			name:    "string comparison filter, path to path, match",
			filter:  "@.x==@.y && @x==@z",
			jsonDoc: `{ "x": "a", "y": "a", "z": "a" }`,
			match:   true,
		},
		{
			name:    "string comparison filter, path to literal, match",
			filter:  `@.x=="a"`,
			jsonDoc: `{ "x": "a" }`,
			match:   true,
		},
		{
			name:    "string comparison filter, path to path, no match",
			filter:  "@.x==@.y",
			jsonDoc: `{ "x": "a", "y": "b" }`,
			match:   false,
		},
		{
			name:    "comparison filter, string literal to numeric literal, no match",
			filter:  "'x'==7",
			jsonDoc: "",
			match:   false,
		},
		{
			name:    "comparison filter, numeric literal to string literal, no match",
			filter:  "7=='x'",
			jsonDoc: "",
			match:   false,
		},
		{
			name:    "boolean comparison filter, path to literal, match",
			filter:  `@.x==true`,
			jsonDoc: `{ "x": true }`,
			match:   true,
		},
		{
			name:    "boolean comparison filter, path to literal, no match",
			filter:  `@.x==true`,
			jsonDoc: `{ "x": "true" }`,
			match:   false,
		},
		{
			name:    "null comparison filter, path to literal, match",
			filter:  `@.x==null`,
			jsonDoc: `{ "x": null }`,
			match:   true,
		},
		{
			name:    "null comparison filter, path to literal, no match",
			filter:  `@.x==null`,
			jsonDoc: `{ "x": "null" }`,
			match:   false,
		},
		{
			name:    "null comparison filter, path to literal, match on relaxed spelling",
			filter:  `@.x==null`,
			jsonDoc: `{ "x": null }`,
			match:   true,
		},
		{
			name:    "existence || existence filter",
			filter:  "@.a || @.b",
			jsonDoc: `{ "a": 1 }`,
			match:   true,
		},
		{
			name:    "existence || existence filter",
			filter:  "@.a || @.b",
			jsonDoc: `{ "b": "x" }`,
			match:   true,
		},
		{
			name:    "existence || existence filter",
			filter:  "@.a || @.b",
			jsonDoc: `{ "c": "x" }`,
			match:   false,
		},
		{
			name:    "comparison || existence filter",
			filter:  "@.a>1 || @.b",
			jsonDoc: `{ "a": 0 }`,
			match:   false,
		},
		{
			name:    "comparison || existence filter",
			filter:  "@.a>1 || @.b",
			jsonDoc: `{ "a": 2 }`,
			match:   true,
		},
		{
			name:    "comparison || existence filter",
			filter:  "@.a>1 || @.b",
			jsonDoc: `{ "b": "x" }`,
			match:   true,
		},
		{
			name:    "existence || existence && existence filter",
			filter:  "@.a || @.b && @.c",
			jsonDoc: `{ "a": "x" }`,
			match:   true,
		},
		{
			name:    "existence || existence && existence filter",
			filter:  "@.a || @.b && @.c",
			jsonDoc: `{ "b": "x" }`,
			match:   false,
		},
		{
			name:    "existence || existence && existence filter",
			filter:  "@.a || @.b && @.c",
			jsonDoc: `{ "c": "x" }`,
			match:   false,
		},
		{
			name:    "existence || existence && existence filter",
			filter:  "@.a || @.b && @.c",
			jsonDoc: `{ "b": "x", "c": "x" }`,
			match:   true,
		},
		{
			// test just a single case of parentheses as these do not end up in the parse tree
			name:    "(existence || existence) && existence filter",
			filter:  "(@.a || @.b) && @.c",
			jsonDoc: `{ "a": "x" }`,
			match:   false,
		},
		{
			name:    "nested filter (edge case), match",
			filter:  "@.y[?(@.z==1)].w==2",
			jsonDoc: `{ "y": [ { "z": 1, "w": 2 } ] }`,
			match:   true,
		},
		{
			name:    "nested filter (edge case), no match",
			filter:  "@.y[?(@.z==5)].w==2",
			jsonDoc: `{ "y": [ { "z": 1, "w": 2 } ] }`,
			match:   false,
		},
		{
			name:    "nested filter (edge case), no match",
			filter:  "@.y[?(@.z==1)].w==4",
			jsonDoc: `{ "y": [ { "z": 1, "w": 2 } ] }`,
			match:   false,
		},
		{
			name:    "filter involving root on right, match",
			filter:  "@.price==$.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			rootDoc: `{ "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "filter involving root on left, match",
			filter:  "$.price==@.price",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			rootDoc: `{ "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "negated existence filter, no match",
			filter:  "!@.category",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "negated existence filter, match",
			filter:  "!@.nosuch",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "negated parentheses",
			filter:  "!(@.a) && @.c",
			jsonDoc: `{ "c": "x" }`,
			match:   true,
		},
		{
			name:    "regular expression filter at path, match",
			filter:  "@.category=~/ref.*ce/",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "regular expression filter at path, no match",
			filter:  "@.category=~/.*x/",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "regular expression filter root path, match",
			filter:  "$.category=~/ref.*ce/",
			rootDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   true,
		},
		{
			name:    "regular expression filter root path, no match",
			filter:  "$.category=~/.*x/",
			jsonDoc: `{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 }`,
			match:   false,
		},
		{
			name:    "literal boolean predicate",
			filter:  "true",
			rootDoc: `-1`,
			match:   true,
		},
		{
			name:    "boolean expression involving literals",
			filter:  "!false",
			rootDoc: `-1`,
			match:   true,
		},
	}

	focussed := false
	for _, tc := range cases {
		if tc.focus {
			focussed = true
			break
		}
	}

	for _, tc := range cases {
		if focussed && !tc.focus {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			n := unmarshalDoc(t, tc.jsonDoc)
			root := unmarshalDoc(t, tc.rootDoc)

			parseTree := parseFilterString(tc.filter)
			match := newFilter(parseTree)(n, root)
			require.Equal(t, tc.match, match)
		})
	}

	if focussed {
		t.Fatalf("testcase(s) still focussed")
	}
}

func unmarshalDoc(t *testing.T, doc string) any {
	// empty document
	if doc == "" {
		return []any{}
	}
	var v any
	err := json.Unmarshal([]byte(doc), &v)
	require.NoError(t, err)
	return v
}

func parseFilterString(filter string) *filterNode {
	path := fmt.Sprintf("$[?(%s)]", filter)
	lexer := lex("Path lexer", path)

	lexemes := []lexeme{}
	for {
		lexeme := lexer.nextLexeme()
		if lexeme.typ == lexemeError {
			return newFilterNode(lexemes[2:])
		}
		if lexeme.typ == lexemeEOF {
			break
		}
		lexemes = append(lexemes, lexeme)
	}

	return newFilterNode(lexemes[2 : len(lexemes)-2])
}
