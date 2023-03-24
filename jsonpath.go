/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

func Get(data any, expression string, options ...Option) ([]any, error) {
	// create Path
	path, err := NewPath(expression)
	if err != nil {
		return nil, err
	}
	// evaluate it
	return path.Evaluate(data)
}

func Set(data any, expression string, value any, options ...Option) error {
	return nil
}
