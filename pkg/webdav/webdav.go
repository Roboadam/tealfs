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

package webdav

import (
	"context"
	"net/http"
	"tealfs/pkg/model"

	"golang.org/x/net/webdav"
)

type Webdav struct {
	webdavMgrGets      chan model.BlockId
	webdavMgrPuts      chan model.Block
	mgrWebdavGets      chan model.BlockResponse
	mgrWebdavPuts      chan model.BlockIdResponse
	mgrWebdavIsPrimary chan bool
	fileSystem         FileSystem
	nodeId             model.NodeId
	pendingReads       map[model.BlockId]chan model.BlockResponse
	pendingPuts        map[model.BlockId]chan model.BlockIdResponse
	lockSystem         webdav.LockSystem
	bindAddress        string
	server             *http.Server
}

func New(
	nodeId model.NodeId,
	webdavMgrGets chan model.BlockId,
	webdavMgrPuts chan model.Block,
	mgrWebdavGets chan model.BlockResponse,
	mgrWebdavPuts chan model.BlockIdResponse,
	bindAddress string,
	ctx context.Context,
) Webdav {
	w := Webdav{
		webdavMgrGets: webdavMgrGets,
		webdavMgrPuts: webdavMgrPuts,
		mgrWebdavGets: mgrWebdavGets,
		mgrWebdavPuts: mgrWebdavPuts,
		fileSystem:    NewFileSystem(nodeId),
		nodeId:        nodeId,
		pendingReads:  make(map[model.BlockId]chan model.BlockResponse),
		pendingPuts:   make(map[model.BlockId]chan model.BlockIdResponse),
		lockSystem:    webdav.NewMemLS(),
		bindAddress:   bindAddress,
	}
	w.start(ctx)
	return w
}

func (w *Webdav) start(ctx context.Context) {
	go w.eventLoop(ctx)

	handler := &webdav.Handler{
		Prefix:     "/",
		FileSystem: &w.fileSystem,
		LockSystem: w.lockSystem,
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	w.server = &http.Server{
		Addr:    w.bindAddress,
		Handler: mux,
	}
	go w.server.ListenAndServe()
}

func (w *Webdav) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.server.Shutdown(context.Background())
		case r := <-w.mgrWebdavGets:
			ch, ok := w.pendingReads[r.Block.Id]
			if ok {
				ch <- r
				delete(w.pendingReads, r.Block.Id)
			}
		case r := <-w.mgrWebdavPuts:
			ch, ok := w.pendingPuts[r.BlockId]
			if ok {
				ch <- r
				delete(w.pendingPuts, r.BlockId)
			}
		case r := <-w.fileSystem.ReadReqResp:
			w.webdavMgrGets <- r.Req
			w.pendingReads[r.Req] = r.Resp
		case r := <-w.fileSystem.WriteReqResp:
			w.webdavMgrPuts <- r.Req
			w.pendingPuts[r.Req.Id] = r.Resp
		}
	}
}
