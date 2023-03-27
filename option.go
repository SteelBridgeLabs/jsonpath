/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

// Option configures the behavior of the JsonPath expression evaluation.
type Option struct {
	setup func(ctx *pathContext)
}

// ReturnNullForMissingLeaf forces the result to be null if the path is definite and the leaf value is missing.
func ReturnNullForMissingLeaf() Option {
	return Option{
		setup: func(ctx *pathContext) {
			ctx.returnNullForMissingLeaf = true
		},
	}
}

// AlwaysReturnList forces the result to be a list even if the path is definite.
func AlwaysReturnList() Option {
	return Option{
		setup: func(ctx *pathContext) {
			ctx.returnList = true
		},
	}
}
