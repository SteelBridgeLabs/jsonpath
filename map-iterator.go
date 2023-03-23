//go:build !test

/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

// the GO compiler will inline this function!
func loopMap(m map[string]any, callback func(k string, v any)) {
	// loop over map
	for k, v := range m {
		// call func
		callback(k, v)
	}
}
