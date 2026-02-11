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

package controller

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"os-artificer/saber/pkg/constant"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ResponseHandler is called for each AgentResponse received from the controller (server).
type ResponseHandler func(*proto.AgentResponse)

// ControllerClient maintains a gRPC long connection to the controller service with bidirectional
// messaging and automatic reconnect on connection failure.
type ControllerClient struct {
	endpoint             string
	clientID             string
	conn                 *grpc.ClientConn
	stream               grpc.BidiStreamingClient[proto.AgentRequest, proto.AgentResponse]
	ctx                  context.Context
	cancel               context.CancelFunc
	mu                   sync.RWMutex
	closed               bool
	reconnecting         bool
	reconnectAttempts    int
	reconnectInterval    time.Duration
	maxReconnectAttempts int
	onResponse           ResponseHandler
}

// NewControllerClient creates a new controller client. Call Run() to establish the connection and
// start the recv loop; use Send() to send requests and set OnResponse for server messages.
func NewControllerClient(ctx context.Context, endpoint string, clientID string) *ControllerClient {
	ctxCancel, cancel := context.WithCancel(ctx)
	return &ControllerClient{
		endpoint:             endpoint,
		clientID:             clientID,
		ctx:                  ctxCancel,
		cancel:               cancel,
		reconnectInterval:    constant.DefaultClientReconnectInterval,
		maxReconnectAttempts: constant.DefaultClientMaxReconnectAttempts,
	}
}

// OnResponse sets the callback invoked for each AgentResponse received from the server.
func (c *ControllerClient) OnResponse(h ResponseHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onResponse = h
}

func (c *ControllerClient) newConn() (*grpc.ClientConn, error) {
	kacp := keepalive.ClientParameters{
		Time:                constant.DefaultKeepalivePingInterval,
		Timeout:             constant.DefaultPingTimeout,
		PermitWithoutStream: true,
	}
	return grpc.NewClient(
		c.endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(constant.DefaultMaxReceiveMessageSize),
			grpc.MaxCallSendMsgSize(constant.DefaultMaxSendMessageSize),
		),
	)
}

// connect establishes the gRPC connection and Connect() bidi stream, sends the first message with
// clientID, and runs the recv loop until error or context cancel. Caller should run this in a loop
// with reconnect backoff on return.
func (c *ControllerClient) connect() error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("controller client is closed")
	}
	c.mu.RUnlock()

	conn, err := c.newConn()
	if err != nil {
		return fmt.Errorf("controller dial: %w", err)
	}

	client := proto.NewControllerServiceClient(conn)
	stream, err := client.Connect(c.ctx)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("controller Connect stream: %w", err)
	}

	// Server expects first message to carry clientID.
	first := &proto.AgentRequest{
		ClientID: c.clientID,
		Headers:  nil,
		Payload:  nil,
	}
	if err := stream.Send(first); err != nil {
		_ = conn.Close()
		return fmt.Errorf("controller send first message: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.stream = stream
	c.reconnectAttempts = 0
	c.mu.Unlock()

	go c.monitorConnection()

	// Recv loop: server can push AgentResponse at any time.
	for {
		msg, err := stream.Recv()
		if err != nil {
			c.mu.Lock()
			c.conn = nil
			c.stream = nil
			c.mu.Unlock()
			_ = conn.Close()
			return err
		}
		c.mu.RLock()
		cb := c.onResponse
		c.mu.RUnlock()
		if cb != nil {
			cb(msg)
		}
	}
}

// handleDisconnect runs a unified retry loop: backoff then connect() until connection
// succeeds (then connect blocks in recv), or max attempts reached, or client closed.
// First and subsequent connection failures both use this same backoff and retry path.
func (c *ControllerClient) handleDisconnect() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	oldConn := c.conn
	c.conn = nil
	c.stream = nil
	maxAttempts := c.maxReconnectAttempts
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	if oldConn != nil {
		_ = oldConn.Close()
	}

	for attempt := 1; ; attempt++ {
		c.mu.RLock()
		closed := c.closed
		c.mu.RUnlock()
		if closed {
			return
		}
		if maxAttempts > 0 && attempt > maxAttempts {
			logger.Warnf("controller client: max reconnect attempts (%d) reached", maxAttempts)
			return
		}

		backoff := c.reconnectInterval * time.Duration(1<<uint(attempt-1))
		backoff += time.Duration(rand.Int63n(int64(backoff / 2)))
		logger.Infof("controller client: reconnect attempt %d in %v", attempt, backoff)
		time.Sleep(backoff)

		c.mu.RLock()
		closed = c.closed
		c.mu.RUnlock()
		if closed {
			return
		}

		logger.Infof("controller client: reconnecting to %s", c.endpoint)
		if err := c.connect(); err != nil {
			logger.Warnf("controller client: reconnect failed: %v", err)
			continue
		}
		// connect() succeeded and blocks in recv loop until stream error
		logger.Infof("controller client: reconnect successful")
		return
	}
}

// Send sends an AgentRequest to the controller. It is safe to call from multiple goroutines.
func (c *ControllerClient) Send(ctx context.Context, req *proto.AgentRequest) error {
	if req == nil {
		return fmt.Errorf("req is nil")
	}
	c.mu.RLock()
	stream := c.stream
	c.mu.RUnlock()
	if stream == nil {
		return fmt.Errorf("controller client not connected")
	}
	return stream.Send(req)
}

// Run establishes the long-lived connection and runs the recv loop, reconnecting automatically on
// failure until Close() is called or ctx is cancelled. First and subsequent connection failures
// both go through handleDisconnect() for consistent backoff and retry.
func (c *ControllerClient) Run() error {
	for {
		err := c.connect()
		if err == nil {
			// connect() only returns when stream fails or ctx done.
			return c.ctx.Err()
		}

		c.mu.RLock()
		closed := c.closed
		c.mu.RUnlock()
		if closed {
			return nil
		}

		logger.Warnf("controller client: connection lost: %v", err)
		c.handleDisconnect()

		// Reset attempt count so next session gets full retries (never give up permanently).
		c.mu.Lock()
		c.reconnectAttempts = 0
		c.mu.Unlock()
	}
}

// Close shuts down the client and releases the connection.
func (c *ControllerClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.cancel()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.stream = nil
	}
	return nil
}

// monitorConnection periodically checks conn state and triggers reconnect on TransientFailure/Shutdown.
func (c *ControllerClient) monitorConnection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			if c.closed {
				c.mu.RUnlock()
				return
			}
			conn := c.conn
			c.mu.RUnlock()
			if conn == nil {
				return
			}
			state := conn.GetState()
			if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				logger.Warnf("controller client: connection state %s, reconnecting", state.String())
				go c.handleDisconnect()
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}
