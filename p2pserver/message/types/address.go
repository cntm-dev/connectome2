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

	comm "github.com/cntmio/cntmology/p2pserver/common"
)

type Addr struct {
	Hdr       MsgHdr
	NodeCnt   uint64
	NodeAddrs []comm.PeerAddr
}

//Check whether header is correct
func (msg Addr) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg Addr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg.Hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, msg.NodeCnt)
	if err != nil {
		return nil, err
	}
	for _, v := range msg.NodeAddrs {
		err = binary.Write(&buf, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.NodeCnt))
	msg.NodeAddrs = make([]comm.PeerAddr, msg.NodeCnt)
	for i := 0; i < int(msg.NodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.NodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}
