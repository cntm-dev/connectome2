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

package cntm

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/serialization"
	cstates "github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	UNBOUND_TIME_OFFSET = "unboundTimeOffset"
	TOTAL_SUPPLY_NAME   = "totalSupply"
	INIT_NAME           = "init"
	TRANSFER_NAME       = "transfer"
	APPROVE_NAME        = "approve"
	TRANSFERFROM_NAME   = "transferFrom"
	NAME_NAME           = "name"
	SYMBOL_NAME         = "symbol"
	DECIMALS_NAME       = "decimals"
	TOTALSUPPLY_NAME    = "totalSupply"
	BALANCEOF_NAME      = "balanceOf"
	ALLOWANCE_NAME      = "allowance"
)

func AddNotifications(native *native.NativeService, ccntmract common.Address, state *State) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}

func GetToUInt64StorageItem(toBalance, value uint64) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteUint64(bf, toBalance+value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func GenTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], TOTAL_SUPPLY_NAME...)
}

func GenBalanceKey(ccntmract, addr common.Address) []byte {
	return append(ccntmract[:], addr[:]...)
}

func Transfer(native *native.NativeService, ccntmract common.Address, state *State) (uint64, uint64, error) {
	if !native.CcntmextRef.CheckWitness(state.From) {
		return 0, 0, errors.NewErr("authentication failed!")
	}

	fromBalance, err := fromTransfer(native, GenBalanceKey(ccntmract, state.From), state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, GenBalanceKey(ccntmract, state.To), state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func GenApproveKey(ccntmract, from, to common.Address) []byte {
	temp := append(ccntmract[:], from[:]...)
	return append(temp, to[:]...)
}

func TransferedFrom(native *native.NativeService, currentCcntmract common.Address, state *TransferFrom) (uint64, uint64, error) {
	if native.CcntmextRef.CheckWitness(state.Sender) == false {
		return 0, 0, errors.NewErr("authentication failed!")
	}

	if err := fromApprove(native, genTransferFromKey(currentCcntmract, state), state.Value); err != nil {
		return 0, 0, err
	}

	fromBalance, err := fromTransfer(native, GenBalanceKey(currentCcntmract, state.From), state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, GenBalanceKey(currentCcntmract, state.To), state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func getUnboundOffset(native *native.NativeService, ccntmract, address common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native, genAddressUnboundOffsetKey(ccntmract, address))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func genTransferFromKey(ccntmract common.Address, state *TransferFrom) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func fromApprove(native *native.NativeService, fromApproveKey []byte, value uint64) error {
	approveValue, err := utils.GetStorageUInt64(native, fromApproveKey)
	if err != nil {
		return err
	}
	if approveValue < value {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue, value)
	} else if approveValue == value {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromApproveKey, utils.GenUInt64StorageItem(approveValue-value))
	}
	return nil
}

func fromTransfer(native *native.NativeService, fromKey []byte, value uint64) (uint64, error) {
	fromBalance, err := utils.GetStorageUInt64(native, fromKey)
	if err != nil {
		return 0, err
	}
	if fromBalance < value {
		return 0, errors.NewErr("[Transfer] balance insufficient!")
	} else if fromBalance == value {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromKey, utils.GenUInt64StorageItem(fromBalance-value))
	}
	return fromBalance, nil
}

func toTransfer(native *native.NativeService, toKey []byte, value uint64) (uint64, error) {
	toBalance, err := utils.GetStorageUInt64(native, toKey)
	if err != nil {
		return 0, err
	}
	native.CloneCache.Add(scommon.ST_STORAGE, toKey, GetToUInt64StorageItem(toBalance, value))
	return toBalance, nil
}

func genAddressUnboundOffsetKey(ccntmract, address common.Address) []byte {
	temp := append(ccntmract[:], UNBOUND_TIME_OFFSET...)
	return append(temp, address[:]...)
}
