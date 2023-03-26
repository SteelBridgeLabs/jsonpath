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
	path, err := NewPath("")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$.*")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$.*")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$.a")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$..a")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$..[1]")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("$..*")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("x")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath("x~")
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath(`x["a"]`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath(`["1", "a"]`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestBracketChildPath3(t *testing.T) {
	// arrange
	value := map[string]any{"x": map[string]any{"a": "test1"}, "y": map[string]any{"a": "test2"}}
	path, err := NewPath(`x["a"]~`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
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
	path, err := NewPath(`["1"]~`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
	// act
	result := path.Evaluate(value)
	// assert
	if diff := cmp.Diff([]any{}, result); diff != "" {
		t.Errorf("invalid result: %s", diff)
	}
}

func TestFilterOnRecursiveDescentPath1(t *testing.T) {
	// arrange
	value := map[string]any{
		"store": map[string]any{
			"book": []any{
				map[string]any{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				map[string]any{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": map[string]any{
				"color": "red",
				"price": 19.95,
			},
		},
	}
	path, err := NewPath(`$..book[?(@.isbn)]`)
	if err != nil {
		t.Errorf("invalid path: %s", err)
	}
	expected := []any{
		map[string]any{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
		map[string]any{
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

func TestFilterOnRecursiveDescentPath2(t *testing.T) {
	// arrange
	value := map[string]any{
		"store": map[string]any{
			"book": []any{
				map[string]any{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				map[string]any{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": map[string]any{
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
		map[string]any{
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
