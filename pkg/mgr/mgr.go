// Copyright (C) 2025 Adam Hess
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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Mgr struct {
	UiMgrConnectTos    chan model.UiMgrConnectTo
	ConnsMgrStatuses   chan model.NetConnectionStatus
	ConnsMgrReceives   chan model.ConnsMgrReceive
	DiskMgrReads       chan model.ReadResult
	DiskMgrWrites      chan model.WriteResult
	WebdavMgrGets      chan model.BlockId
	WebdavMgrPuts      chan model.Block
	MgrConnsConnectTos chan model.MgrConnsConnectTo
	MgrConnsSends      chan model.MgrConnsSend
	MgrDiskWrites      chan model.WriteRequest
	MgrDiskReads       chan model.ReadRequest
	MgrUiStatuses      chan model.UiConnectionStatus
	MgrWebdavGets      chan model.BlockResponse
	MgrWebdavPuts      chan model.BlockIdResponse

	nodesAddressMap    map[model.NodeId]string
	nodeConnMap        set.Bimap[model.NodeId, model.ConnId]
	NodeId             model.NodeId
	connAddress        map[model.ConnId]string
	mirrorDistributer  dist.MirrorDistributer
	xorDistributer     dist.XorDistributer
	blockType          model.BlockType
	nodeAddress        string
	savePath           string
	fileOps            disk.FileOps
	pendingBlockWrites pendingBlockWrites
	freeBytes          uint32
}

func NewWithChanSize(chanSize int, nodeAddress string, savePath string, fileOps disk.FileOps, blockType model.BlockType, freeBytes uint32) *Mgr {
	nodeId, err := readNodeId(savePath, fileOps)
	if err != nil {
		panic(err)
	}

	mgr := Mgr{
		UiMgrConnectTos:    make(chan model.UiMgrConnectTo, chanSize),
		ConnsMgrStatuses:   make(chan model.NetConnectionStatus, chanSize),
		ConnsMgrReceives:   make(chan model.ConnsMgrReceive, chanSize),
		DiskMgrWrites:      make(chan model.WriteResult),
		DiskMgrReads:       make(chan model.ReadResult, chanSize),
		WebdavMgrGets:      make(chan model.BlockId, chanSize),
		WebdavMgrPuts:      make(chan model.Block, chanSize),
		MgrConnsConnectTos: make(chan model.MgrConnsConnectTo, chanSize),
		MgrConnsSends:      make(chan model.MgrConnsSend, chanSize),
		MgrDiskWrites:      make(chan model.WriteRequest, chanSize),
		MgrDiskReads:       make(chan model.ReadRequest, chanSize),
		MgrUiStatuses:      make(chan model.UiConnectionStatus, chanSize),
		MgrWebdavGets:      make(chan model.BlockResponse, chanSize),
		MgrWebdavPuts:      make(chan model.BlockIdResponse, chanSize),
		nodesAddressMap:    make(map[model.NodeId]string),
		NodeId:             nodeId,
		connAddress:        make(map[model.ConnId]string),
		nodeConnMap:        set.NewBimap[model.NodeId, model.ConnId](),
		mirrorDistributer:  dist.NewMirrorDistributer(),
		xorDistributer:     dist.NewXorDistributer(),
		blockType:          blockType,
		nodeAddress:        nodeAddress,
		savePath:           savePath,
		fileOps:            fileOps,
		pendingBlockWrites: newPendingBlockWrites(),
		freeBytes:          freeBytes,
	}
	mgr.mirrorDistributer.SetWeight(mgr.NodeId, int(freeBytes))
	mgr.xorDistributer.SetWeight(mgr.NodeId, int(freeBytes))

	return &mgr
}

func readNodeId(savePath string, fileOps disk.FileOps) (model.NodeId, error) {
	data, err := fileOps.ReadFile(filepath.Join(savePath, "node_id"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			nodeId := model.NewNodeId()
			err = fileOps.WriteFile(filepath.Join(savePath, "node_id"), []byte(nodeId))
			if err != nil {
				return "", err
			}
			return nodeId, nil
		}
		return "", err
	}
	return model.NodeId(data), nil
}

func (m *Mgr) Start() error {
	err := m.loadNodeAddressMap()
	if err != nil {
		return err
	}
	go m.eventLoop()
	for nodeId, address := range m.nodesAddressMap {
		if nodeId != m.NodeId {
			m.UiMgrConnectTos <- model.UiMgrConnectTo{
				Address: address,
			}
		}
	}
	return nil
}

