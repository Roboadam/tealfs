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

package model_test

import (
	"tealfs/pkg/model"
	"testing"
)

func TestSyncNodes(t *testing.T) {
	n1 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:1",
	}
	n2 := struct {
		Node    model.NodeId
		Address string
	}{
		Node:    model.NewNodeId(),
		Address: "node:2",
	}
	sn1 := model.NewSyncNodes()
	sn1.Nodes.Add(n1)
	sn1.Nodes.Add(n2)
	sn2 := model.NewSyncNodes()
	sn2.Nodes.Add(n2)
	sn2.Nodes.Add(n1)

	if !sn1.Equal(&sn2) {
		t.Error("should be equal")
	}

	bytes1 := sn1.ToBytes()
	sn3 := model.ToSyncNodes(bytes1[1:])

	if !sn1.Equal(sn3) {
		t.Error("should be equal")
	}
}

func TestReadResult(t *testing.T) {
	rr1 := model.ReadResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller",
		Ptrs: []model.DiskPointer{
			{
				NodeId:   "node1",
				FileName: "fileName1",
			},
			{
				NodeId:   "node2",
				FileName: "fileName2",
			},
		},
		Data: model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   "node1",
				FileName: "fileName1",
			},
			Data: []byte{1, 2, 3},
		},
		BlockId: "blockId",
	}
	rr2 := model.ReadResult{
		Ok:      false,
		Message: "some message",
		Caller:  "some caller",
		Ptrs: []model.DiskPointer{
			{
				NodeId:   "node1",
				FileName: "fileName1",
			},
			{
				NodeId:   "node2",
				FileName: "fileName2",
			},
		},
		Data: model.RawData{
			Ptr: model.DiskPointer{
				NodeId:   "node1",
				FileName: "fileName1",
			},
			Data: []byte{1, 2, 3},
		},
		BlockId: "blockId",
	}
	if rr1.Equal(&rr2) {
		t.Error("should not be equal")
	}

	rr2.Ok = true

	if !rr1.Equal(&rr2) {
		t.Error("should be equal")
	}

	bytes1 := rr1.ToBytes()
	rr3 := model.ToReadResult(bytes1[1:])

	if !rr1.Equal(rr3) {
		t.Error("should be equal")
	}
}

func TestWriteResult(t *testing.T) {
	wr1 := model.WriteResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller",
		Ptr: model.DiskPointer{
			NodeId:   "nodeId",
			FileName: "fileName",
		},
	}
	wr2 := model.WriteResult{
		Ok:      true,
		Message: "some message",
		Caller:  "some caller 2",
		Ptr: model.DiskPointer{
			NodeId:   "nodeId",
			FileName: "fileName",
		},
	}

	if wr1.Equal(&wr2) {
		t.Error("should not be equal")
		return
	}

	wr2.Caller = "some caller"

	if !wr1.Equal(&wr2) {
		t.Error("should be equal")
		return
	}

	bytes1 := wr1.ToBytes()
	rr3 := model.ToWriteResult(bytes1[1:])

	if !wr1.Equal(rr3) {
		t.Error("should be equal")
		return
	}
}

func TestReadRequest(t *testing.T) {
	rr1 := model.ReadRequest{
		Caller: "caller1",
		Ptrs: []model.DiskPointer{
			{
				NodeId:   "nodeId1",
				FileName: "filename1",
			},
		},
		BlockId: "blockId1",
	}
	rr2 := model.ReadRequest{
		Caller: "caller1",
		Ptrs: []model.DiskPointer{
			{
				NodeId:   "nodeId1",
				FileName: "filename1",
			},
		},
		BlockId: "blockId2",
	}

	if rr1.Equal(&rr2) {
		t.Error("should not be equal")
		return
	}

	rr2.BlockId = "blockId1"

	if !rr1.Equal(&rr2) {
		t.Error("should be equal")
		return
	}

	bytes1 := rr1.ToBytes()
	rr3 := model.ToReadRequest(bytes1[1:])

	if !rr1.Equal(rr3) {
		t.Error("should be equal")
		return
	}
}
