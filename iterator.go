/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

type Iterator func() (any, bool)

func (it Iterator) ToSlice() []any {
	// create slice
	values := []any{}
	// iterate values
	for value, ok := it(); ok; value, ok = it() {
		// append value to slice
		values = append(values, value)
	}
	// return slice
	return values
}

func (it Iterator) RecurseValues() Iterator {
	// stack
	var stack []any
	// return iterator
	return func() (any, bool) {
		// result
		var value any
		var ok bool
		// check if stack is empty
		if len(stack) > 0 {
			// pop
			value = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			// indicate we have a value
			ok = true
		} else {
			// get next value from iterator
			value, ok = it()
			if !ok {
				// exit
				return nil, false
			}
		}
		// process value type, add values to stack if value is a container
		switch v := value.(type) {

		case []any:
			// iterate backwards (debugging and unit test consistency)
			for i := len(v) - 1; i >= 0; i-- {
				// append to stack
				stack = append(stack, v[i])
			}

		case map[string]any:
			// iterate map
			loopMap(v, func(_ string, mv any) {
				// append to stack
				stack = append(stack, mv)
			})

		case Array:
			// backwards iterator (debugging and unit test consistency)
			it := v.Values(true)
			// loop over values
			for iv, ok := it(); ok; iv, ok = it() {
				// append to stack
				stack = append(stack, iv)
			}

		case Map:
			// iterator
			it := v.Values()
			// loop over values
			for iv, ok := it(); ok; iv, ok = it() {
				// append to stack
				stack = append(stack, iv)
			}
		}
		return value, ok
	}
}

func FromValues(reverse bool, values ...any) Iterator {
	// check reverse flag
	if reverse {
		// initial index
		index := len(values) - 1
		// return iterator
		return func() (any, bool) {
			// check if index is out of bounds
			if index < 0 {
				return nil, false
			}
			// current value
			value := values[index]
			// decrement index
			index--
			// return value
			return value, true
		}
	}
	// initial index
	index := 0
	// return iterator
	return func() (any, bool) {
		// check if index is out of bounds
		if index >= len(values) {
			return nil, false
		}
		// current value
		value := values[index]
		// increment index
		index++
		// return value
		return value, true
	}
}

func FromIterators(its ...Iterator) Iterator {
	// return iterator
	return func() (any, bool) {
		// iterate
		for {
			// check if there are no more iterators
			if len(its) == 0 {
				// exit
				return nil, false
			}
			// next iterator
			next := its[0]
			// evaluate it
			vale, ok := next()
			// check if iterator is done
			if ok {
				return vale, true
			}
			// next iterator
			its = its[1:]
		}
	}
}
