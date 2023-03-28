/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var knownParsingErrors = map[string]string{
	"union_with_keys_after_recursive_descent": `child name or array access or filter missing after recursive descent at position 3, following "$.."`,
}

var knownEvaluationErrors = map[string]string{}

var knownDifferences = map[string]string{
	"filter_expression_with_value_after_dot_notation_with_wildcard_on_array_of_objects": `returns [{ "key": "value" }] instead of []`,
	"filter_expression_with_equals_on_object_with_key_matching_query":                   `returns [{ "id": 2 }] instead of []`,
}

func loadTestSuite() (map[string]any, error) {
	// read file content
	content, err := os.ReadFile("testdata/regression_suite.yaml")
	if err != nil {
		return nil, err
	}
	// test suite
	var suite map[string]any
	// read yaml file
	err = yaml.Unmarshal(content, &suite)
	if err != nil {
		return nil, err
	}
	return suite, nil
}

func TestRegressionDocument(t *testing.T) {
	// load test suite
	testSuite, err := loadTestSuite()
	if err != nil {
		t.Errorf("Error loading test suite: %v", err)
	}
	// focused tests
	focused := map[string]struct{}{
		//"a": {},
	}
	// queries
	queries := testSuite["queries"].([]any)
	// loop test cases
	for _, item := range queries {
		// test case
		testCase := item.(map[string]any)
		// test case id
		id := testCase["id"].(string)
		// excluded tests
		if excluded, ok := testCase["exclude"]; ok && excluded == true {
			continue
		}
		// check focused tests
		if _, ok := focused[id]; ok {
			// execute
			executeTestCase(testCase, id, t, focused)
			// next
			continue
		}
		// others
		if len(focused) == 0 {
			// skip tests known to have parsing errors
			if _, ok := knownParsingErrors[id]; ok {
				continue
			}
			// execute
			executeTestCase(testCase, id, t, focused)
		}
	}
}

func executeTestCase(testCase map[string]any, id string, t *testing.T, focused map[string]struct{}) bool {
	// execute test case
	return t.Run(id, func(t *testing.T) {
		// selector
		selector := testCase["selector"].(string)
		// parse selector
		path, err := NewPath(selector)
		// check consensus exists
		if consensus, ok := testCase["consensus"]; ok {
			// not supported
			if consensus == "NOT_SUPPORTED" {
				// parse must fail
				require.Error(t, err, "NewPath allowed selector `%s` not be supported by the consensus", selector)
			} else {
				// parse must succeed
				require.NoError(t, err, "NewPath failed to parse selector `%s`", selector)
			}
		}
		// skip tests that failed to parse selector
		if err != nil {
			// exit
			return
		}
		// skip tests known to have evaluation errors
		if _, ok := knownEvaluationErrors[id]; ok && len(focused) == 0 {
			// exit
			return
		}
		// evaluate
		result := path.Evaluate(testCase["document"])
		// check consensus exists
		if consensus, ok := testCase["consensus"]; ok {
			// skip tests known to have differences
			if _, ok := knownDifferences[id]; ok && len(focused) == 0 {
				// exit
				return
			}
			// compare result
			compareWithConsensus(t, consensus, result)
		}
	})
}

func compareWithConsensus(t *testing.T, consensus, result any) {
	// consensus must be an array
	consensusArray, ok := consensus.([]any)
	if !ok {
		// fail
		t.Errorf("consensus is not an array: %v", consensus)
	}
	// result must be an array
	resultArray, ok := result.([]any)
	if !ok {
		// fail
		t.Errorf("result is not an array: %v", consensus)
	}
	// lengths must match
	if len(consensusArray) != len(resultArray) {
		// fail
		t.Errorf("result and consensus have different lengths: %d != %d", len(resultArray), len(consensusArray))
	}
	// create a copy of the result
	copy := append([]any(nil), resultArray...)
	// loop consensus
	for _, c := range consensusArray {
		// result must have a matching item
		for j, r := range copy {
			// compare
			if cmp.Equal(c, r) {
				// remove from result
				copy = append(copy[:j], copy[j+1:]...)
				// exit
				break
			}
		}
	}
	// copy must be empty
	if len(copy) > 0 {
		// fail
		t.Errorf("result and consensus have different items:\n%s", cmp.Diff(consensusArray, resultArray))
	}
}
