package node

import (
	"net"
	"time"
)

type RemoteNode struct {
	NodeId  Id
	Address string
	Conn    net.Conn
}

func (node *RemoteNode) Connect() {
	if node.Conn == nil {
		node.connectUntilSuccess()
	}
}

func (node *RemoteNode) Disconnect() {
	if node.Conn != nil {
		node.Conn.Close()
	}
}

func (node *RemoteNode) connectUntilSuccess() {
	var err error
	node.Conn, err = net.Dial("tcp", node.Address)
	for err == nil {
		time.Sleep(time.Second * 2)
		node.Conn, err = net.Dial("tcp", node.Address)
	}
}