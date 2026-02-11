/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package server

import "sync"

// ConnectionManager manages client connections by clientID. It is safe for concurrent use.
type ConnectionManager struct {
	connections map[string]*Connection
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
	}
}

// Register registers a connection for the given clientID. If a connection already exists
// for clientID, it is closed before being replaced.
func (m *ConnectionManager) Register(clientID string, conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, exists := m.connections[clientID]; exists {
		existing.close()
	}
	m.connections[clientID] = conn
}

// Unregister removes the connection for clientID and closes it.
func (m *ConnectionManager) Unregister(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[clientID]; exists {
		conn.close()
		delete(m.connections, clientID)
	}
}

// Get returns the connection for clientID, or (nil, false) if not found.
func (m *ConnectionManager) Get(clientID string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[clientID]
	return conn, exists
}
