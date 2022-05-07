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

package event

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/vm/neovm/types"
)

const (
	CcntmRACT_STATE_FAIL    byte = 0
	CcntmRACT_STATE_SUCCESS byte = 1
)

// NotifyEventArgs describe smart ccntmract event notify arguments struct
type NotifyEventArgs struct {
	CcntmractAddress common.Address
	States          types.StackItems
}

// NotifyEventInfo describe smart ccntmract event notify info struct
type NotifyEventInfo struct {
	CcntmractAddress common.Address
	States          interface{}
}

type ExecuteNotify struct {
	TxHash      common.Uint256
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo
}