func (m *Mgr) loadNodeAddressMap() error {
	data, err := m.fileOps.ReadFile(filepath.Join(m.savePath, "cluster.json"))
	if err != nil {
		return nil // TODO, this should only be for file not found, not other errors
	}
	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, &m.nodesAddressMap)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mgr) saveNodeAddressMap() error {
	data, err := json.Marshal(m.nodesAddressMap)
	if err != nil {
		return err
	}

	err = m.fileOps.WriteFile(filepath.Join(m.savePath, "cluster.json"), data)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mgr) eventLoop() {
	for {
		select {
		case r := <-m.UiMgrConnectTos:
			m.handleConnectToReq(r)
		case r := <-m.ConnsMgrStatuses:
			m.handleNetConnectedStatus(r)
		case r := <-m.ConnsMgrReceives:
			m.handleReceives(r)
		case r := <-m.DiskMgrReads:
			m.handleDiskReadResult(r)
		case r := <-m.DiskMgrWrites:
			m.handleDiskWriteResult(r)
		case r := <-m.WebdavMgrGets:
			m.handleWebdavGets(r)
		case r := <-m.WebdavMgrPuts:
			m.handleWebdavWriteRequest(r)
		}
	}
}

func (m *Mgr) handleConnectToReq(i model.UiMgrConnectTo) {
	m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: string(i.Address)}
}

func (m *Mgr) syncNodesPayloadToSend() model.SyncNodes {
	result := model.NewSyncNodes()
	for node := range m.nodesAddressMap {
		connId, success := m.nodeConnMap.Get1(node)
		if success {
			if address, ok := m.connAddress[connId]; ok {
				result.Nodes.Add(struct {
					Node    model.NodeId
					Address string
				}{Node: node, Address: address})
			}
		}
	}
	return result
}

func (m *Mgr) handleReceives(i model.ConnsMgrReceive) {
	switch p := i.Payload.(type) {
	case *model.IAm:
		m.connAddress[i.ConnId] = p.Address
		m.MgrUiStatuses <- model.UiConnectionStatus{
			Type:          model.Connected,
			RemoteAddress: p.Address,
			Id:            p.NodeId,
		}
		_ = m.addNodeToCluster(*p, i.ConnId)
		syncNodes := m.syncNodesPayloadToSend()
		for n := range m.nodesAddressMap {
			connId, ok := m.nodeConnMap.Get1(n)
			if ok {
				m.MgrConnsSends <- model.MgrConnsSend{
					ConnId:  connId,
					Payload: &syncNodes,
				}
			}
		}
	case *model.SyncNodes:
		remoteNodes := p.GetNodes()
		localNodes := set.NewSetFromMapKeys(m.nodesAddressMap)
		localNodes.Add(m.NodeId)
		missing := remoteNodes.Minus(&localNodes)
		for _, n := range missing.GetValues() {
			address := p.AddressForNode(n)
			m.MgrConnsConnectTos <- model.MgrConnsConnectTo{Address: address}
		}
	case *model.WriteRequest:
		caller, ok := m.nodeConnMap.Get2(i.ConnId)
		if !ok {
			payload := model.WriteResult{
				Ok:      false,
				Message: "connection error",
				Caller:  caller,
				Ptr:     p.Data.Ptr,
			}
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  i.ConnId,
				Payload: &payload,
			}
		}
		m.MgrDiskWrites <- *p
	case *model.WriteResult:
		m.handleDiskWriteResult(*p)
	case *model.ReadRequest:
		m.MgrDiskReads <- *p
	case *model.ReadResult:
		m.handleDiskReadResult(*p)
	default:
		panic("Received unknown payload")
	}
}

func (m *Mgr) handleDiskWriteResult(r model.WriteResult) {
	if r.Caller == m.NodeId {
		resolved, blockId := m.pendingBlockWrites.resolve(r.Ptr)
		var err error = nil
		if !r.Ok {
			err = errors.New(r.Message)
			m.pendingBlockWrites.cancel(blockId)
		}
		switch resolved {
		case done:
			m.MgrWebdavPuts <- model.BlockIdResponse{
				BlockId: blockId,
				Err:     err,
			}
		}
	} else {
		c, ok := m.nodeConnMap.Get1(r.Caller)
		if ok {
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  c,
				Payload: &r,
			}
		} else {
			panic("not connected")
		}
	}
}

