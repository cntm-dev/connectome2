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

	"github.com/cntmio/cntmology/common/log"
)

type Consensus struct {
	MsgHdr
	Cons ConsensusPayload
}

//Serialize message payload
func (msg *Consensus) Serialization() ([]byte, error) {

	tmpBuffer := bytes.NewBuffer([]byte{})
	msg.Cons.Serialize(tmpBuffer)
	checkSumBuf := CheckSum(tmpBuffer.Bytes())
	msg.MsgHdr.Init("consensus", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewConsensus The message payload length is ", msg.MsgHdr.Length)

	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	err = msg.Cons.Deserialize(buf)
	return err
}