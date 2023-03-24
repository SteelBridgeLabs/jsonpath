/*
 * Copyright 2020 VMware, Inc.
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Changes:
 *   - Changed package name from github.com/vmware-labs/yamlpath to github.com/SteelBridgeLabs/jsonpath
 *   - Removed YAML implementation and added JSON implementation
 */

package jsonpath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func slice(index string, length int) ([]int, error) {
	// split index "1, 2, 3"
	if union := strings.Split(index, ","); len(union) > 1 {
		// resulting array
		combination := []int{}
		// loop over union members
		for i, idx := range union {
			// check wildcard, it cannot be used in union
			if strings.TrimSpace(idx) == "*" {
				return nil, fmt.Errorf("error in union member %d: wildcard cannot be used in union", i)
			}
			// process index @i
			sl, err := slice(idx, length)
			if err != nil {
				return nil, fmt.Errorf("error in union member %d: %s", i, err)
			}
			// append to result
			combination = append(combination, sl...)
		}
		return combination, nil
	}
	// trim
	index = strings.TrimSpace(index)
	// wildcard case "*"
	if index == "*" {
		// generate all index values [0, length], step 1
		return indices(0, length, 1, length), nil
	}
	// range case "1:2:3"
	subscr := strings.Split(index, ":")
	if len(subscr) > 3 {
		// not possible
		return nil, errors.New("malformed array index, too many colons")
	}
	type subscript struct {
		present bool
		value   int
	}
	var subscripts []subscript = []subscript{{false, 0}, {false, 0}, {false, 0}}
	const (
		sFrom = iota
		sTo
		sStep
	)
	for i, s := range subscr {
		s = strings.TrimSpace(s)
		if s != "" {
			n, err := strconv.Atoi(s)
			if err != nil {
				return nil, errors.New("non-integer array index")
			}
			subscripts[i] = subscript{
				present: true,
				value:   n,
			}
		}
	}

	// pick out the case of a single subscript first since the "to" value needs special-casing
	if len(subscr) == 1 {
		if !subscripts[sFrom].present {
			return nil, errors.New("array index missing")
		}
		from := subscripts[sFrom].value
		if from < 0 {
			from += length
		}
		return indices(from, from+1, 1, length), nil
	}

	var from, to, step int

	if subscripts[sStep].present {
		step = subscripts[sStep].value
		if step == 0 {
			return nil, errors.New("array index step value must be non-zero")
		}
	} else {
		step = 1
	}

	if subscripts[sFrom].present {
		from = subscripts[sFrom].value
		if from < 0 {
			from += length
		}
	} else {
		if step > 0 {
			from = 0
		} else {
			from = length - 1
		}
	}

	if subscripts[sTo].present {
		to = subscripts[sTo].value
		if to < 0 {
			to += length
		}
	} else {
		if step > 0 {
			to = length
		} else {
			to = -1
		}
	}

	return indices(from, to, step, length), nil
}

func indices(from, to, step, length int) []int {
	slice := []int{}
	if step > 0 {
		if from < 0 {
			from = 0 // avoid CPU attack
		}
		if to > length {
			to = length // avoid CPU attack
		}
		for i := from; i < to; i += step {
			if 0 <= i && i < length {
				slice = append(slice, i)
			}
		}
	} else if step < 0 {
		if from > length {
			from = length // avoid CPU attack
		}
		if to < -1 {
			to = -1 // avoid CPU attack
		}
		for i := from; i > to; i += step {
			if 0 <= i && i < length {
				slice = append(slice, i)
			}
		}
	}
	return slice
}
