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

package conns

import (
	"context"
	"net"
	"tealfs/pkg/model"
	"tealfs/pkg/tnet"
)

type Conns struct {
	netConns      map[model.ConnId]net.Conn
	nextId        model.ConnId
	acceptedConns chan AcceptedConns
	outStatuses   chan<- model.NetConnectionStatus
	outReceives   chan<- model.ConnsMgrReceive
	inConnectTo   <-chan model.MgrConnsConnectTo
	inSends       <-chan model.MgrConnsSend
	Address       string
	provider      ConnectionProvider
	nodeId        model.NodeId
	listener      net.Listener
}

func NewConns(
	outStatuses chan<- model.NetConnectionStatus,
	outReceives chan<- model.ConnsMgrReceive,
	inConnectTo <-chan model.MgrConnsConnectTo,
	inSends <-chan model.MgrConnsSend,
	provider ConnectionProvider,
	address string,
	nodeId model.NodeId,
	ctx context.Context) Conns {

	listener, err := provider.GetListener(address)
	if err != nil {
		panic(err)
	}
	c := Conns{
		netConns:      make(map[model.ConnId]net.Conn, 3),
		nextId:        model.ConnId(0),
		acceptedConns: make(chan AcceptedConns),
		outStatuses:   outStatuses,
		outReceives:   outReceives,
		inConnectTo:   inConnectTo,
		inSends:       inSends,
		provider:      provider,
		nodeId:        nodeId,
		listener:      listener,
	}

	go c.consumeChannels(ctx)
	go c.listen()

	return c
}

func (c *Conns) consumeChannels(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.listener.Close()
			return
		case acceptedConn := <-c.acceptedConns:
			id := c.saveNetConn(acceptedConn.netConn)
			c.outStatuses <- model.NetConnectionStatus{
				Type: model.Connected,
				Msg:  "Success",
				Id:   id,
			}
			go c.consumeData(id)
		case connectTo := <-c.inConnectTo:
			// Todo: this needs to be non blocking
			id, err := c.connectTo(connectTo.Address)
			if err == nil {
				c.outStatuses <- model.NetConnectionStatus{
					Type: model.Connected,
					Msg:  "Success",
					Id:   id,
				}
				go c.consumeData(id)
			} else {
				c.outStatuses <- model.NetConnectionStatus{
					Type: model.NotConnected,
					Msg:  "Failure connecting",
					Id:   id,
				}
			}
		case sendReq := <-c.inSends:
			_, ok := c.netConns[sendReq.ConnId]
			if !ok {
				c.handleSendFailure(sendReq)
			} else {
				//Todo maybe this should be async
				err := tnet.SendPayload(c.netConns[sendReq.ConnId], sendReq.Payload.ToBytes())
				if err != nil {
					c.handleSendFailure(sendReq)
				}
			}
		}
	}
}

func (c *Conns) handleSendFailure(sendReq model.MgrConnsSend) {
	payload := sendReq.Payload
	switch p := payload.(type) {
	case *model.ReadRequest:
		if len(p.Ptrs) > 0 {
			ptrs := p.Ptrs[1:]
			rr := model.ReadRequest{
				Caller:  p.Caller,
				Ptrs:    ptrs,
				BlockId: p.BlockId,
			}

			c.outReceives <- model.ConnsMgrReceive{
				ConnId:  sendReq.ConnId,
				Payload: &rr,
			}
		} else {
			c.outReceives <- model.ConnsMgrReceive{
				ConnId: sendReq.ConnId,
				Payload: &model.ReadResult{
					Ok:      false,
					Message: "no pointers in read request",
					Caller:  p.Caller,
					BlockId: p.BlockId,
				},
			}
		}
	}
}

func (c *Conns) listen() {
	for {
		conn, err := c.listener.Accept()
		if err == nil {
			incomingConnReq := AcceptedConns{netConn: conn}
			c.acceptedConns <- incomingConnReq
		} else {
			return
		}
	}
}

type AcceptedConns struct {
	netConn net.Conn
}

func (c *Conns) consumeData(conn model.ConnId) {
	for {
		netConn := c.netConns[conn]
		bytes, err := tnet.ReadPayload(netConn)
		if err != nil {
			_ = netConn.Close()
			delete(c.netConns, conn)
			c.outStatuses <- model.NetConnectionStatus{
				Type: model.NotConnected,
				Msg:  "Connection closed",
				Id:   conn,
			}
			return
		}
		payload := model.ToPayload(bytes)
		c.outReceives <- model.ConnsMgrReceive{
			ConnId:  conn,
			Payload: payload,
		}
	}
}

func (c *Conns) connectTo(address string) (model.ConnId, error) {
	netConn, err := c.provider.GetConnection(address)
	if err != nil {
		return 0, err
	}
	id := c.saveNetConn(netConn)
	return id, nil
}

func (c *Conns) saveNetConn(netConn net.Conn) model.ConnId {
	id := c.nextId
	c.nextId++
	c.netConns[id] = netConn
	return id
}
