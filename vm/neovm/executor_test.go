/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package cntmvm

import (
	"testing"

	"github.com/conntectome/cntm/vm/cntmvm/types"
)

func BenchmarkNewExecutor(b *testing.B) {
	code := []byte{byte(PUSH1)}

	N := 50000
	for i := 0; i < N; i++ {
		code = append(code, byte(PUSH1), byte(ADD))
	}

	for i := 0; i < b.N; i++ {
		exec := NewExecutor(code, VmFeatureFlag{})
		err := exec.Execute()
		if err != nil {
			panic(err)
		}
		val, err := exec.EvalStack.PopAsIntValue()
		if err != nil {
			panic(err)
		}
		if val != types.IntValFromInt(int64(N+1)) {
			panic("wrong result")
		}
	}
}

func BenchmarkNative(b *testing.B) {

	N := 50000

	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < N; j++ {
			sum += 1
		}
	}
}
