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
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	cntmErrors "github.com/Ontology/errors"
	bactor "github.com/Ontology/http/base/actor"
	"github.com/cntmio/cntmology-crypto/keypair"
)

type BalanceOfRsp struct {
	Ont string `json:"cntm"`
	Ong string `json:"cntm"`
	OngAppove string `json:"cntm_appove"`
}

type TxAttributeInfo struct {
	Usage types.TransactionAttributeUsage
	Data  string
}

type AmountMap struct {
	Key   common.Uint256
	Value common.Fixed64
}

type Fee struct {
	Amount common.Fixed64
	Payer  string
}

type Sig struct {
	PubKeys []string
	M       uint8
	SigData []string
}
type Transactions struct {
	Version    byte
	Nonce      uint32
	TxType     types.TransactionType
	Payload    PayloadInfo
	Attributes []TxAttributeInfo
	Fee        []Fee
	NetworkFee common.Fixed64
	Sigs       []Sig
	Hash       string
}

type BlockHead struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	BlockRoot        string
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	NextBookkeeper   string

	Bookkeepers []string
	SigData     []string

	Hash string
}

type BlockInfo struct {
	Hash         string
	Header       *BlockHead
	Transactions []*Transactions
}

type NodeInfo struct {
	NodeState   uint   // node status
	NodePort    uint16 // The nodes's port
	ID          uint64 // The nodes's id
	NodeTime    int64
	NodeVersion uint32   // The network protocol the node used
	NodeType    uint64   // The services the node supplied
	Relay       bool     // The relay capability of the node (merge into capbility flag)
	Height      uint32   // The node latest block height
	TxnCnt      []uint64 // The transactions be transmit by this node
	//RxTxnCnt uint64 // The transaction received by this node
}

type ConsensusInfo struct {
	// TODO
}

type TXNAttrInfo struct {
	Height  uint32
	Type    int
	ErrCode int
}

type TXNEntryInfo struct {
	Txn   Transactions  // transaction which has been verified
	Fee   int64         // Total fee per transaction
	Attrs []TXNAttrInfo // the result from each validator
}

func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.Nonce = ptx.Nonce
	trans.Payload = TransPayloadToHex(ptx.Payload)

	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
	for i, v := range ptx.Attributes {
		trans.Attributes[i].Usage = v.Usage
		trans.Attributes[i].Data = common.ToHexString(v.Data)
	}
	trans.Fee = []Fee{}
	for _, fee := range ptx.Fee {
		e := Fee{fee.Amount, common.ToHexString(fee.Payer[:])}
		trans.Fee = append(trans.Fee, e)
	}
	trans.Sigs = []Sig{}
	for _, sig := range ptx.Sigs {
		e := Sig{M: sig.M}
		for i := 0; i < len(sig.PubKeys); i++ {
			key := keypair.SerializePublicKey(sig.PubKeys[i])
			e.PubKeys = append(e.PubKeys, common.ToHexString(key))
		}
		for i := 0; i < len(sig.SigData); i++ {
			e.SigData = append(e.SigData, common.ToHexString(sig.SigData[i]))
		}
		trans.Sigs = append(trans.Sigs, e)
	}
	networkfee := ptx.GetNetworkFee()
	trans.NetworkFee = networkfee

	mhash := ptx.Hash()
	trans.Hash = common.ToHexString(mhash.ToArray())
	return trans
}

func VerifyAndSendTx(txn *types.Transaction) cntmErrors.ErrCode {
	// if transaction is verified unsucessfully then will not put it into transaction pool
	if errCode := bactor.AppendTxToPool(txn); errCode != cntmErrors.ErrNoError {
		log.Warn("Can NOT add the transaction to TxnPool")
		return errCode
	}
	return cntmErrors.ErrNoError
}

func GetBlockInfo(block *types.Block) BlockInfo {
	hash := block.Hash()
	var bookKeepers = []string{}
	var sigData = []string{}
	for i := 0; i < len(block.Header.SigData); i++ {
		s := common.ToHexString(block.Header.SigData[i])
		sigData = append(sigData, s)
	}
	for i := 0; i < len(block.Header.Bookkeepers); i++ {
		e := block.Header.Bookkeepers[i]
		key := keypair.SerializePublicKey(e)
		bookKeepers = append(bookKeepers, common.ToHexString(key))
	}

	blockHead := &BlockHead{
		Version:          block.Header.Version,
		PrevBlockHash:    common.ToHexString(block.Header.PrevBlockHash.ToArray()),
		TransactionsRoot: common.ToHexString(block.Header.TransactionsRoot.ToArray()),
		BlockRoot:        common.ToHexString(block.Header.BlockRoot.ToArray()),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		NextBookkeeper:   block.Header.NextBookkeeper.ToBase58(),
		Bookkeepers:      bookKeepers,
		SigData:          sigData,
		Hash:             common.ToHexString(hash.ToArray()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         common.ToHexString(hash.ToArray()),
		Header:       blockHead,
		Transactions: trans,
	}
	return b
}
