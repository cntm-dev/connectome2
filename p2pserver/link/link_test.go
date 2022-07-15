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

package link

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/payload"
	ct "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/p2pserver/common"
	mt "github.com/cntmio/cntmology/p2pserver/message/types"
)

var (
	cliLink    *Link
	serverLink *Link
	cliChan    chan *common.MsgPayload
	serverChan chan *common.MsgPayload
	cliAddr    string
	serAddr    string
)

func init() {
	log.Init(log.Stdout)

	cliLink = NewLink()
	serverLink = NewLink()

	cliLink.id = 0x733936
	serverLink.id = 0x8274950

	cliLink.port = 50338
	serverLink.port = 50339

	cliChan = make(chan *common.MsgPayload, 100)
	serverChan = make(chan *common.MsgPayload, 100)
	//listen ip addr
	cliAddr = "127.0.0.1:50338"
	serAddr = "127.0.0.1:50339"

}

func TestNewLink(t *testing.T) {

	id := 0x74936295
	port := 40339

	if cliLink.GetID() != 0x733936 {
		t.Fatal("link GetID failed")
	}

	cliLink.SetID(uint64(id))
	if cliLink.GetID() != uint64(id) {
		t.Fatal("link SetID failed")
	}

	if cliLink.GetPort() != 50338 {
		t.Fatal("link GetPort failed")
	}

	cliLink.SetPort(uint16(port))
	if cliLink.GetPort() != uint16(port) {
		t.Fatal("link SetPort failed")
	}

	cliLink.SetChan(cliChan)
	serverLink.SetChan(serverChan)

	cliLink.UpdateRXTime(time.Now())

	msg := &common.MsgPayload{
		Id:      cliLink.id,
		Addr:    cliLink.addr,
		Payload: []byte{},
	}
	go func() {
		time.Sleep(5000000)
		cliChan <- msg
	}()

	timeout := time.NewTimer(time.Second)
	select {
	case <-cliLink.recvChan:
		t.Log("read data from channel")
	case <-timeout.C:
		timeout.Stop()
		t.Fatal("can`t read data from link channel")
	}

}

func TestUnpackBufNode(t *testing.T) {
	cliLink.SetChan(cliChan)

	msgType := "block"
	var buf []byte
	var err error

	switch msgType {
	case "addr":
		var newaddrs []common.PeerAddr
		for i := 0; i < 10000000; i++ {
			newaddrs = append(newaddrs, common.PeerAddr{
				Time: time.Now().Unix(),
				ID:   uint64(i),
			})
		}
		var addr mt.Addr
		addr.NodeAddrs = newaddrs
		buf, err = addr.Serialization()
		assert.Nil(t, err)
	case "consensuspayload":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := &mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}
		buf = payload.ToArray()
	case "consensus":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := &mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}
		consensus := mt.Consensus{
			Cons: *payload,
		}
		buf, err = consensus.Serialization()
		assert.Nil(t, err)
	case "blkheader":
		var headers []ct.Header
		blkHeader := &mt.BlkHeader{}
		for i := 0; uint32(i) < 100000000; i++ {
			header := ct.Header{}
			header.Height = uint32(i)
			header.Bookkeepers = make([]keypair.PublicKey, 0)
			header.SigData = make([][]byte, 0)
			headers = append(headers, header)
		}
		blkHeader.Cnt = uint32(len(headers))
		blkHeader.BlkHdr = headers
		buf, err = blkHeader.Serialization()
		assert.Nil(t, err)
	case "tx":
		var tx ct.Transaction
		trn := &mt.Trn{}
		sig := ct.Sig{}
		sigCnt := 100000000
		for i := 0; i < sigCnt; i++ {
			data := [][]byte{
				{byte(i)},
			}
			sig.SigData = append(sig.SigData, data...)
		}
		sigs := [1]*ct.Sig{&sig}
		tx.Payload = new(payload.DeployCode)
		tx.Sigs = sigs[:]
		trn.Txn = tx
		buf, err = trn.Serialization()
		assert.Nil(t, err)
	case "block":
		var blk ct.Block
		mBlk := &mt.Block{}
		var txs []*ct.Transaction
		header := ct.Header{}
		header.Height = uint32(1)
		header.Bookkeepers = make([]keypair.PublicKey, 0)
		header.SigData = make([][]byte, 0)
		blk.Header = &header

		for i := 0; i < 2400000; i++ {
			var tx ct.Transaction
			sig := ct.Sig{}
			sig.SigData = append(sig.SigData, [][]byte{
				{byte(1)},
			}...)
			sigs := [1]*ct.Sig{&sig}
			tx.Payload = new(payload.DeployCode)
			tx.Sigs = sigs[:]
			txs = append(txs, &tx)
		}

		blk.Transactions = txs
		mBlk.Blk = blk

		buf, err = mBlk.Serialization()
		assert.Nil(t, err)
	}

	unpackNodeBuf(cliLink, buf)
	assert.Nil(t, cliLink.conn)
	assert.Equal(t, cliLink.rxBuf.len, 0)
	assert.Nil(t, cliLink.rxBuf.p)
}
