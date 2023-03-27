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

type MyArray []any

func (a MyArray) Len() int {
	return len(a)
}

func (a MyArray) Values(reverse bool, indexes ...int) Iterator {
	// check we need specific keys
	if len(indexes) > 0 {
		//  values in map
		values := make([]any, 0, len(indexes))
		// loop indexes
		for _, i := range indexes {
			// check bounds
			if i >= 0 && i < len(a) {
				// append value
				values = append(values, a[i])
			}
		}
		return FromValues(reverse, values...)
	}
	// all values
	return FromValues(reverse, a...)
}

func (a MyArray) Set(index int, value any) {
	a[index] = value
}

type MyMap map[string]any

func (o MyMap) Keys(keys ...string) Iterator {
	// check we need specific keys
	if len(keys) > 0 {
		//  values in map
		values := make([]any, 0, len(keys))
		// loop keys
		for _, k := range keys {
			// find key in map
			if _, ok := o[k]; ok {
				// append value
				values = append(values, k)
			}
		}
		return FromValues(false, values...)
	}
	// all keys in map
	values := make([]any, 0, len(o))
	// loop keys
	loopMap(o, func(k string, _ any) {
		// append value
		values = append(values, k)
	})
	return FromValues(false, values...)
}

func (o MyMap) Values(keys ...string) Iterator {
	// check we need specific keys
	if len(keys) > 0 {
		//  values in map
		values := make([]any, 0, len(keys))
		// loop keys
		for _, k := range keys {
			// find value in map
			if mv, ok := o[k]; ok {
				// append value
				values = append(values, mv)
			}
		}
		return FromValues(false, values...)
	}
	// all values in map
	values := make([]any, 0, len(o))
	// loop keys
	loopMap(o, func(_ string, mv any) {
		// append value
		values = append(values, mv)
	})
	return FromValues(false, values...)
}

func (o MyMap) Set(key string, value any) {
	o[key] = value
}

func TestIdentityStructPath(t *testing.T) {
	// arrange
	value := MyArray{}
	path, _ := NewPath("")
	// act
	result := path.Evaluate(value)
	// assert
	if len(result) != 1 {
		t.Error("expected 1 result")
	}
	if diff := cmp.Diff([]any{MyArray{}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRootStructPath(t *testing.T) {
	// arrange
	value := MyArray{1, 2, 3}
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
	value := MyArray{1, 2, 3}
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
	value := MyMap{"a": "va", "b": "vb", "c": "vc"}
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
	value := MyMap{"a": "test"}
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
	value := MyMap{"x": MyMap{"a": "test"}}
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
	value := MyArray{0, 1, MyArray{10, 11}}
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
	value := MyMap{"x": MyMap{"a": "test1"}, "y": MyMap{"a": "test2"}}
	path, _ := NewPath("$..*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{MyMap{"a": "test1"}, MyMap{"a": "test2"}, "test2", "test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath1(t *testing.T) {
	// arrange
	value := MyMap{"x": MyMap{"a": "test1"}, "y": MyMap{"a": "test2"}}
	path, _ := NewPath("x")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{MyMap{"a": string("test1")}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath2(t *testing.T) {
	// arrange
	value := MyMap{"x": MyMap{"a": "test1"}, "y": MyMap{"a": "test2"}}
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
	value := MyMap{"x": MyMap{"a": "test1"}, "y": MyMap{"a": "test2"}}
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
	value := MyArray{1, 2, 3}
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
	value := MyMap{"x": MyMap{"a": "test1"}, "y": MyMap{"a": "test2"}}
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
	value := MyArray{1, 2, 3}
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
	value := MyMap{
		"store": MyMap{
			"book": MyArray{
				MyMap{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				MyMap{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				MyMap{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				MyMap{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": MyMap{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	path, _ := NewPath(`$..book[?(@.isbn)]`)
	expected := []any{
		MyMap{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
		MyMap{
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
	value := MyMap{
		"store": MyMap{
			"book": MyArray{
				MyMap{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				MyMap{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				MyMap{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				MyMap{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": MyMap{
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
		MyMap{
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
