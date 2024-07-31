// Copyright (C) 2024 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package mgr

import (
	"tealfs/pkg/hash"
	"tealfs/pkg/nodes"
	"tealfs/pkg/proto"
	"tealfs/pkg/store"
	"testing"
)

func TestConnectToMgr(t *testing.T) {
	const expectedAddress = "some-address:123"

	m := NewWithChanSize(0)
	m.Start()

	m.UiMgrConnectTos <- UiMgrConnectTo{
		Address: expectedAddress,
	}

	expectedMessage := <-m.MgrConnsConnectTos

	if expectedMessage.Address != expectedAddress {
		t.Error("Received address", expectedMessage.Address)
	}
}

func TestConnectToSuccess(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)
}

func TestReceiveSyncNodes(t *testing.T) {
	const sharedAddress = "some-address:123"
	const sharedConnectionId = 1
	var sharedNodeId = nodes.NewNodeId()
	const localAddress = "some-address2:234"
	const localConnectionId = 2
	var localNodeId = nodes.NewNodeId()
	const remoteAddress = "some-address3:345"
	var remoteNodeId = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: sharedAddress, conn: sharedConnectionId, node: sharedNodeId},
		{address: localAddress, conn: localConnectionId, node: localNodeId},
	}, t)

	sn := proto.NewSyncNodes()
	sn.Nodes.Add(struct {
		Node    nodes.Id
		Address string
	}{Node: sharedNodeId, Address: sharedAddress})
	sn.Nodes.Add(struct {
		Node    nodes.Id
		Address string
	}{Node: remoteNodeId, Address: remoteAddress})
	m.ConnsMgrReceives <- ConnsMgrReceive{
		ConnId:  sharedConnectionId,
		Payload: &sn,
	}

	expectedConnectTo := <-m.MgrConnsConnectTos
	if expectedConnectTo.Address != remoteAddress {
		t.Error("expected to connect to", remoteAddress)
	}
}

func TestReceiveSaveData(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	ids := []store.Id{}
	for range 100 {
		ids = append(ids, store.NewId())
	}

	value := []byte("123")

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, id := range ids {
		m.ConnsMgrReceives <- ConnsMgrReceive{
			ConnId: expectedConnectionId1,
			Payload: &proto.SaveData{
				Block: store.Block{
					Id:   store.Id(id),
					Data: value,
					Hash: hash.ForData(value),
				},
			},
		}

		select {
		case w := <-m.MgrDiskWrites:
			meCount++
			if w.Id != id {
				t.Error("expected to write to 1, got", w.Id)
			}
		case s := <-m.MgrConnsSends:
			//Todo: s.Payload should be checked for the correct value
			if s.ConnId == expectedConnectionId1 {
				oneCount++
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
			} else {
				t.Error("expected to connect to", s.ConnId)
			}
		}
	}
	if meCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to get some data")
	}
}

func TestReceiveDiskRead(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	storeId1 := store.NewId()
	data1 := []byte{0x00, 0x01, 0x02}
	hash1 := hash.ForData(data1)

	rr := proto.ReadResult{
		Ok:      true,
		Message: "",
		Caller:  m.nodeId,
		Block: store.Block{
			Id:   storeId1,
			Data: data1,
			Hash: hash1,
		},
	}

	m.DiskMgrReads <- rr

	toWebdav := <-m.MgrWebdavGets

	if !rr.Equal(&toWebdav) {
		t.Errorf("rr didn't equal toWebdav")
	}

	rr2 := proto.ReadResult{
		Ok:      true,
		Message: "",
		Caller:  expectedNodeId1,
		Block: store.Block{
			Id:   storeId1,
			Data: data1,
			Hash: hash1,
		},
	}

	m.DiskMgrReads <- rr2
	sent2 := <-m.MgrConnsSends

	expectedMCS2 := MgrConnsSend{
		ConnId:  expectedConnectionId1,
		Payload: &rr2,
	}

	if !sent2.Equal(&expectedMCS2) {
		t.Errorf("sent2 not equal expectedMCS2")
	}
}

func TestWebdavGet(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	ids := []store.Id{}
	for range 100 {
		ids = append(ids, store.NewId())
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, id := range ids {
		m.WebdavMgrGets <- proto.ReadRequest{
			Caller:  m.nodeId,
			BlockId: id,
		}

		select {
		case r := <-m.MgrDiskReads:
			meCount++
			if r.BlockId != id {
				t.Error("expected to read to 1, got", r.BlockId)
			}
		case s := <-m.MgrConnsSends:
			if s.ConnId == expectedConnectionId1 {
				oneCount++
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
			} else {
				t.Error("expected to connect to", s.ConnId)
			}
		}
	}
	if meCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to get some data")
	}
}

