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
package connect_ccntmroller

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cntmio/cntmology/p2pserver/common"
	"github.com/cntmio/cntmology/p2pserver/handshake"
	"github.com/cntmio/cntmology/p2pserver/peer"
	"github.com/scylladb/go-set/strset"
)

const INBOUND_INDEX = 0
const OUTBOUND_INDEX = 1

var ErrHandshakeSelf = errors.New("the node handshake with itself")

type connectedPeer struct {
	connectId uint64
	addr      string
	peer      *peer.PeerInfo
}

type ConnectCcntmroller struct {
	ConnCtrlOption

	selfId   *common.PeerKeyId
	peerInfo *peer.PeerInfo

	mutex                sync.Mutex
	inoutbounds          [2]*strset.Set // in/outbounds address list
	inboundListenAddress *strset.Set    // in bound listen address
	connecting           *strset.Set
	peers                map[common.PeerId]*connectedPeer // all connected peers

	ownListenAddr string
	nextConnectId uint64

	logger common.Logger
}

func NewConnectCcntmroller(peerInfo *peer.PeerInfo, keyid *common.PeerKeyId,
	option ConnCtrlOption, logger common.Logger) *ConnectCcntmroller {
	ccntmrol := &ConnectCcntmroller{
		ConnCtrlOption:       option,
		selfId:               keyid,
		peerInfo:             peerInfo,
		inoutbounds:          [2]*strset.Set{strset.New(), strset.New()},
		inboundListenAddress: strset.New(),
		connecting:           strset.New(),
		peers:                make(map[common.PeerId]*connectedPeer),
		logger:               logger,
	}

	return ccntmrol
}

func (self *ConnectCcntmroller) OwnAddress() string {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.ownListenAddr
}

func (self *ConnectCcntmroller) SetOwnAddress(listenAddr string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.ownListenAddr = listenAddr
}

func (self *ConnectCcntmroller) getConnectId() uint64 {
	return atomic.AddUint64(&self.nextConnectId, 1)
}

func (self *ConnectCcntmroller) hasInbound(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.inoutbounds[INBOUND_INDEX].Has(addr)
}

func (self *ConnectCcntmroller) OutboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[OUTBOUND_INDEX].Size())
}

func (self *ConnectCcntmroller) InboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[INBOUND_INDEX].Size())
}

func (self *ConnectCcntmroller) isBoundFull(index int) bool {
	count := self.boundsCount(index)
	if index == INBOUND_INDEX {
		return count >= self.MaxConnInBound
	}
	return count >= self.MaxConnOutBound
}

func (self *ConnectCcntmroller) boundsCount(index int) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[index].Size())
}

func (self *ConnectCcntmroller) hasBoundAddr(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	has := self.inoutbounds[INBOUND_INDEX].Has(addr) || self.inoutbounds[OUTBOUND_INDEX].Has(addr)
	has = has || self.inboundListenAddress.Has(addr)
	return has
}

func (self *ConnectCcntmroller) tryAddConnecting(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.connecting.Has(addr) {
		return false
	}
	self.connecting.Add(addr)

	return true
}

func (self *ConnectCcntmroller) removeConnecting(addr string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.connecting.Remove(addr)
}

func (self *ConnectCcntmroller) checkReservedPeers(remoteAddr string) error {
	if self.ReservedPeers.Ccntmains(remoteAddr) {
		return nil
	}

	return fmt.Errorf("the remote addr: %s not in reserved list", remoteAddr)
}

func (self *ConnectCcntmroller) getInboundCountWithIp(ip string) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	var count uint
	self.inoutbounds[INBOUND_INDEX].Each(func(addr string) bool {
		ipRecord, _ := common.ParseIPAddr(addr)
		if ipRecord == ip {
			count += 1
		}

		return true
	})

	return count
}

