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

package actor

import (
	common2 "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/event"
	types3 "github.com/cntmio/cntmology/smartccntmract/service/evm/types"
	cstate "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
)

const (
	REQ_TIMEOUT    = 5
	ERR_ACTOR_COMM = "[http] Actor comm error: %v"
)

//GetHeaderByHeight from ledger
func GetHeaderByHeight(height uint32) (*types.Header, error) {
	return ledger.DefLedger.GetHeaderByHeight(height)
}

//GetBlockByHeight from ledger
func GetBlockByHeight(height uint32) (*types.Block, error) {
	return ledger.DefLedger.GetBlockByHeight(height)
}

//GetBlockHashFromStore from ledger
func GetBlockHashFromStore(height uint32) common.Uint256 {
	return ledger.DefLedger.GetBlockHash(height)
}

//CurrentBlockHash from ledger
func CurrentBlockHash() common.Uint256 {
	return ledger.DefLedger.GetCurrentBlockHash()
}

//GetBlockFromStore from ledger
func GetBlockFromStore(hash common.Uint256) (*types.Block, error) {
	return ledger.DefLedger.GetBlockByHash(hash)
}

//GetCurrentBlockHeight from ledger
func GetCurrentBlockHeight() uint32 {
	return ledger.DefLedger.GetCurrentBlockHeight()
}

//GetTransaction from ledger
func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	tx, _, err := ledger.DefLedger.GetTransaction(hash)
	return tx, err
}

//GetStorageItem from ledger
func GetStorageItem(address common.Address, key []byte) ([]byte, error) {
	return ledger.DefLedger.GetStorageItem(address, key)
}

//GetCcntmractStateFromStore from ledger
func GetCcntmractStateFromStore(hash common.Address) (*payload.DeployCode, error) {
	hash = updateNativeSCAddr(hash)
	return ledger.DefLedger.GetCcntmractState(hash)
}

//GetTxnWithHeightByTxHash from ledger
func GetTxnWithHeightByTxHash(hash common.Uint256) (uint32, *types.Transaction, error) {
	tx, height, err := ledger.DefLedger.GetTransaction(hash)
	return height, tx, err
}

//PreExecuteCcntmract from ledger
func PreExecuteCcntmract(tx *types.Transaction) (*cstate.PreExecResult, error) {
	return ledger.DefLedger.PreExecuteCcntmract(tx)
}

func PreExecuteCcntmractBatch(tx []*types.Transaction, atomic bool) ([]*cstate.PreExecResult, uint32, error) {
	return ledger.DefLedger.PreExecuteCcntmractBatch(tx, atomic)
}

//GetEventNotifyByTxHash from ledger
func GetEventNotifyByTxHash(txHash common.Uint256) (*event.ExecuteNotify, error) {
	return ledger.DefLedger.GetEventNotifyByTx(txHash)
}

//GetEventNotifyByHeight from ledger
func GetEventNotifyByHeight(height uint32) ([]*event.ExecuteNotify, error) {
	return ledger.DefLedger.GetEventNotifyByBlock(height)
}

//GetMerkleProof from ledger
func GetMerkleProof(proofHeight uint32, rootHeight uint32) ([]common.Uint256, error) {
	return ledger.DefLedger.GetMerkleProof(proofHeight, rootHeight)
}

func GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error) {
	return ledger.DefLedger.GetCrossChainMsg(height)
}

func GetCrossStatesProof(height uint32, key []byte) ([]byte, error) {
	return ledger.DefLedger.GetCrossStatesProof(height, key)
}

func GetEthAccount(address common2.Address) (*storage.EthAccount, error) {
	return ledger.DefLedger.GetEthAccount(address)
}

func GetEthCode(hash common2.Hash) ([]byte, error) {
	return ledger.DefLedger.GetEthCode(hash)
}

func GetEthStorage(addr common2.Address, key common2.Hash) ([]byte, error) {
	return ledger.DefLedger.GetEthState(addr, key)
}

func PreExecuteEip155Tx(msg types2.Message) (*types3.ExecutionResult, error) {
	res, err := ledger.DefLedger.PreExecuteEip155Tx(msg)
	return res, err
}
