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
	"fmt"
	"net"
	"sync"
	"time"

	"os-artificer/saber/pkg/constant"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var ErrConnectionNotFound = errors.New("connection not found")

type AgentServer struct {
	proto.UnimplementedControllerServiceServer

	ctx       context.Context
	address   string
	serviceID string
	manager   *ConnectionManager
	grpcSvr   *grpc.Server
}

func New(ctx context.Context, address string, serviceID string) *AgentServer {
	return &AgentServer{
		ctx:       ctx,
		address:   address,
		serviceID: serviceID,
		manager:   NewConnectionManager(),
	}
}

// extractClientInfo is unused; clientID is read from first AgentRequest in Connect.
func (s *AgentServer) extractClientInfo(ctx context.Context) (string, map[string]string, error) {
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	metadata := make(map[string]string)
	_ = ctx
	return clientID, metadata, nil
}

// SendToClient sends an AgentResponse to the connected client identified by clientID.
// Returns ErrConnectionNotFound if no connection exists for clientID, or the error from
// Connection.TrySend (e.g. ErrConnectionClosed, ErrSendChanFull).
func (s *AgentServer) SendToClient(ctx context.Context, clientID string, resp *proto.AgentResponse) error {
	if resp == nil {
		return fmt.Errorf("resp is nil")
	}
	conn, exists := s.manager.Get(clientID)
	if !exists {
		return ErrConnectionNotFound
	}
	return conn.TrySend(resp)
}

func (s *AgentServer) Connect(stream proto.ControllerService_ConnectServer) error {
	// Read first message to get clientID; same clientID reconnecting will close previous session.
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	clientID := req.GetClientID()
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}
	metadata := make(map[string]string)
	if h := req.GetHeaders(); h != nil {
		for k, v := range h {
			metadata[k] = v
		}
	}

	conn := &Connection{
		ClientID:   clientID,
		Stream:     stream,
		SendChan:   make(chan *proto.AgentResponse, 100),
		LastActive: time.Now(),
		Metadata:   metadata,
		FirstReq:   req,
	}

	s.manager.Register(clientID, conn) // closes any existing connection with same clientID
	defer s.manager.Unregister(clientID)

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()
		conn.sendMessages(s.ctx)
	}()

	go func() {
		defer wg.Done()
		conn.receiveMessages(s.ctx)
	}()

	wg.Wait()
	return nil
}

func (s *AgentServer) Run() error {

	kasp := keepalive.ServerParameters{
		Time:    constant.DefaultKeepalivePingInterval,
		Timeout: constant.DefaultPingTimeout,
	}

	kacp := keepalive.EnforcementPolicy{
		MinTime:             constant.DefaultKeepalivePingInterval,
		PermitWithoutStream: true,
	}

	svr := grpc.NewServer(
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kacp),
		grpc.MaxRecvMsgSize(constant.DefaultMaxReceiveMessageSize),
		grpc.MaxSendMsgSize(constant.DefaultMaxSendMessageSize),
	)

	proto.RegisterControllerServiceServer(svr, s)
	s.grpcSvr = svr
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	logger.Infof("Server listening at %v", lis.Addr())
	return svr.Serve(lis)
}

// Close stops the gRPC server gracefully for use with controller’s setupGracefulShutdown.
func (s *AgentServer) Close() error {
	if s.grpcSvr != nil {
		s.grpcSvr.GracefulStop()
		s.grpcSvr = nil
	}
	return nil
}
