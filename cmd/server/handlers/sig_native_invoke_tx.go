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

package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cntmio/cntmology/cmd/abi"
	clisvrcom "github.com/cntmio/cntmology/cmd/server/common"
	cliutil "github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
)

type SigNativeInvokeTxReq struct {
	GasPrice uint64        `json:"gas_price"`
	GasLimit uint64        `json:"gas_limit"`
	Address  string        `json:"address"`
	Method   string        `json:"method"`
	Params   []interface{} `json:"params"`
	Version  byte          `json:"version"`
}

type SigNativeInvokeTxRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigNativeInvokeTx(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigNativeInvokeTxReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx json.Unmarshal SigNativeInvokeTxReq:%s error:%s", req.Qid, req.Params, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	addrData, err := hex.DecodeString(rawReq.Address)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx hex.DecodeString address:%s error:%s", rawReq.Address, err)
		return
	}
	ccntmractAddr, err := common.AddressParseFromBytes(addrData)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx AddressParseFromBytes:%s error:%s", req.Qid, rawReq.Address, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	nativeAbi := abi.DefAbiMgr.GetNativeAbi(rawReq.Address)
	if nativeAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	funcAbi := nativeAbi.GetFunc(rawReq.Method)
	if funcAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	data, err := cliutil.ParseNativeParam(rawReq.Params, funcAbi.Parameters)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = fmt.Sprintf("Parse param error:%s", err)
		return
	}
	tx, err := cliutil.InvokeNativeCcntmractTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Version, ccntmractAddr, rawReq.Method, data)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx InvokeNativeCcntmractTx error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	signer := clisvrcom.DefAccount
	err = cliutil.SignTransaction(signer, tx)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = tx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigNativeInvokeTx tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigNativeInvokeTxRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
