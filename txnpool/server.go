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

// Package txnpool privides a function to start micro service txPool for
// external process
package txnpool

import (
	"github.com/cntmio/cntmology-eventbus/actor"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/events"
	"github.com/cntmio/cntmology/events/message"
	tc "github.com/cntmio/cntmology/txnpool/common"
	tp "github.com/cntmio/cntmology/txnpool/proc"
	"fmt"
)

// startActor starts an actor with the proxy and unique id,
// and return the pid.
func startActor(obj interface{}, id string) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid, _ := actor.SpawnNamed(props, id)
	if pid == nil {
		log.Error("Fail to start actor")
		return nil
	}
	return pid
}

// StartTxnPoolServer starts the txnpool server and registers
// actors to handle the msgs from the network, http, consensus
// and validators. Meanwhile subscribes the block complete  event.
func StartTxnPoolServer() (*tp.TXPoolServer, error) {
	var s *tp.TXPoolServer

	/* Start txnpool server to receive msgs from p2p,
	 * consensus and valdiators
	 */
	s = tp.NewTxPoolServer(tc.MAX_WORKER_NUM)

	// Initialize an actor to handle the msgs from valdiators
	rspActor := tp.NewVerifyRspActor(s)
	rspPid := startActor(rspActor, "txVerifyRsp")
	if rspPid == nil {
		return nil, fmt.Errorf("Start verify rsp actor failed")
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	// Initialize an actor to handle the msgs from consensus
	tpa := tp.NewTxPoolActor(s)
	txPoolPid := startActor(tpa, "txPool")
	if txPoolPid == nil {
		return nil, fmt.Errorf("Fail to start txnpool actor")
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	// Initialize an actor to handle the msgs from p2p and api
	ta := tp.NewTxActor(s)
	txPid := startActor(ta, "tx")
	if txPid == nil {
		return nil, fmt.Errorf("Fail to start txn actor")
	}
	s.RegisterActor(tc.TxActor, txPid)

	// Subscribe the block complete event
	var sub = events.NewActorSubscriber(txPoolPid)
	sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	return s, nil
}
