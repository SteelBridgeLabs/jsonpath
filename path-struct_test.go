/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIdentityStructPath(t *testing.T) {
	// arrange
	value := TestArray{}
	path, _ := NewPath("")
	// act
	result := path.Evaluate(value)
	// assert
	if len(result) != 1 {
		t.Error("expected 1 result")
	}
	if diff := cmp.Diff([]any{TestArray{}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRootStructPath(t *testing.T) {
	// arrange
	value := TestArray{1, 2, 3}
	path, _ := NewPath("$")
	// act
	result := path.Evaluate(value)
	// assert
	if len(result) != 1 {
		t.Error("expected 1 result")
	}
	if diff := cmp.Diff(value, result[0]); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestDotChildStructPath1(t *testing.T) {
	// arrange
	value := TestArray{1, 2, 3}
	path, _ := NewPath("$.*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{1, 2, 3}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestDotChildStructPath2(t *testing.T) {
	// arrange
	value := TestMap{"a": "va", "b": "vb", "c": "vc"}
	path, _ := NewPath("$.*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"va", "vb", "vc"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestDotChildStructPath3(t *testing.T) {
	// arrange
	value := TestMap{"a": "test"}
	path, _ := NewPath("$.a")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentStructPath1(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test"}}
	path, _ := NewPath("$..a")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentStructPath2(t *testing.T) {
	// arrange
	value := TestArray{0, 1, TestArray{10, 11}}
	path, _ := NewPath("$..[1]")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{1, 11}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentStructPath3(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test1"}, "y": TestMap{"a": "test2"}}
	path, _ := NewPath("$..*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{TestMap{"a": "test1"}, TestMap{"a": "test2"}, "test2", "test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath1(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test1"}, "y": TestMap{"a": "test2"}}
	path, _ := NewPath("x")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{TestMap{"a": string("test1")}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath2(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test1"}, "y": TestMap{"a": "test2"}}
	path, _ := NewPath("x~")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"x"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildStructPath1(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test1"}, "y": TestMap{"a": "test2"}}
	path, _ := NewPath(`x["a"]`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildStructPath2(t *testing.T) {
	// arrange
	value := TestArray{1, 2, 3}
	path, _ := NewPath(`["1", "a"]`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildStructPath3(t *testing.T) {
	// arrange
	value := TestMap{"x": TestMap{"a": "test1"}, "y": TestMap{"a": "test2"}}
	path, _ := NewPath(`x["a"]~`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"a"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildStructPath4(t *testing.T) {
	// arrange
	value := TestArray{1, 2, 3}
	path, _ := NewPath(`["1"]~`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestFilterOnRecursiveDescentStructPath1(t *testing.T) {
	// arrange
	value := TestMap{
		"store": TestMap{
			"book": TestArray{
				TestMap{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				TestMap{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				TestMap{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				TestMap{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": TestMap{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	path, _ := NewPath(`$..book[?(@.isbn)]`)
	expected := []any{
		TestMap{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
		TestMap{
			"category": "fiction",
			"author":   "J. R. R. Tolkien",
			"title":    "The Lord of the Rings",
			"isbn":     "0-395-19395-8",
			"price":    22.99,
		},
	}
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestFilterOnRecursiveDescentStructPath2(t *testing.T) {
	// arrange
	value := TestMap{
		"store": TestMap{
			"book": TestArray{
				TestMap{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				TestMap{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				TestMap{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				TestMap{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": TestMap{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	path, err := NewPath(`$..book[?(@.author =~ /(?i).*REES/)]`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
	expected := []any{
		TestMap{
			"category": "reference",
			"author":   "Nigel Rees",
			"title":    "Sayings of the Century",
			"price":    8.95,
		},
	}
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}
