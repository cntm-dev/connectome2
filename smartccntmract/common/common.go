/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/vm/neovm/types"
	"github.com/cntmio/cntmology/common/log"
)

func ConvertReturnTypes(item types.StackItems) interface{} {
	if item == nil {
		return nil
	}
	switch v := item.(type) {
	case *types.ByteArray:
		return common.ToHexString(v.GetByteArray())
	case *types.Integer:
		if item.GetBigInteger().Sign() == 0 {
			return common.ToHexString([]byte{0})
		} else {
			return common.ToHexString(types.ConvertBigIntegerToBytes(v.GetBigInteger()))
		}
	case *types.Boolean:
		if v.GetBoolean() {
			return common.ToHexString([]byte{1})
		} else {
			return common.ToHexString([]byte{0})
		}
	case *types.Array:
		var arr []interface{}
		for _, val := range v.GetArray() {
			arr = append(arr, ConvertReturnTypes(val))
		}
		return arr
	case *types.Interop:
		return common.ToHexString(v.GetInterface().ToArray())
	case types.StackItems:
		return ConvertReturnTypes(v)
	default:
		log.Error("[ConvertTypes] Invalid Types!")
		return nil
	}
}

