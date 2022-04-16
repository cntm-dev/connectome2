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

package vbft

import (
	"fmt"
	"sync"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/consensus/vbft/config"
)

type Peer struct {
	Index          uint32
	PubKey         *keypair.PublicKey
	handShake      *peerHandshakeMsg
	LatestInfo     *peerHeartbeatMsg // latest heartbeat msg
	LastUpdateTime time.Time         // time received heartbeat from peer
	connected      bool
}

type PeerPool struct {
	lock    sync.RWMutex
	maxSize int

	server  *Server
	configs map[uint32]*vconfig.PeerConfig // peer index to peer
	IDMap   map[vconfig.NodeID]uint32

	peers map[uint32]*Peer
}

func NewPeerPool(maxSize int, server *Server) *PeerPool {
	return &PeerPool{
		maxSize: maxSize,
		server:  server,
		configs: make(map[uint32]*vconfig.PeerConfig),
		IDMap:   make(map[vconfig.NodeID]uint32),
		peers:   make(map[uint32]*Peer),
	}
}

func (pool *PeerPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.configs = make(map[uint32]*vconfig.PeerConfig)
	pool.IDMap = make(map[vconfig.NodeID]uint32)
	pool.peers = make(map[uint32]*Peer)
}

func (pool *PeerPool) isNewPeer(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if _, present := pool.peers[peerIdx]; present {
		return !pool.peers[peerIdx].connected
	}

	return true
}

func (pool *PeerPool) addPeer(config *vconfig.PeerConfig) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	peerPK, err := config.ID.Pubkey()
	if err != nil {
		return fmt.Errorf("failed to unmarshal peer pubkey: %s", err)
	}
	pool.configs[config.Index] = config
	pool.IDMap[config.ID] = config.Index
	pool.peers[config.Index] = &Peer{
		Index:          config.Index,
		PubKey:         peerPK,
		LastUpdateTime: time.Unix(0, 0),
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) getActivePeerCount() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	n := 0
	for _, p := range pool.peers {
		if p.connected {
			n++
		}
	}
	return n
}

func (pool *PeerPool) peerConnected(peerIdx uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// new peer, rather than modify
	pool.peers[peerIdx] = &Peer{
		Index:     peerIdx,
		PubKey:    pool.peers[peerIdx].PubKey,
		connected: true,
	}
	return nil
}

func (pool *PeerPool) peerDisconnected(peerIdx uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	var lastUpdateTime time.Time
	if p, present := pool.peers[peerIdx]; present {
		lastUpdateTime = p.LastUpdateTime
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		LastUpdateTime: lastUpdateTime,
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) peerHandshake(peerIdx uint32, msg *peerHandshakeMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      msg,
		LatestInfo:     pool.peers[peerIdx].LatestInfo,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) peerHeartbeat(peerIdx uint32, msg *peerHeartbeatMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      pool.peers[peerIdx].handShake,
		LatestInfo:     msg,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) getNeighbours() []*Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peers := make([]*Peer, 0)
	for _, p := range pool.peers {
		if p.connected {
			peers = append(peers, p)
		}
	}
	return peers
}

func (pool *PeerPool) GetPeerIndex(nodeId vconfig.NodeID) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
}

func (pool *PeerPool) GetPeerPubKey(peerIdx uint32) *keypair.PublicKey {
	pool.lock.RLock()
	pool.lock.RUnlock()

	if p, present := pool.peers[peerIdx]; present && p != nil {
		return p.PubKey
	}

	return nil
}

func (pool *PeerPool) isPeerAlive(peerIdx uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p := pool.peers[peerIdx]
	if p == nil || !p.connected {
		return false
	}
	if time.Now().Sub(p.LastUpdateTime) > peerHandshakeTimeout*2 {
		if p.LastUpdateTime.Unix() > 0 {
			log.Errorf("server %d: peer %d sems disconnected, %v, %v", pool.server.Index, peerIdx, time.Now(), p.LastUpdateTime)
		}
		return false
	}
	return true
}

func (pool *PeerPool) getPeer(idx uint32) *Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peer := pool.peers[idx]
	if peer != nil {
		if peer.PubKey == nil {
			peer.PubKey, _ = pool.configs[idx].ID.Pubkey()
		}
		return peer
	}

	return nil
}