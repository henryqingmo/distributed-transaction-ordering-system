package manager

import (
	"cs425_mp1/internal/config"
	"fmt"
	"net"
	"sync"
	"time"
)

type Manager struct {
	self     config.NodeInfo
	peers    map[string]net.Conn // nodeID of TCP connection
	peerIDs  []string            // all peer IDs
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
		go m.acceptHandshake(conn)
	}
}


func (m *Manager) acceptHandshake(conn net.Conn) {
	msg, err := ReadMsg(conn)
	if err != nil || msg.Type != TypeHandshake || msg.SenderID == "" {
		conn.Close()
		return
	}
	peerID := msg.SenderID
	m.mu.Lock()
	m.peers[peerID] = conn
	m.mu.Unlock()
	m.handleConnection(peerID, conn)
}

func (m *Manager) ConnectToPeers(nodes []config.NodeInfo) error {
	for _, node := range nodes {
		if node.ID == m.self.ID {
			continue
		}
		m.peerIDs = append(m.peerIDs, node.ID)
		addr := net.JoinHostPort(node.Host, node.Port)
		conn, err := dialWithRetry(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to peer %s: %v", node.ID, err)
		}
		// Send handshake so the peer knows who we are.
		if err := WriteMsg(conn, Message{Type: TypeHandshake, SenderID: m.self.ID}); err != nil {
			conn.Close()
			return fmt.Errorf("handshake to peer %s failed: %v", node.ID, err)
		}
		m.mu.Lock()
		m.peers[node.ID] = conn
		m.mu.Unlock()
		go m.handleConnection(node.ID, conn)
	}
	return nil
}

// dialWithRetry keeps trying to connect until successful, with 500ms between attempts.
// Heartbeat 
func dialWithRetry(addr string) (net.Conn, error) {
	for {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			return conn, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *Manager) Broadcast(msg Message) {
	for _, id := range m.peerIDs {
		if err := m.Send(id, msg); err != nil {
			m.markDead(id)
		}
	}
}

func (m *Manager) Send(nodeID string, msg Message) error {
	m.mu.Lock()
	conn, ok := m.peers[nodeID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("peer %s is dead", nodeID)
	}
	if err := WriteMsg(conn, msg); err != nil {
		m.markDead(nodeID)
		return err
	}
	return nil
}

// markDead removes the peer from the live map and signals the failure channel 
func (m *Manager) markDead(nodeID string) {
	m.mu.Lock()
	_, alive := m.peers[nodeID]
	if alive {
		m.peers[nodeID].Close()
		delete(m.peers, nodeID)
	}
	m.mu.Unlock()
	if alive {
		m.failures <- nodeID
	}
}

func (m *Manager) handleConnection(nodeID string, conn net.Conn) {
	defer m.markDead(nodeID)
	for {
		msg, err := ReadMsg(conn)
		if err != nil {
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
