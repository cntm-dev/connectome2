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
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/cntmio/cntmology/cmd/sigsvr/common"
	cliutil "github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common/log"
)

type SigDataReq struct {
	RawData string `json:"raw_data"`
}

type SigDataRsp struct {
	SignedData string `json:"signed_data"`
}

func SigData(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigDataReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	rawData, err := hex.DecodeString(rawReq.RawData)
	if err != nil {
		log.Infof("Cli Qid:%s SigData hex.DecodeString error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	signer := clisvrcom.DefAccount
	sigData, err := cliutil.Sign(rawData, signer)
	if err != nil {
		log.Infof("Cli Qid:%s SigData Sign error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigDataRsp{
		SignedData: hex.EncodeToString(sigData),
	}
}