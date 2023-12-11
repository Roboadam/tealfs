package mgr

import (
	"fmt"
	"tealfs/pkg/model/cmds"
	"tealfs/pkg/node"
	"tealfs/pkg/proto"
	"tealfs/pkg/tnet"
)

type Mgr struct {
	node     node.Node
	userCmds chan cmds.User
	tNet     tnet.TNet
	conns    *tnet.Conns
}

func New(userCmds chan cmds.User, tNet tnet.TNet) Mgr {
	myNodeId := node.NewNodeId()
	base := node.Node{Id: myNodeId, Address: node.NewAddress(tNet.GetBinding())}
	return Mgr{
		node:     base,
		userCmds: userCmds,
		conns:    tnet.NewConns(tNet, myNodeId),
		tNet:     tNet,
	}
}

func (m *Mgr) Start() {
	go m.handleUiCommands()
	go m.readPayloads()
	go m.handleNewlyConnectdNodes()
}

func (m *Mgr) Close() {
	m.tNet.Close()
}

func (m *Mgr) GetId() node.Id {
	return m.node.Id
}

func (m *Mgr) handleNewlyConnectdNodes() {
	for {
		m.conns.AddedNode()
		m.syncNodes()
	}
}

func (m *Mgr) readPayloads() {
	for {
		remoteId, payload := m.conns.ReceivePayload()

		switch p := payload.(type) {
		case *proto.SyncNodes:
			fmt.Println("readPayloads SyncNodes")
			missingConns := findMyMissingConns(*m.conns, p)
			for _, c := range missingConns.GetValues() {
				m.conns.Add(c)
			}
			if remoteIsMissingNodes(*m.conns, p) {
				toSend := m.BuildSyncNodesPayload()
				m.conns.SendPayload(remoteId, &toSend)
			}
		default:
			fmt.Println("readPayloads default case ")
		}
	}
}

func (m *Mgr) BuildSyncNodesPayload() proto.SyncNodes {
	myNodes := m.conns.GetNodes()
	myNodes.Add(m.node)
	toSend := proto.SyncNodes{Nodes: myNodes}
	return toSend
}

func (m *Mgr) addRemoteNode(cmd cmds.User) {
	remoteAddress := node.NewAddress(cmd.Argument)
	m.conns.Add(tnet.NewConn(remoteAddress))
	m.syncNodes()
}

func (n *Mgr) handleUiCommands() {
	for {
		command := <-n.userCmds
		switch command.CmdType {
		case cmds.ConnectTo:
			n.addRemoteNode(command)
		case cmds.AddStorage:
			n.addStorage(command)
		}
	}
}

func (n *Mgr) addStorage(cmd cmds.User) {
	fmt.Println("Received command: add-storage, location:" + cmd.Argument)
}

func (m *Mgr) syncNodes() {
	allIds := m.conns.GetIds()
	for _, id := range allIds.GetValues() {
		payload := m.BuildSyncNodesPayload()
		fmt.Println("mgr.syncNodes to " + id.String())
		m.conns.SendPayload(id, &payload)
	}
}