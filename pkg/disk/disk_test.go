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

package disk_test

import (
	"bytes"
	"path/filepath"
	"tealfs/pkg/disk"
	"tealfs/pkg/hash"
	"tealfs/pkg/model"
	"testing"
)

func TestWriteData(t *testing.T) {
	f, path, _, mgrDiskWrites, _, diskMgrWrites, _, _ := newDiskService()
	blockId := model.NewBlockId()
	data := []byte{0, 1, 2, 3, 4, 5}
	h := hash.ForData(data)
	expectedPath := filepath.Join(path.String(), string(blockId))
	mgrDiskWrites <- model.Block{
		Id:   blockId,
		Data: data,
		Hash: h,
	}
	result := <-diskMgrWrites
	if !result.Ok {
		t.Error("Bad write result")
	}
	if !bytes.Equal(f.WrittenData, data) {
		t.Error("Written data is wrong")
	}
	if f.WritePath != expectedPath {
		t.Error("Written path is wrong")
	}
}

func TestReadData(t *testing.T) {
	f, path, _, _, mgrDiskReads, _, diskMgrReads, _ := newDiskService()
	blockId := model.NewBlockId()
	caller := model.NewNodeId()
	data := []byte{0, 1, 2, 3, 4, 5}
	f.ReadData = data
	expectedPath := filepath.Join(path.String(), string(blockId))
	mgrDiskReads <- model.ReadRequest{
		Caller:  caller,
		BlockId: blockId,
	}
	result := <-diskMgrReads
	if !result.Ok {
		t.Error("Bad write result")
	}
	if !bytes.Equal(result.Block.Data, data) {
		t.Error("Written data is wrong")
	}
	if f.ReadPath != expectedPath {
		t.Error("Written path is wrong")
	}
}

func newDiskService() (*disk.MockFileOps, disk.Path, model.NodeId, chan model.Block, chan model.ReadRequest, chan model.WriteResult, chan model.ReadResult, disk.Disk) {
	f := disk.MockFileOps{}
	path := disk.NewPath("/some/fake/path", &f)
	id := model.NewNodeId()
	mgrDiskWrites := make(chan model.Block)
	mgrDiskReads := make(chan model.ReadRequest)
	diskMgrWrites := make(chan model.WriteResult)
	diskMgrReads := make(chan model.ReadResult)
	d := disk.New(path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads)
	return &f, path, id, mgrDiskWrites, mgrDiskReads, diskMgrWrites, diskMgrReads, d
}
