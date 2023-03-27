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

func TestReturnNullOnMissingLeaf(t *testing.T) {
	// arrange
	var data = []any{
		map[string]any{"a": 1},
		map[string]any{"b": 2},
		map[string]any{"c": 3},
	}
	var path = "$..b"
	var expected = []any{nil, 2, nil}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult1(t *testing.T) {
	// arrange
	var data = map[string]any{"a": 1}
	var path = "$.a"
	var expected = 1
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult2(t *testing.T) {
	// arrange
	var data = map[string]any{"a": 1}
	var path = "$.a"
	var expected = []any{1}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf(), AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult3(t *testing.T) {
	// arrange
	var data = map[string]any{"a": []any{}}
	var path = "$.a"
	var expected = []any{}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult4(t *testing.T) {
	// arrange
	var data = map[string]any{"a": []any{}}
	var path = "$.a"
	var expected = []any{[]any{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf(), AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult5(t *testing.T) {
	// arrange
	var data = map[string]any{"a": 1}
	var path = "$.b"
	var expected any = nil
	// act
	result, err := Get(data, path)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult6(t *testing.T) {
	// arrange
	var data = map[string]any{"a": 1}
	var path = "$.b"
	var expected = []any{}
	// act
	result, err := Get(data, path, AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult7(t *testing.T) {
	// arrange
	var data = map[string]any{"a": []any{}}
	var path = "$..a"
	var expected = []any{[]any{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult8(t *testing.T) {
	// arrange
	var data = map[string]any{"a": []any{}}
	var path = "$..a"
	var expected = []any{[]any{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf(), AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField1(t *testing.T) {
	// arrange
	var data = map[string]any{}
	var path = "$.b"
	var expected = map[string]any{"b": 1}
	// act
	err := Set(data, path, 1)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetArrayField1(t *testing.T) {
	// arrange
	var data = []any{2}
	var path = "$[0]"
	var expected = []any{1}
	// act
	err := Set(data, path, 1)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}
