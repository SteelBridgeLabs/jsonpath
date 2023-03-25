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
