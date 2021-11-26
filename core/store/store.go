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

package store

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	states "github.com/Ontology/core/states"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
)

// ILedgerStore provides func with store package.
type ILedgerStore interface {
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block, defaultBookKeeper []*crypto.PubKey) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block) error
	GetCurrentBlockHash() common.Uint256
	GetCurrentBlockHeight() uint32
	GetCurrentHeaderHeight() uint32
	GetCurrentHeaderHash() common.Uint256
	GetBlockHash(height uint32) common.Uint256
	GetHeaderByHash(blockHash common.Uint256) (*types.Header, error)
	GetHeaderByHeight(height uint32) (*types.Header, error)
	GetBlockByHash(blockHash common.Uint256) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error)
	IsCcntmainBlock(blockHash common.Uint256) (bool, error)
	IsCcntmainTransaction(txHash common.Uint256) (bool, error)
	GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256
	GetCcntmractState(ccntmractHash common.Address) (*payload.DeployCode, error)
	GetBookKeeperState() (*states.BookKeeperState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteCcntmract(tx *types.Transaction) ([]interface{}, error)
}
