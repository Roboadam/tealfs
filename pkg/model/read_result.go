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

package model

import (
	"bytes"
)

type ReadResult struct {
	Ok      bool
	Message string
	Caller  NodeId
	Block   Block
}

func (r *ReadResult) Equal(p Payload) bool {

	if o, ok := p.(*ReadResult); ok {
		if r.Ok != o.Ok {
			return false
		}
		if r.Message != o.Message {
			return false
		}
		if r.Caller != o.Caller {
			return false
		}
		if !r.Block.Equal(&o.Block) {
			return false
		}

		return true
	}
	return false
}

func (r *ReadResult) ToBytes() []byte {
	ok := BoolToBytes(r.Ok)
	message := StringToBytes(r.Message)
	caller := StringToBytes(string(r.Caller))
	block := BlockToBytes(r.Block)
	payload := bytes.Join([][]byte{ok, message, caller, block}, []byte{})
	return AddType(ReadResultType, payload)
}

func ToReadResult(data []byte) *ReadResult {
	ok, remainder := BoolFromBytes(data)
	message, remainder := StringFromBytes(remainder)
	caller, remainder := StringFromBytes(remainder)
	block, _ := BlockFromBytes(remainder)
	return &ReadResult{
		Ok:      ok,
		Message: message,
		Caller:  NodeId(caller),
		Block:   block,
	}
}