func (m *Mgr) handleDiskReadResult(r model.ReadResult) {
	if r.Ok {
		if r.Caller == m.NodeId {
			m.MgrWebdavGets <- model.BlockResponse{
				Block: model.Block{
					Id:   r.BlockId,
					Type: model.Mirrored,
					Data: r.Data.Data,
				},
				Err: nil,
			}
		} else {
			c, ok := m.nodeConnMap.Get1(r.Caller)
			if ok {
				m.MgrConnsSends <- model.MgrConnsSend{
					ConnId:  c,
					Payload: &r,
				}
			} else {
				fmt.Println("handleDiskReadResult: not connected")
			}
		}
	} else {
		m.readDiskPtr(r.Ptrs, r.BlockId)
	}
}

func (m *Mgr) addNodeToCluster(iam model.IAm, c model.ConnId) error {
	m.nodesAddressMap[iam.NodeId] = iam.Address
	err := m.saveNodeAddressMap()
	if err != nil {
		return err
	}
	m.nodeConnMap.Add(iam.NodeId, c)
	m.mirrorDistributer.SetWeight(iam.NodeId, int(iam.FreeBytes))
	m.xorDistributer.SetWeight(iam.NodeId, int(iam.FreeBytes))
	return nil
}

func (m *Mgr) handleNetConnectedStatus(cs model.NetConnectionStatus) {
	switch cs.Type {
	case model.Connected:
		m.MgrConnsSends <- model.MgrConnsSend{
			ConnId: cs.Id,
			Payload: &model.IAm{
				NodeId:    m.NodeId,
				Address:   m.nodeAddress,
				FreeBytes: m.freeBytes,
			},
		}
	case model.NotConnected:
		address := m.connAddress[cs.Id]
		delete(m.connAddress, cs.Id)
		// Todo: need a mechanism to back off
		m.MgrConnsConnectTos <- model.MgrConnsConnectTo{
			Address: address,
		}
		// Todo: reflect this in the ui
		fmt.Println("Not Connected")
	}
}

func (m *Mgr) handleWebdavGets(blockId model.BlockId) {
	ptrs := m.mirrorDistributer.PointersForId(blockId)
	if len(ptrs) == 0 {
		m.MgrWebdavGets <- model.BlockResponse{
			Block: model.Block{},
			Err:   errors.New("not found"),
		}
	} else {
		m.readDiskPtr(ptrs, blockId)
	}
}

func (m *Mgr) readDiskPtr(ptrs []model.DiskPointer, blockId model.BlockId) {
	if len(ptrs) == 0 {
		return
	}
	n := ptrs[0].NodeId
	rr := model.ReadRequest{
		Caller:  m.NodeId,
		Ptrs:    ptrs,
		BlockId: blockId,
	}
	if m.NodeId == n {
		m.MgrDiskReads <- rr
	} else {
		c, ok := m.nodeConnMap.Get1(n)
		if ok {
			m.MgrConnsSends <- model.MgrConnsSend{
				ConnId:  c,
				Payload: &rr,
			}
		} else {
			m.MgrWebdavGets <- model.BlockResponse{
				Block: model.Block{},
				Err:   errors.New("no connection"),
			}
		}
	}
}

func (m *Mgr) handleWebdavWriteRequest(w model.Block) {
	switch w.Type {
	case model.Mirrored:
		m.handleMirroredWriteRequest(w)
	case model.XORed:
		m.handleXoredWriteRequest(w)
	default:
		panic("unknown block type")
	}
}

func (m *Mgr) handleMirroredWriteRequest(b model.Block) {
	ptrs := m.mirrorDistributer.PointersForId(b.Id)
	for _, ptr := range ptrs {
		m.pendingBlockWrites.add(b.Id, ptr)
		data := model.RawData{
			Data: b.Data,
			Ptr:  ptr,
		}
		writeRequest := model.WriteRequest{
			Data:   data,
			Caller: m.NodeId,
		}
		if ptr.NodeId == m.NodeId {
			m.MgrDiskWrites <- writeRequest
		} else {
			c, ok := m.nodeConnMap.Get1(ptr.NodeId)
			if ok {
				m.MgrConnsSends <- model.MgrConnsSend{
					ConnId:  c,
					Payload: &writeRequest,
				}
			} else {
				m.pendingBlockWrites.cancel(b.Id)
				m.MgrWebdavPuts <- model.BlockIdResponse{
					BlockId: b.Id,
					Err:     errors.New("not connected"),
				}
				return
			}
		}
	}
}

func (m *Mgr) handleXoredWriteRequest(b model.Block) {
	panic("not implemented yet")
}
