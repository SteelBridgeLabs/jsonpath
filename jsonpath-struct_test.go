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

func TestReturnNullOnMissingLeafWithStruct(t *testing.T) {
	// arrange
	var data = TestArray{
		TestMap{"a": 1},
		TestMap{"b": 2},
		TestMap{"c": 3},
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

func TestDefinitiveResult1WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1}
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

func TestDefinitiveResult2WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1}
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

func TestDefinitiveResult3WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": TestArray{}}
	var path = "$.a"
	var expected = TestArray{}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult4WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": TestArray{}}
	var path = "$.a"
	var expected = []any{TestArray{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf(), AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult5WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1}
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

func TestDefinitiveResult6WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1}
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

func TestDefinitiveResult7WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": TestArray{}}
	var path = "$..a"
	var expected = []any{TestArray{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestDefinitiveResult8WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": TestArray{}}
	var path = "$..a"
	var expected = []any{TestArray{}}
	// act
	result, err := Get(data, path, ReturnNullForMissingLeaf(), AlwaysReturnList())
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField1WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{}
	var path = "$.b"
	var expected = TestMap{"b": 1}
	// act
	err := Set(data, path, 1)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField2WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1, "b": 2}
	var path = "$.*"
	var expected = TestMap{"a": 3, "b": 3}
	// act
	err := Set(data, path, 3)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField3WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1, "b": 2, "c": 3}
	var path = `$["a", "c"]`
	var expected = TestMap{"a": nil, "b": 2, "c": nil}
	// act
	err := Set(data, path, nil)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField4WithStruct(t *testing.T) {
	// arrange
	var data = TestMap{"a": 1, "b": 2, "c": 3}
	var path = `$[*]`
	var expected = TestMap{"a": nil, "b": nil, "c": nil}
	// act
	err := Set(data, path, nil)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetObjectField5WithStruct(t *testing.T) {
	// arrange
	var data = TestArray{TestMap{"a": 1}}
	var path = `$[*].*`
	var expected = TestArray{TestMap{"a": nil}}
	// act
	err := Set(data, path, nil)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetArrayField1WithStruct(t *testing.T) {
	// arrange
	var data = TestArray{2}
	var path = "$[0]"
	var expected = TestArray{1}
	// act
	err := Set(data, path, 1)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetArrayField2WithStruct(t *testing.T) {
	// arrange
	var data = TestArray{1, 1, 1}
	var path = "$.*"
	var expected = TestArray{3, 3, 3}
	// act
	err := Set(data, path, 3)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetArrayField3WithStruct(t *testing.T) {
	// arrange
	var data = TestArray{1, 2, 3}
	var path = `$[0, 2]`
	var expected = TestArray{nil, 2, nil}
	// act
	err := Set(data, path, nil)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}

func TestSetArrayField4WithStruct(t *testing.T) {
	// arrange
	var data = TestArray{1, 2, 3}
	var path = `$[*]`
	var expected = TestArray{nil, nil, nil}
	// act
	err := Set(data, path, nil)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	if diff := cmp.Diff(expected, data); diff != "" {
		t.Errorf("Unexpected result: %v", diff)
	}
}
