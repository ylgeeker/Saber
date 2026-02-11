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

package reporter

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"os-artificer/saber/internal/agent/config"
	"os-artificer/saber/pkg/constant"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"
	"os-artificer/saber/pkg/tools"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var _ Reporter = (*TransferReporter)(nil)

func init() {
	RegisterReporter("transfer", newTransferReporterFromOpts)
}

// newTransferReporterFromOpts builds a TransferReporter from config.ReporterOpts (used by RegisterReporter).
func newTransferReporterFromOpts(ctx context.Context, opts any) (Reporter, error) {
	o, ok := opts.(*config.ReporterOpts)
	if !ok {
		return nil, fmt.Errorf("transfer reporter expects *config.ReporterOpts, got %T", opts)
	}
	endpoints, _ := o.Config["endpoints"].(string)
	if endpoints == "" {
		return nil, fmt.Errorf("transfer reporter config missing or invalid endpoints")
	}

	poolSize := constant.DefaultTransferPoolSize
	if v, ok := o.Config["pool_size"]; ok {
		if n, ok := toInt(v); ok && n > 0 {
			poolSize = n
		}
	}

	if v, ok := o.Config["connection_count"]; ok && poolSize == constant.DefaultTransferPoolSize {
		if n, ok := toInt(v); ok && n > 0 {
			poolSize = n
		}
	}

	clientID, err := tools.MachineID("saber-agent")
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine-id: %v", err)
	}

	return NewTransferReporter(ctx, endpoints, clientID, poolSize)
}

func toInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}

// poolEntry holds one gRPC connection and its PushData stream for the pool.
type poolEntry struct {
	mu                sync.RWMutex
	conn              *grpc.ClientConn
	client            proto.TransferServiceClient
	stream            proto.TransferService_PushDataClient
	reconnecting      bool
	reconnectAttempts int
}

// TransferReporter is the reporter for transfer using a pool of gRPC connections.
type TransferReporter struct {
	serverAddr           string
	poolSize             int
	pool                 []*poolEntry
	nextIndex            atomic.Uint32
	ctx                  context.Context
	cancel               context.CancelFunc
	clientId             string
	closed               bool
	mu                   sync.RWMutex
	reconnectInterval    time.Duration
	maxReconnectAttempts int
}

// NewTransferReporter creates a new transfer reporter with a connection pool.
func NewTransferReporter(ctx context.Context, serverAddr string, clientId string, poolSize int) (*TransferReporter, error) {
	if poolSize <= 0 {
		poolSize = constant.DefaultTransferPoolSize
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	pool := make([]*poolEntry, poolSize)
	for i := range pool {
		pool[i] = &poolEntry{}
	}

	return &TransferReporter{
		serverAddr:           serverAddr,
		poolSize:             poolSize,
		pool:                 pool,
		ctx:                  ctxCancel,
		cancel:               cancel,
		clientId:             clientId,
		reconnectInterval:    constant.DefaultClientReconnectInterval,
		maxReconnectAttempts: constant.DefaultClientMaxReconnectAttempts,
	}, nil
}

func (c *TransferReporter) newConn() (*grpc.ClientConn, error) {
	kacp := keepalive.ClientParameters{
		Time:                constant.DefaultKeepalivePingInterval,
		Timeout:             constant.DefaultPingTimeout,
		PermitWithoutStream: true,
	}

	return grpc.NewClient(
		c.serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(constant.DefaultMaxReceiveMessageSize),
			grpc.MaxCallSendMsgSize(constant.DefaultMaxSendMessageSize),
		),
	)
}

// connectSlot establishes connection and stream for one pool slot.
func (c *TransferReporter) connectSlot(i int) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	conn, err := c.newConn()
	if err != nil {
		return err
	}

	client := proto.NewTransferServiceClient(conn)
	stream, err := client.PushData(c.ctx)
	if err != nil {
		_ = conn.Close()
		return err
	}

	entry := c.pool[i]
	entry.mu.Lock()
	entry.conn = conn
	entry.client = client
	entry.stream = stream
	entry.reconnectAttempts = 0
	entry.mu.Unlock()

	go c.sendConnectionEstablished(i)
	go c.monitorConnection(i)

	return nil
}

