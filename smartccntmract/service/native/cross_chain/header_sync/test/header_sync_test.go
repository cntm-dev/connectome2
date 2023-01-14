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

package test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	vconfig "github.com/cntmio/cntmology/consensus/vbft/config"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store/leveldbstore"
	"github.com/cntmio/cntmology/core/store/overlaydb"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	cccom "github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/header_sync"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/stretchr/testify/assert"
)

var (
	acct *account.Account

	setAcct = func() {
		acct = account.NewAccount("")
	}

	generateSomeAcct = func() *account.Account {
		return account.NewAccount("")
	}

	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		if db == nil {
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		}

		return &native.NativeService{
			CacheDB: db,
			Input:   args,
			CcntmextRef: &smartccntmract.SmartCcntmract{
				Config: &smartccntmract.Config{
					Tx: &types.Transaction{
						SignedAddr: []common.Address{acct.Address},
					},
				},
			},
		}
	}

	getGenesisHeader = func() []byte {
		blkInfo := &vconfig.VbftBlockInfo{
			NewChainConfig: &vconfig.ChainConfig{
				Peers: []*vconfig.PeerConfig{
					{Index: 0, ID: hex.EncodeToString(keypair.SerializePublicKey(acct.PublicKey))},
				},
			},
		}
		payload, _ := json.Marshal(blkInfo)
		bd := &cccom.Header{
			Version:          0,
			Height:           0,
			ChainID:          0,
			Bookkeepers:      []keypair.PublicKey{acct.PublicKey},
			ConsensusPayload: payload,
			NextBookkeeper:   acct.Address,
		}
		hash := bd.Hash()
		sig, _ := signature.Sign(acct, hash[:])
		bd.SigData = [][]byte{sig}
		sink := common.NewZeroCopySink(nil)
		bd.Serialization(sink)

		return sink.Bytes()
	}

	getHeaders = func(n uint32) [][]byte {
		hdrs := make([][]byte, 0)

		blkInfo := &vconfig.VbftBlockInfo{
			NewChainConfig: &vconfig.ChainConfig{
				Peers: []*vconfig.PeerConfig{
					{Index: 0, ID: vconfig.PubkeyID(acct.PublicKey)},
				},
			},
		}
		payload, _ := json.Marshal(blkInfo)
		for i := uint32(1); i <= n; i++ {
			bd := &cccom.Header{
				Version:          0,
				Height:           i,
				ChainID:          0,
				Bookkeepers:      []keypair.PublicKey{acct.PublicKey},
				ConsensusPayload: payload,
				NextBookkeeper:   acct.Address,
			}

			hash := bd.Hash()
			sig, _ := signature.Sign(acct, hash[:])
			bd.SigData = [][]byte{sig}
			sink := common.NewZeroCopySink(nil)
			bd.Serialization(sink)
			hdrs = append(hdrs, sink.Bytes())
		}

		return hdrs
	}
)

func init() {
	setAcct()
}

func TestSyncGenesisHeader(t *testing.T) {
	// normal case: with peers
	sink := common.NewZeroCopySink(nil)
	p := &header_sync.SyncGenesisHeaderParam{
		GenesisHeader: getGenesisHeader(),
	}
	p.Serialization(sink)

	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, acct.Address)
	si := &states.StorageItem{Value: bf.Bytes()}

	ns := getNativeFunc(sink.Bytes(), nil)
	ns.CacheDB.Put(global_params.GenerateOperatorKey(utils.ParamCcntmractAddress), si.ToArray())

	ok, err := header_sync.SyncGenesisHeader(ns)
	assert.NoError(t, err)
	assert.Equal(t, utils.BYTE_TRUE, ok, "wrcntm result")

	// wrcntm owner
	ns.CcntmextRef.(*smartccntmract.SmartCcntmract).Config.Tx.SignedAddr = []common.Address{generateSomeAcct().Address}
	ok, err = header_sync.SyncGenesisHeader(ns)
	assert.EqualError(t, err, "SyncGenesisHeader, checkWitness error: validateOwner, authentication failed!",
		"not the right error")
}

func TestSyncBlockHeader(t *testing.T) {
	// first, we need to sync genesis header
	sink := common.NewZeroCopySink(nil)
	p := &header_sync.SyncGenesisHeaderParam{
		GenesisHeader: getGenesisHeader(),
	}
	p.Serialization(sink)

	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, acct.Address)
	si := &states.StorageItem{Value: bf.Bytes()}

	ns := getNativeFunc(sink.Bytes(), nil)
	ns.CacheDB.Put(global_params.GenerateOperatorKey(utils.ParamCcntmractAddress), si.ToArray())

	_, _ = header_sync.SyncGenesisHeader(ns)

	// 1. next to check normal case
	sink = common.NewZeroCopySink(nil)
	param := &header_sync.SyncBlockHeaderParam{
		Address: acct.Address,
		Headers: getHeaders(3),
	}
	param.Serialization(sink)

	ns.Input = sink.Bytes()
	ok, err := header_sync.SyncBlockHeader(ns)
	assert.NoError(t, err)
	assert.Equal(t, utils.BYTE_TRUE, ok, "wrcntm result")

	// 2.more case?
}
