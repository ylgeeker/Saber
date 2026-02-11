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

package source

import (
	"context"
	"fmt"
	"io"
	"net"

	"os-artificer/saber/pkg/constant"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
)

var _ Source = (*AgentSource)(nil)

// AgentSource implements Source by accepting gRPC PushData streams from agents,
// tracking long-lived connections via ConnectionManager, and delivering each
// TransferRequest to the Handler.
type AgentSource struct {
	address string
	connMgr *ConnectionManager
}

// NewAgentSource returns a Source that listens on address and serves TransferService PushData.
// Long-lived client connections are tracked by an internal ConnectionManager.
// Pass nil for connMgr to use a default manager with no connection limit.
func NewAgentSource(address string, connMgr *ConnectionManager) *AgentSource {
	if connMgr == nil {
		connMgr = NewConnectionManager(0)
	}
	return &AgentSource{address: address, connMgr: connMgr}
}

// Run starts the gRPC server and blocks until ctx is done or server stops.
// Each received TransferRequest is passed to h.OnTransferRequest.
func (p *AgentSource) Run(ctx context.Context, h Handler) error {
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

	pushSrv := &pushDataServer{handler: h, connMgr: p.connMgr}
	proto.RegisterTransferServiceServer(svr, pushSrv)

	lis, err := net.Listen("tcp", p.address)
	if err != nil {
		return err
	}

	logger.Infof("Server listening at %v", lis.Addr())

	go func() {
		<-ctx.Done()
		svr.GracefulStop()
	}()

	return svr.Serve(lis)
}

// pushDataServer implements proto.TransferServiceServer and forwards each request to Handler.
type pushDataServer struct {
	proto.UnimplementedTransferServiceServer
	handler Handler
	connMgr *ConnectionManager
}

func (s *pushDataServer) PushData(stream proto.TransferService_PushDataServer) error {
	ctx := stream.Context()
	remoteAddr := ""
	if p, ok := peer.FromContext(ctx); ok {
		remoteAddr = p.Addr.String()
	}

	connID := fmt.Sprintf("%s#%d", remoteAddr, s.connMgr.NextID())
	meta := &ConnMeta{RemoteAddr: remoteAddr}
	if err := s.connMgr.Register(connID, meta); err != nil {
		logger.Warnf("connection rejected: %s, err: %v", connID, err)
		return err
	}

	var clientID string
	defer func() {
		s.connMgr.Unregister(connID)
		logger.Infof("connection closed: %s (clientID=%s), active=%d", connID, clientID, s.connMgr.ActiveCount())
	}()

	logger.Infof("connection established: %s, active=%d", connID, s.connMgr.ActiveCount())

	for {
		select {
		case <-ctx.Done():
			logger.Infof("stream context done: %s", connID)
			return nil

		default:
			req, err := stream.Recv()
			if err == io.EOF {
				return nil
			}

			if err != nil {
				logger.Warnf("stream recv error: %s, err: %v", connID, err)
				return err
			}

			if clientID == "" && req.GetClientID() != "" {
				clientID = req.GetClientID()
				s.connMgr.UpdateMeta(connID, &ConnMeta{ClientID: clientID})
			}

			if len(req.GetPayload()) == 0 && req.GetClientID() == "" {
				continue
			}

			if err := s.handler.OnTransferRequest(req); err != nil {
				logger.Warnf("handler OnTransferRequest failed: %s, err: %v", connID, err)
			}
		}
	}
}