func (c *TransferReporter) handleDisconnect(i int) {
	entry := c.pool[i]
	entry.mu.Lock()
	if c.closed {
		entry.mu.Unlock()
		return
	}
	if entry.reconnecting {
		entry.mu.Unlock()
		return
	}
	entry.reconnecting = true
	oldConn := entry.conn
	entry.conn = nil
	entry.client = nil
	entry.stream = nil
	entry.reconnectAttempts++
	attempts := entry.reconnectAttempts
	maxAttempts := c.maxReconnectAttempts
	entry.mu.Unlock()

	defer func() {
		entry.mu.Lock()
		entry.reconnecting = false
		entry.mu.Unlock()
	}()

	if oldConn != nil {
		_ = oldConn.Close()
	}

	if maxAttempts > 0 && attempts > maxAttempts {
		logger.Warnf("transfer pool slot %d: max reconnect attempts (%d) reached", i, maxAttempts)
		return
	}

	backoffInterval := c.reconnectInterval * time.Duration(1<<uint(attempts-1))
	backoffInterval += time.Duration(rand.Int63n(int64(backoffInterval / 2)))

	logger.Infof("transfer pool slot %d: reconnect attempt %d in %v", i, attempts, backoffInterval)
	time.Sleep(backoffInterval)

	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return
	}

	logger.Infof("transfer pool slot %d: attempting reconnect", i)
	err := c.connectSlot(i)
	if err != nil {
		logger.Warnf("transfer pool slot %d: reconnect failed: %v", i, err)
		go c.handleDisconnect(i)
	} else {
		logger.Infof("transfer pool slot %d: reconnect successful", i)
	}
}

func (c *TransferReporter) monitorConnection(i int) {
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
			c.mu.RUnlock()

			entry := c.pool[i]
			entry.mu.RLock()
			conn := entry.conn
			entry.mu.RUnlock()
			if conn == nil {
				return
			}
			state := conn.GetState()
			if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				logger.Warnf("transfer pool slot %d: connection state %s, starting reconnect", i, state.String())
				go c.handleDisconnect(i)
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *TransferReporter) sendConnectionEstablished(i int) {
	msg := &proto.TransferRequest{}

	entry := c.pool[i]
	entry.mu.RLock()
	stream := entry.stream
	entry.mu.RUnlock()

	if stream == nil {
		return
	}

	if err := stream.Send(msg); err != nil {
		logger.Errorf("transfer pool slot %d: error sending connection established: %v", i, err)
	}
}

func (c *TransferReporter) Run() error {
	for i := 0; i < c.poolSize; i++ {
		i := i
		tools.Go(func() {
			for {
				err := c.connectSlot(i)
				if err == nil {
					return
				}
				c.mu.RLock()
				closed := c.closed
				c.mu.RUnlock()
				if closed {
					return
				}
				logger.Warnf("transfer pool slot %d: initial connect failed: %v, retrying", i, err)
				time.Sleep(c.reconnectInterval)
			}
		})
	}

	<-c.ctx.Done()
	return c.ctx.Err()
}

func (c *TransferReporter) SendMessage(ctx context.Context, content []byte) error {
	c.mu.RLock()
	closed := c.closed
	clientId := c.clientId
	poolSize := c.poolSize
	c.mu.RUnlock()

	if closed {
		return fmt.Errorf("client is closed")
	}
	if poolSize == 0 {
		return fmt.Errorf("no pool slots")
	}

	msg := &proto.TransferRequest{
		ClientID: clientId,
		Payload:  content,
	}

	for try := 0; try < poolSize; try++ {
		idx := int(c.nextIndex.Add(1) % uint32(poolSize))
		entry := c.pool[idx]

		entry.mu.RLock()
		stream := entry.stream
		entry.mu.RUnlock()

		if stream == nil {
			continue
		}

		err := stream.Send(msg)
		if err != nil {
			go c.handleDisconnect(idx)
			continue
		}
		return nil
	}

	return fmt.Errorf("no healthy stream: all %d slots unavailable or send failed", poolSize)
}

func (c *TransferReporter) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.cancel != nil {
		c.cancel()
	}

	var firstErr error
	for _, entry := range c.pool {
		entry.mu.Lock()
		conn := entry.conn
		entry.conn = nil
		entry.client = nil
		entry.stream = nil
		entry.mu.Unlock()
		if conn != nil {
			if err := conn.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// GetConnectionState returns the state of the first non-nil connection in the pool.
func (c *TransferReporter) GetConnectionState() connectivity.State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return connectivity.Shutdown
	}
	for _, entry := range c.pool {
		entry.mu.RLock()
		conn := entry.conn
		entry.mu.RUnlock()
		if conn != nil {
			return conn.GetState()
		}
	}
	return connectivity.Idle
}
