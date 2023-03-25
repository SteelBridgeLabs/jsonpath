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

type Array []any

func (a Array) Len() int {
	return len(a)
}

func (a Array) Values(reverse bool, indexes ...int) Iterator {
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

type Object map[string]any

func (o Object) Keys(keys ...string) Iterator {
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

func (o Object) Values(keys ...string) Iterator {
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

func TestIdentityStructPath(t *testing.T) {
	// arrange
	value := Array{}
	path, _ := NewPath("")
	// act
	result := path.Evaluate(value)
	// assert
	if len(result) != 1 {
		t.Error("expected 1 result")
	}
	if diff := cmp.Diff([]any{Array{}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRootStructPath(t *testing.T) {
	// arrange
	value := Array{1, 2, 3}
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
	value := Array{1, 2, 3}
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
	value := Object{"a": "va", "b": "vb", "c": "vc"}
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
	value := Object{"a": "test"}
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
	value := Object{"x": Object{"a": "test"}}
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
	value := Array{0, 1, Array{10, 11}}
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
	value := Object{"x": Object{"a": "test1"}, "y": Object{"a": "test2"}}
	path, _ := NewPath("$..*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{Object{"a": "test1"}, Object{"a": "test2"}, "test2", "test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath1(t *testing.T) {
	// arrange
	value := Object{"x": Object{"a": "test1"}, "y": Object{"a": "test2"}}
	path, _ := NewPath("x")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{Object{"a": string("test1")}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildStructPath2(t *testing.T) {
	// arrange
	value := Object{"x": Object{"a": "test1"}, "y": Object{"a": "test2"}}
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
	value := Object{"x": Object{"a": "test1"}, "y": Object{"a": "test2"}}
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
	value := Array{1, 2, 3}
	path, _ := NewPath(`["1", "a"]`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{2}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildStructPath3(t *testing.T) {
	// arrange
	value := Object{"x": Object{"a": "test1"}, "y": Object{"a": "test2"}}
	path, _ := NewPath(`x["a"]~`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"a"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}
