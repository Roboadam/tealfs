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
	"time"

	"golang.org/x/net/webdav"
)

func TestConfirmRequest(t *testing.T) {
	cr := model.LockConfirmRequest{
		Now:   time.Now(),
		Name0: "name0val",
		Name1: "name1val",
		Conditions: []webdav.Condition{
			{
				Not:   false,
				Token: "token1",
				ETag:  "etag1",
			},
			{
				Not:   true,
				Token: "token2",
				ETag:  "etag2",
			},
		},
		Id:     "lockMessageId1",
		Caller: "caller1",
	}

	serialized := cr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockConfirmRequest:
		if !cr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}

}

func TestConfirmResponse(t *testing.T) {
	cr := model.LockConfirmResponse{
		Ok:        false,
		Message:   "message1",
		ReleaseId: "releaseId1",
		Id:        "abc123",
		Caller:    "caller1",
	}
	serialized := cr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockConfirmResponse:
		if !cr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestCreateRequest(t *testing.T) {
	cr := model.LockCreateRequest{
		Now: time.Now(),
		Details: webdav.LockDetails{
			Root:      "root1",
			Duration:  time.Duration(1234),
			OwnerXML:  "<a href=\"https://example.com\">example</a>",
			ZeroDepth: false,
		},
		Id:     "lockMessageId1",
		Caller: "caller1",
	}
	serialized := cr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockCreateRequest:
		if !cr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestCreateResponse(t *testing.T) {
	cr := model.LockCreateResponse{
		Token:   "token1",
		Ok:      true,
		Message: "message1",
		Caller:  "caller1",
		Id:      "abc123",
	}
	serialized := cr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockCreateResponse:
		if !cr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestRefreshRequest(t *testing.T) {
	rr := model.LockRefreshRequest{
		Now:      time.Now(),
		Token:    "token1",
		Duration: time.Duration(1234),
		Id:       "lockMessageId1",
		Caller:   "caller1",
	}
	serialized := rr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockRefreshRequest:
		if !rr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestRefreshResponse(t *testing.T) {
	rr := model.LockRefreshResponse{
		Details: webdav.LockDetails{
			Root:      "root1",
			Duration:  time.Duration(1234),
			OwnerXML:  "<a href=\"https://example.com\">example</a>",
			ZeroDepth: true,
		},
		Ok:      true,
		Message: "message1",
		Caller:  "caller1",
		Id:      "abc123",
	}
	serialized := rr.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockRefreshResponse:
		if !rr.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestUnlockRequest(t *testing.T) {
	ur := model.LockUnlockRequest{
		Now:    time.Now(),
		Token:  "token1",
		Id:     "lockMessageId1",
		Caller: "caller1",
	}
	serialized := ur.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockUnlockRequest:
		if !ur.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestUnlockResponse(t *testing.T) {
	ur := model.LockUnlockResponse{
		Ok:      true,
		Message: "message1",
		Caller:  "caller1",
		Id:      "abc123",
	}
	serialized := ur.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockUnlockResponse:
		if !ur.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}

func TestLockReleaseId(t *testing.T) {
	id := model.LockMessageId("abc123")
	serialized := id.ToBytes()
	switch p := model.ToPayload(serialized).(type) {
	case *model.LockMessageId:
		if !id.Equal(p) {
			t.Error("Expected values to be equal")
			return
		}
	default:
		t.Error("Unexpected payload", p)
		return
	}
}