func TestWebdavPut(t *testing.T) {
	const expectedAddress1 = "some-address:123"
	const expectedConnectionId1 = 1
	var expectedNodeId1 = nodes.NewNodeId()
	const expectedAddress2 = "some-address2:234"
	const expectedConnectionId2 = 2
	var expectedNodeId2 = nodes.NewNodeId()

	m := mgrWithConnectedNodes([]connectedNode{
		{address: expectedAddress1, conn: expectedConnectionId1, node: expectedNodeId1},
		{address: expectedAddress2, conn: expectedConnectionId2, node: expectedNodeId2},
	}, t)

	blocks := []store.Block{}
	for i := range 100 {
		data := []byte{byte(i)}
		hash := hash.ForData(data)
		block := store.Block{
			Id:   store.NewId(),
			Data: data,
			Hash: hash,
		}
		blocks = append(blocks, block)
	}

	meCount := 0
	oneCount := 0
	twoCount := 0

	for _, block := range blocks {
		m.WebdavMgrPuts <- block

		select {
		case w := <-m.MgrDiskWrites:
			meCount++
			if !w.Equal(&block) {
				t.Error("expected the origial block")
			}
		case s := <-m.MgrConnsSends:
			if s.ConnId == expectedConnectionId1 {
				oneCount++
			} else if s.ConnId == expectedConnectionId2 {
				twoCount++
			} else {
				t.Error("expected to connect to", s.ConnId)
			}
		}
	}
	if meCount == 0 || oneCount == 0 || twoCount == 0 {
		t.Error("Expected everyone to fetch some data")
	}
}

type connectedNode struct {
	address string
	conn    ConnId
	node    nodes.Id
}

func mgrWithConnectedNodes(nodes []connectedNode, t *testing.T) Mgr {
	m := NewWithChanSize(0)
	m.Start()
	var nodesInCluster []connectedNode

	for _, n := range nodes {
		// Send a message to Mgr indicating another
		// node has connected
		m.ConnsMgrStatuses <- ConnsMgrStatus{
			Type:          Connected,
			RemoteAddress: n.address,
			Id:            n.conn,
		}

		// Then Mgr should send an Iam payload to
		// the appropriate connection id with its
		// own node id
		expectedIam := <-m.MgrConnsSends
		payload := expectedIam.Payload
		switch p := payload.(type) {
		case *proto.IAm:
			if p.NodeId != m.nodeId {
				t.Error("Unexpected nodeId")
			}
			if expectedIam.ConnId != n.conn {
				t.Error("Unexpected connId")
			}
		default:
			t.Error("Unexpected payload", p)
		}

		// Send a message to Mgr indicating the newly
		// connected node has sent us an Iam payload
		iamPayload := proto.IAm{
			NodeId: n.node,
		}
		m.ConnsMgrReceives <- ConnsMgrReceive{
			ConnId:  n.conn,
			Payload: &iamPayload,
		}

		nodesInCluster = append(nodesInCluster, n)
		var payloadsFromMgr []MgrConnsSend

		for range nodesInCluster {
			payloadsFromMgr = append(payloadsFromMgr, <-m.MgrConnsSends)
		}

		expectedSyncNodes := expectedSyncNodesForCluster(nodesInCluster)
		syncNodesWeSent := assertAllPayloadsSyncNodes(t, payloadsFromMgr)

		if !cIdSnSliceEquals(expectedSyncNodes, syncNodesWeSent) {
			t.Error("Expected sync nodes to match", expectedSyncNodes, syncNodesWeSent)
		}
	}

	return m
}

func assertAllPayloadsSyncNodes(t *testing.T, mcs []MgrConnsSend) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes
	for _, mc := range mcs {
		switch p := mc.Payload.(type) {
		case *proto.SyncNodes:
			results = append(results, struct {
				ConnId  ConnId
				Payload proto.SyncNodes
			}{ConnId: mc.ConnId, Payload: *p})
		default:
			t.Error("Unexpected payload", p)
		}
	}
	return results
}

type connIdAndSyncNodes struct {
	ConnId  ConnId
	Payload proto.SyncNodes
}

func cIdSnSliceEquals(a, b []connIdAndSyncNodes) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		oneEqual := false
		for j := range b {
			if cIdSnEquals(a[i], b[j]) {
				oneEqual = true
			}
		}
		if !oneEqual {
			return false
		}
	}
	return true
}

func cIdSnEquals(a, b connIdAndSyncNodes) bool {
	if a.ConnId != b.ConnId {
		return false
	}
	return a.Payload.Equal(&b.Payload)
}

func expectedSyncNodesForCluster(cluster []connectedNode) []connIdAndSyncNodes {
	var results []connIdAndSyncNodes

	sn := proto.NewSyncNodes()
	for _, node := range cluster {
		sn.Nodes.Add(struct {
			Node    nodes.Id
			Address string
		}{Node: node.node, Address: node.address})
	}

	for _, node := range cluster {
		results = append(results, struct {
			ConnId  ConnId
			Payload proto.SyncNodes
		}{ConnId: node.conn, Payload: sn})
	}
	return results
}
