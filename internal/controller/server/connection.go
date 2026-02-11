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

import (
	"context"
	"errors"
	"sync"
	"time"

	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrSendChanFull     = errors.New("send channel full")
)

type Connection struct {
	ClientID   string
	Stream     proto.ControllerService_ConnectServer
	SendChan   chan *proto.AgentResponse
	LastActive time.Time
	Metadata   map[string]string
	FirstReq   *proto.AgentRequest // first message already read in Connect to get clientID
	mu         sync.RWMutex
	closed     bool
}

func (c *Connection) sendMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-c.SendChan:
			if !ok || c.isClosed() {
				return
			}

			c.updateLastActive()

			if err := c.Stream.Send(msg); err != nil {
				c.close()
				return
			}
		}
	}
}

func (c *Connection) receiveMessages(ctx context.Context) {
	// handle first message already read in Connect (for clientID)
	c.mu.Lock()
	firstReq := c.FirstReq
	c.FirstReq = nil
	c.mu.Unlock()
	if firstReq != nil {
		c.updateLastActive()
		logger.Debugf("Received from %s: %v", c.ClientID, firstReq.GetPayload())
	}

	for {
		select {
		case <-ctx.Done():
			return

		default:
			if c.isClosed() {
				return
			}

			msg, err := c.Stream.Recv()
			if err != nil {
				c.close()
				return
			}

			c.updateLastActive()

			logger.Debugf("Received from %s: %v", c.ClientID, msg.GetPayload())
		}
	}
}

func (c *Connection) updateLastActive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActive = time.Now()
}

func (c *Connection) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

func (c *Connection) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.SendChan)
	}
}

// TrySend enqueues msg for sending to the client. It checks isClosed() to avoid writing to a
// closed channel and uses non-blocking send so callers are not stuck when the channel is full.
// Returns ErrConnectionClosed if the connection is closed, ErrSendChanFull if SendChan is full.
func (c *Connection) TrySend(msg *proto.AgentResponse) error {
	if c.isClosed() {
		return ErrConnectionClosed
	}
	select {
	case c.SendChan <- msg:
		return nil
	default:
		return ErrSendChanFull
	}
}
