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

package transfer

import (
	"context"
	"net"
	"strings"

	"os-artificer/saber/internal/controller/server"
	"os-artificer/saber/internal/transfer/config"

	"github.com/spf13/viper"
)

type Service struct {
	svr       *server.AgentServer
	serviceID string
}

// CreateService creates a new transfer service.
func CreateService(ctx context.Context, address string, serviceID string) *Service {
	return &Service{
		svr:       server.New(ctx, address, serviceID),
		serviceID: serviceID,
	}
}

// Run starts the transfer service.
func (s *Service) Run() error {
	return s.svr.Run()
}

// Close stops the transfer service.
func (s *Service) Close() error {
	return s.svr.Close()
}

// loadTransferConfig reads config from ConfigFilePath and returns the listen address.
func loadTransferConfig() string {
	address := ":26689"

	if ConfigFilePath == "" {
		return address
	}

	viper.SetConfigFile(ConfigFilePath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return address
	}

	if err := viper.Unmarshal(&config.Cfg); err != nil {
		return address
	}

	if config.Cfg.Service.ListenAddress != "" {
		address = parseListenAddress(config.Cfg.Service.ListenAddress)
	}

	return address
}

// parseListenAddress parses the listen address.
func parseListenAddress(addr string) string {
	addr = strings.TrimPrefix(addr, "tcp://")
	if _, port, err := net.SplitHostPort(addr); err == nil && port != "" {
		return ":" + port
	}
	return addr
}
