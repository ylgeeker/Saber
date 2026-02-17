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
	"fmt"

	"os-artificer/saber/internal/transfer/config"
	"os-artificer/saber/internal/transfer/sink"
	"os-artificer/saber/internal/transfer/source"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/sbnet"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// transferUnmarshalOpt composes default viper hooks with string->Endpoint so
// service.listenAddress (string) unmarshals into sbnet.Endpoint.
var transferUnmarshalOpt = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	sbnet.StringToEndpointHookFunc(),
))

type Service struct {
	svr       *source.AgentSource
	handler   *source.ConnectionHandler
	sink      sink.Sink
	serviceID string
}

// CreateService creates a new transfer service.
// Sink is built from config.Cfg.Sink; if creation fails, returns an error.
func CreateService(ctx context.Context, serviceID string) (*Service, error) {
	snk, err := NewSinkFromConfig(config.Cfg.Sink)
	if err != nil {
		return nil, err
	}

	handler := source.NewConnectionHandler(snk)

	for _, src := range config.Cfg.Source {
		if src.Type != "agent" {
			continue
		}

		address, err := sbnet.NewEndpointFromString(src.Config["endpoint"].(string))
		if err != nil {
			return nil, err
		}

		return &Service{
			svr:       source.NewAgentSource(*address, nil),
			handler:   handler,
			sink:      snk,
			serviceID: serviceID,
		}, nil
	}

	return nil, fmt.Errorf("source type not supported")
}

// Run starts the transfer service.
func (s *Service) Run() error {
	return s.svr.Run(context.Background(), s.handler)
}

// Close stops the transfer service: closes the connection handler then the sink.
func (s *Service) Close() error {
	if s.handler != nil {
		s.handler.Close()
		s.handler = nil
	}
	if s.sink != nil {
		_ = s.sink.Close()
		s.sink = nil
	}
	return nil
}

// loadTransferConfig reads config from ConfigFilePath into config.Cfg.
func loadTransferConfig() {
	if ConfigFilePath == "" {
		return
	}

	viper.SetConfigFile(ConfigFilePath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Failed to read config file: %v", err)
		return
	}

	if err := viper.Unmarshal(&config.Cfg, transferUnmarshalOpt); err != nil {
		logger.Errorf("Failed to unmarshal config: %v", err)
		return
	}

	logger.Infof("Loaded transfer config: %+v", config.Cfg)
}
