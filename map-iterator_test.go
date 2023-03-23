//go:build test

/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

import (
	"sort"
)

func loopMap(m map[string]any, callback func(k string, v any)) {
	// map keys
	keys := make([]string, 0, len(m))
	// collect map keys
	for key := range m {
		keys = append(keys, key)
	}
	// sort Keys
	sort.Strings(keys)
	// loop keys
	for _, key := range keys {
		// call func
		callback(key, m[key])
	}
}
