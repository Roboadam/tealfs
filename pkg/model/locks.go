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
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
)

type LockReleaseId string
type LockToken string

func NewLockReleaseId() LockReleaseId {
	idValue := uuid.New()
	return LockReleaseId(idValue.String())
}

func NewLockToken() LockToken {
	idValue := uuid.New()
	return LockToken(idValue.String())
}

type LockConfirmRequest struct {
	Now          time.Time
	Name0, Name1 string
	Conditions   []webdav.Condition
}

func (l *LockConfirmRequest) ToBytes() []byte {
	now := Int64ToBytes(l.Now.UnixNano())
	name0 := StringToBytes(l.Name0)
	name1 := StringToBytes(l.Name1)
	numConditions := IntToBytes(uint32(len(l.Conditions)))
	conditionBytes := []byte{}
	for _, condition := range l.Conditions {
		conditionBytes = append(conditionBytes, ConditionToBytes(&condition)...)
	}
	return bytes.Join([][]byte{now, name0, name1, numConditions, conditionBytes}, []byte{})
}

func ConditionToBytes(condition *webdav.Condition) []byte {
	not := BoolToBytes(condition.Not)
	token := StringToBytes(condition.Token)
	etag := StringToBytes(condition.ETag)
	return bytes.Join([][]byte{not, token, etag}, []byte{})
}

func ConditionEquals(condition1 *webdav.Condition, condition2 *webdav.Condition) bool {
	if condition1.Not != condition2.Not {
		return false
	}
	if condition1.ETag != condition2.ETag {
		return false
	}
	if condition1.Token != condition2.Token {
		return false
	}
	return true
}

func (l *LockConfirmRequest) Equal(p Payload) bool {
	if o, ok := p.(*LockConfirmRequest); ok {
		if l.Now != o.Now {
			return false
		}
		if l.Name0 != o.Name0 {
			return false
		}
		if l.Name1 != o.Name1 {
			return false
		}
		for i, c := range l.Conditions {
			if !ConditionEquals(&c, &l.Conditions[i]) {
				return false
			}
		}
	}
	return true
}

func ToLockConfirmRequest(data []byte) *LockConfirmRequest {
	panic("not implemented") // TODO: Implement
}

type LockConfirmResponse struct {
	Err       error
	ReleaseId LockReleaseId
}

func (l *LockConfirmResponse) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockConfirmResponse) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockConfirmResponse(data []byte) *LockConfirmResponse {
	panic("not implemented") // TODO: Implement
}

type LockCreateRequest struct {
	Now     time.Time
	Details webdav.LockDetails
}

func (l *LockCreateRequest) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockCreateRequest) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockCreateRequest(data []byte) *LockCreateRequest {
	panic("not implemented") // TODO: Implement
}

type LockCreateResponse struct {
	Token LockToken
	Err   error
}

func (l *LockCreateResponse) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockCreateResponse) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockCreateResponse(data []byte) *LockCreateResponse {
	panic("not implemented") // TODO: Implement
}

type LockRefreshRequest struct {
	Now      time.Time
	Token    LockToken
	Duration time.Duration
}

func (l *LockRefreshRequest) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockRefreshRequest) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockRefreshRequest(data []byte) *LockRefreshRequest {
	panic("not implemented") // TODO: Implement
}

type LockRefreshResponse struct {
	Details webdav.LockDetails
	Err     error
}

func (l *LockRefreshResponse) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockRefreshResponse) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockRefreshResponse(data []byte) *LockRefreshResponse {
	panic("not implemented") // TODO: Implement
}

type LockUnlockRequest struct {
	Now   time.Time
	Token LockToken
}

func (l *LockUnlockRequest) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockUnlockRequest) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockUnlockRequest(data []byte) *LockUnlockRequest {
	panic("not implemented") // TODO: Implement
}

type LockUnlockResponse struct {
	Err error
}

func (l *LockUnlockResponse) ToBytes() []byte {
	panic("not implemented") // TODO: Implement
}

func (l *LockUnlockResponse) Equal(_ Payload) bool {
	panic("not implemented") // TODO: Implement
}

func ToLockUnlockResponse(data []byte) *LockUnlockResponse {
	panic("not implemented") // TODO: Implement
}