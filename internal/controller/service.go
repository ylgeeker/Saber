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
	"net"
	"strings"

	"os-artificer/saber/internal/controller/config"
	"os-artificer/saber/internal/controller/server"

	"github.com/spf13/viper"
)

// Service is the controller service.
type Service struct {
	svr *server.AgentServer
}

// CreateService creates a new controller service.
func CreateService(ctx context.Context, address string, serviceID string) *Service {
	svr := server.New(ctx, address, serviceID)
	return &Service{
		svr: svr,
	}
}

// Run starts the controller service.
func (s *Service) Run() error {
	return s.svr.Run()
}

// Close stops the controller service.
func (s *Service) Close() error {
	return s.svr.Close()
}

// loadControllerConfig reads config from ConfigFilePath and returns the listen address.
func loadControllerConfig() string {
	address := ":26688"

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
