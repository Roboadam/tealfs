package node

import (
	"errors"
	"net"
)

type RemoteNode struct {
	NodeId  NodeId
	Address string
	Conn    net.Conn
}

func (node *RemoteNode) Connect() error {
	if node.Conn == nil {
		return node.connectUnconnectedNode()
	}

	return nil
}

func (node *RemoteNode) connectUnconnectedNode() error {
	tcpConn, error := net.Dial("tcp", node.Address)
	if error == nil {
		node.Conn = tcpConn
	} else {
		return errors.New("can't connect")
	}
	return nil
}