func (self *ConnectCcntmroller) AcceptConnect(conn net.Conn) (*peer.PeerInfo, net.Conn, error) {
	addr := conn.RemoteAddr().String()
	err := self.beforeHandshakeCheck(addr, INBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeServer(self.peerInfo, self.selfId, conn)
	if err != nil {
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, addr)
	if err != nil {
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, INBOUND_INDEX)

	self.logger.Infof("inbound peer %s connected, %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

//Connect used to connect net address under sync or cons mode
// need call Peer.Close to clean up resource.
func (self *ConnectCcntmroller) Connect(addr string) (*peer.PeerInfo, net.Conn, error) {
	err := self.beforeHandshakeCheck(addr, OUTBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	if !self.tryAddConnecting(addr) {
		return nil, nil, fmt.Errorf("node exist in connecting list: %s", addr)
	}
	defer self.removeConnecting(addr)

	conn, err := self.dialer.Dial(addr)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeClient(self.peerInfo, self.selfId, conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, OUTBOUND_INDEX)

	self.logger.Infof("outbound peer %s connected. %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

func (self *ConnectCcntmroller) afterHandshakeCheck(remotePeer *peer.PeerInfo, remoteAddr string) error {
	if err := self.isHandWithSelf(remotePeer, remoteAddr); err != nil {
		return err
	}

	return self.checkPeerIdAndIP(remotePeer, remoteAddr)
}

func (self *ConnectCcntmroller) beforeHandshakeCheck(addr string, index int) error {
	err := self.checkReservedPeers(addr)
	if err != nil {
		return err
	}

	if self.hasBoundAddr(addr) {
		return fmt.Errorf("peer %s already in connection records", addr)
	}

	if self.OwnAddress() == addr {
		return fmt.Errorf("connecting with self address %s", addr)
	}

	if self.isBoundFull(index) {
		return fmt.Errorf("[p2p] bound %d connections reach max limit", index)
	}
	if index == INBOUND_INDEX {
		remoteIp, err := common.ParseIPAddr(addr)
		if err != nil {
			return fmt.Errorf("[p2p]parse ip error %v", err.Error())
		}
		connNum := self.getInboundCountWithIp(remoteIp)
		if connNum >= self.MaxConnInBoundPerIP {
			return fmt.Errorf("connections(%d) with ip(%s) has reach max limit(%d), "+
				"conn closed", connNum, remoteIp, self.MaxConnInBoundPerIP)
		}
	}

	return nil
}

func (self *ConnectCcntmroller) isHandWithSelf(remotePeer *peer.PeerInfo, remoteAddr string) error {
	addrIp, err := common.ParseIPAddr(remoteAddr)
	if err != nil {
		self.logger.Warn(err)
		return err
	}
	nodeAddr := addrIp + ":" + strconv.Itoa(int(remotePeer.Port))
	if remotePeer.Id.ToUint64() == self.selfId.Id.ToUint64() {
		self.SetOwnAddress(nodeAddr)
		return ErrHandshakeSelf
	}

	return nil
}

func (self *ConnectCcntmroller) getPeer(kid common.PeerId) *connectedPeer {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	p := self.peers[kid]
	return p
}

func (self *ConnectCcntmroller) savePeer(conn net.Conn, p *peer.PeerInfo, index int) net.Conn {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	addr := conn.RemoteAddr().String()
	self.inoutbounds[index].Add(addr)
	listen := p.RemoteListenAddress()
	if index == INBOUND_INDEX {
		self.inboundListenAddress.Add(listen)
	}

	cid := self.getConnectId()
	self.peers[p.Id] = &connectedPeer{
		connectId: cid,
		addr:      addr,
		peer:      p,
	}

	return &Conn{
		Conn:       conn,
		connectId:  cid,
		kid:        p.Id,
		addr:       addr,
		listenAddr: listen,
		boundIndex: index,
		ccntmroller: self,
	}
}

func (self *ConnectCcntmroller) removePeer(conn *Conn) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.inoutbounds[conn.boundIndex].Remove(conn.addr)
	if conn.boundIndex == INBOUND_INDEX {
		self.inboundListenAddress.Remove(conn.listenAddr)
	}

	p := self.peers[conn.kid]
	if p == nil || p.peer == nil {
		self.logger.Fatalf("connection %s not in ccntmroller", conn.kid.ToHexString())
	} else if p.connectId == conn.connectId { // connection not replaced
		delete(self.peers, conn.kid)
	}
}

// if connection with peer.Kid exist, but has different IP, return error
func (self *ConnectCcntmroller) checkPeerIdAndIP(peer *peer.PeerInfo, addr string) error {
	oldPeer := self.getPeer(peer.Id)
	if oldPeer == nil {
		return nil
	}

	ipOld, err := common.ParseIPAddr(oldPeer.addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]exist peer ip format is wrcntm %s", oldPeer.addr)
		self.logger.Fatal(err)
		return err
	}
	ipNew, err := common.ParseIPAddr(addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]connecting peer ip format is wrcntm %s, close", addr)
		self.logger.Fatal(err)
		return err
	}

	if ipNew != ipOld {
		err := fmt.Errorf("[createPeer]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
		self.logger.Warn(err)
		return err
	}

	return nil
}
