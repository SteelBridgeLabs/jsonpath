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

func TestIdentityPath(t *testing.T) {
	// arrange
	value := 1
	path, _ := NewPath("")
	// act
	result := path.Evaluate(value)
	// assert
	if len(result) != 1 {
		t.Error("expected 1 result")
	}
	if result[0] != value {
		t.Error("expected value to be returned")
	}
}

func TestRootPath(t *testing.T) {
	// arrange
	value := []any{1, 2, 3}
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

func TestDotChildPath1(t *testing.T) {
	// arrange
	value := []any{1, 2, 3}
	path, _ := NewPath("$.*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff(value, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestDotChildPath2(t *testing.T) {
	// arrange
	value := map[string]any{"a": "va", "b": "vb", "c": "vc"}
	path, _ := NewPath("$.*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"va", "vb", "vc"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestDotChildPath3(t *testing.T) {
	// arrange
	value := map[string]any{"a": "test"}
	path, _ := NewPath("$.a")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentPath1(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test"}}
	path, _ := NewPath("$..a")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentPath2(t *testing.T) {
	// arrange
	value := []any{0, 1, []any{10, 11}}
	path, _ := NewPath("$..[1]")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{1, 11}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestRecursiveDescentPath3(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, _ := NewPath("$..*")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{map[string]any{"a": "test1"}, map[string]any{"a": "test2"}, "test2", "test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildPath1(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, _ := NewPath("x")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{map[string]any{"a": string("test1")}}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestUndottedChildPath2(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, _ := NewPath("x~")
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"x"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildPath1(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, _ := NewPath(`x["a"]`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"test1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildPath2(t *testing.T) {
	// arrange
	value := []any{1, 2, 3}
	path, _ := NewPath(`["1", "a"]`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{2}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildPath3(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, _ := NewPath(`x["a"]~`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"a"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildPath4(t *testing.T) {
	// arrange
	value := []any{1, 2, 3}
	path, _ := NewPath(`["1"]~`)
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{"1"}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}
