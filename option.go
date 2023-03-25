/*
 * Copyright 2023 SteelBridgeLabs, Inc.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonpath

type Option struct {
	setup func(ctx *pathContext)
}

func ReturnNullForMissingLeaf() Option {
	return Option{
		setup: func(ctx *pathContext) {
			ctx.returnNullForMissingLeaf = true
		},
	}
}

func AlwaysReturnList() Option {
	return Option{
		setup: func(ctx *pathContext) {
			ctx.returnList = true
		},
	}
}
