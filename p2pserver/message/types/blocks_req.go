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

package types

import (
	"bytes"
	"encoding/binary"
	"fmt"

	comm "github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/p2pserver/common"
)

type BlocksReq struct {
	HeaderHashCount uint8
	HashStart       comm.Uint256
	HashStop        comm.Uint256
}

//Serialize message payload
func (this *BlocksReq) Serialization() ([]byte, error) {
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, this)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write  error. payload:%v", this))
	}

	return p.Bytes(), nil
}

func (this *BlocksReq) CmdType() string {
	return common.GET_BLOCKS_TYPE
}

//Deserialize message payload
func (this *BlocksReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, this)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read BlocksReq error. buf:%v", buf))
	}
	return nil
}
