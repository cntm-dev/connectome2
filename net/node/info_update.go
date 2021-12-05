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

package node

import (
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node node) GetBlkHdrs() {
	for _, n := range node.local.List {
		h1 := n.GetHeight()
		h2:= node.local.GetLedger().GetLocalBlockChainHeight()
		if (node.GetState() == ESTABLISH) && (h1 > uint64(h2)) {
			buf, _ := NewMsg("getheaders", node.local)
			go node.Tx(buf)
		}
	}
}

func (node *node) SyncBlk() {
	headerHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	currentBlkHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	if currentBlkHeight >= headerHeight {
		return
	}
	var dValue int32
	var reqCnt uint32
	var i uint32
	noders := node.local.GetNeighborNoder()

	for _, n := range noders {
		if uint32(n.GetHeight()) <= currentBlkHeight {
			ccntminue
		}
		n.RemoveFlightHeightLessThan(currentBlkHeight)
		count := MAXREQBLKONCE - uint32(n.GetFlightHeightCnt())
		dValue = int32(headerHeight - currentBlkHeight - reqCnt)
		flights := n.GetFlightHeights()
		if count == 0 {
			for _, f := range flights {
				hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(f)
				if ledger.DefaultLedger.Store.BlockInCache(hash) == false {
					ReqBlkData(n, hash)
				}
			}

		}
		for i = 1; i <= count && dValue >= 0; i++ {
			hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(currentBlkHeight + reqCnt)

			if ledger.DefaultLedger.Store.BlockInCache(hash) == false {
				ReqBlkData(n, hash)
				n.StoreFlightHeight(currentBlkHeight + reqCnt)
			}
			reqCnt++
			dValue--
		}
	}
}

func (node *node) SendPingToNbr() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			buf, err := NewPingMsg()
			if err != nil {
				log.Error("failed build a new ping message")
			} else {
				go n.Tx(buf)
			}
		}
	}
}

func (node *node) HeartBeatMonitor() {
	noders := node.local.GetNeighborNoder()
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodUpdateTime = config.Parameters.GenBlockTime / TIMESOFUPDATETIME
	} else {
		periodUpdateTime = config.DEFAULTGENBLOCKTIME / TIMESOFUPDATETIME
	}
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			t := n.GetLastRXTime()
			if t.Before(time.Now().Add(-1 * time.Second * time.Duration(periodUpdateTime) * KEEPALIVETIMEOUT)) {
				log.Warn("keepalive timeout!!!")
				n.SetState(INACTIVITY)
				n.CloseConn()
			}
		}
	}
}

func (node *node) ReqNeighborList() {
	buf, _ := NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

func (node *node) ConnectSeeds() {
	if node.IsUptoMinNodeCount() {
		return
	}
	seedNodes := config.Parameters.SeedList
	for _, nodeAddr := range seedNodes {
		found := false
		var n Noder
		var ip net.IP
		node.nbrNodes.Lock()
		for _, tn := range node.nbrNodes.List {
			addr := getNodeAddr(tn)
			ip = addr.IpAddr[:]
			addrstring := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
			if nodeAddr == addrstring {
				n = tn
				found = true
				break
			}
		}
		node.nbrNodes.Unlock()
		if found {
			if n.GetState() == ESTABLISH {
				n.ReqNeighborList()
			}
		} else { //not found
			go node.Connect(nodeAddr)
		}
	}
}

func getNodeAddr(n *node) NodeAddr {
	var addr NodeAddr
	addr.IpAddr, _ = n.GetAddr16()
	addr.Time = n.GetTime()
	addr.Services = n.Services()
	addr.Port = n.GetPort()
	addr.ID = n.GetID()
	return addr
}

func (node *node) reconnect() {
	node.RetryConnAddrs.Lock()
	defer node.RetryConnAddrs.Unlock()
	lst := make(map[string]int)
	for addr := range node.RetryAddrs {
		node.RetryAddrs[addr] = node.RetryAddrs[addr] + 1
		rand.Seed(time.Now().UnixNano())
		log.Trace("Try to reconnect peer, peer addr is ", addr)
		<-time.After(time.Duration(rand.Intn(CONNMAXBACK)) * time.Millisecond)
		log.Trace("Back off time`s up, start connect node")
		node.Connect(addr)
		if node.RetryAddrs[addr] < MAXRETRYCOUNT {
			lst[addr] = node.RetryAddrs[addr]
		}
	}
	node.RetryAddrs = lst

}

func (n *node) TryConnect() {
	if n.fetchRetryNodeFromNeiborList() > 0 {
		n.reconnect()
	}
}

func (n *node) fetchRetryNodeFromNeiborList() int {
	n.nbrNodes.Lock()
	defer n.nbrNodes.Unlock()
	var ip net.IP
	neibornodes := make(map[uint64]*node)
	for _, tn := range n.nbrNodes.List {
		addr := getNodeAddr(tn)
		ip = addr.IpAddr[:]
		nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
		if tn.GetState() == INACTIVITY {
			//add addr to retry list
			n.AddInRetryList(nodeAddr)
			//close legacy node
			if tn.conn != nil {
				tn.CloseConn()
			}
		} else {
			//add others to tmp node map
			n.RemoveFromRetryList(nodeAddr)
			neibornodes[tn.GetID()] = tn
		}
	}
	n.nbrNodes.List = neibornodes
	return len(n.RetryAddrs)
}

// FIXME part of node info update function could be a node method itself intead of
// a node map method
// Fixme the Nodes should be a parameter
func (node *node) updateNodeInfo() {
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MINGENBLOCKTIME {
		periodUpdateTime = config.Parameters.GenBlockTime / TIMESOFUPDATETIME
	} else {
		periodUpdateTime = config.DEFAULTGENBLOCKTIME / TIMESOFUPDATETIME
	}
	ticker := time.NewTicker(time.Second * (time.Duration(periodUpdateTime)))
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			node.SendPingToNbr()
			node.GetBlkHdrs()
			node.SyncBlk()
			node.HeartBeatMonitor()
		case <-quit:
			ticker.Stop()
			return
		}
	}
	// TODO when to close the timer
	//close(quit)
}

func (node *node) updateConnection() {
	t := time.NewTimer(time.Second * CONNMONITOR)
	for {
		select {
		case <-t.C:
			node.ConnectSeeds()
			node.TryConnect()
			t.Stop()
			t.Reset(time.Second * CONNMONITOR)
		}
	}

}