package manager

import (
	"cs425_mp1/internal/config"
	"net"
	"sync"
)

type Manager struct {
    self      config.NodeInfo
    peers     map[string]net.Conn   // nodeID -> TCP connection
    mu        sync.Mutex            // protects peers map
    inbox     chan Message           // incoming messages for the node to consume
    failures  chan string            // IDs of peers that died
    listener  net.Listener
}

func NewManager(self config.NodeInfo, inboxSize int) *Manager {
    return &Manager{
        self:     self,
        peers:    make(map[string]net.Conn),
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
        // Establish TCP connection to peer
        conn, err := net.Dial("tcp", net.JoinHostPort(node.Host, node.Port))
        if err != nil {
            return fmt.Errorf("failed to connect to peer %s: %v", node.ID, err)
        }
        m.mu.Lock()
        m.peers[node.ID] = conn
        m.mu.Unlock()
    }
 }
func (m *Manager) Broadcast(msg Message) { }
func (m *Manager) Send(nodeID string, msg Message) error { ... }
func (m *Manager) Inbox() <-chan Message { ... }
func (m *Manager) Failures() <-chan string { ... }
func (m *Manager) Close() { ... }...