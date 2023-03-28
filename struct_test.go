/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

type TestArray []any

func (a TestArray) Len() int {
	return len(a)
}

func (a TestArray) Values(reverse bool, indexes ...int) Iterator {
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

func (a TestArray) Set(index int, value any) {
	a[index] = value
}

type TestMap map[string]any

func (o TestMap) Keys(keys ...string) Iterator {
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

func (o TestMap) Values(keys ...string) Iterator {
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

func (o TestMap) Set(key string, value any) {
	o[key] = value
}

func (o TestMap) Delete(key string) {
	delete(o, key)
}
