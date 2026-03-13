package manager

import (
	"cs425_mp1/internal/config"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Manager struct {
	self     config.NodeInfo
	peers    map[string]net.Conn // nodeID -> TCP connection
	peerIDs  []string            // all peer IDs, set once by ConnectToPeers
	mu       sync.Mutex          // protects peers map
	inbox    chan Message        // incoming messages for the node to consume
	failures chan string         // IDs of peers that died
	listener net.Listener
}

func NewManager(self config.NodeInfo, inboxSize int) *Manager {
	return &Manager{
		self:     self,
		peers:    make(map[string]net.Conn),
		peerIDs:  []string{},
		inbox:    make(chan Message, inboxSize),
		failures: make(chan string, inboxSize),
	}
}

func (m *Manager) Listen() error {
	addr := net.JoinHostPort(m.self.Host, m.self.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.listener = ln
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}
		go m.handleConnection(conn)
	}
}

func (m *Manager) ConnectToPeers(nodes []config.NodeInfo) error {
	for _, node := range nodes {
		if node.ID == m.self.ID {
			continue
		}
		m.peerIDs = append(m.peerIDs, node.ID)
		// Establish TCP connection to peer
		conn, err := net.Dial("tcp", net.JoinHostPort(node.Host, node.Port))
		if err != nil {
			return fmt.Errorf("failed to connect to peer %s: %v", node.ID, err)
		}
		m.mu.Lock()
		m.peers[node.ID] = conn
		m.mu.Unlock()
	}
	return nil
}

func (m *Manager) Broadcast(msg Message) {
	for _, id := range m.peerIDs {
		if err := m.Send(id, msg); err != nil {
			m.failures <- id
		}
	}
}

func (m *Manager) Send(nodeID string, msg Message) error {
	m.mu.Lock()
	conn, ok := m.peers[nodeID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("no connection for peer %s", nodeID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func (m *Manager) handleConnection(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			return
		}
		m.inbox <- msg
	}
}

func (m *Manager) Inbox() <-chan Message   { return m.inbox }
func (m *Manager) Failures() <-chan string { return m.failures }
func (m *Manager) Close() {
	if m.listener != nil {
		m.listener.Close()
	}
}
