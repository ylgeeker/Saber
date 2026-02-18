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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"os-artificer/saber/internal/transfer/apm"
	"os-artificer/saber/internal/transfer/config"
	"os-artificer/saber/internal/transfer/sink"
	"os-artificer/saber/internal/transfer/source"
	"os-artificer/saber/pkg/discovery"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/sbnet"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
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
	svr             *source.AgentSource
	handler         *source.ConnectionHandler
	sink            sink.Sink
	serviceID       string
	apm             *apm.APM
	discoveryClient *discovery.Client
	registry        *discovery.Registry
}

// CreateService creates a new transfer service. APM is initialized later in Run() via InitAPM().
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
			svr:             source.NewAgentSource(*address, nil),
			handler:         handler,
			sink:            snk,
			serviceID:       serviceID,
			apm:             nil,
			discoveryClient: nil,
			registry:        nil,
		}, nil
	}

	return nil, fmt.Errorf("source type not supported")
}

// InitLogger initializes the global logger from config.Cfg.Log (pkg/logger).
func (s *Service) InitLogger() error {
	cfg := &config.Cfg.Log
	logCfg := logger.Config{
		Filename:   cfg.FileName,
		LogLevel:   cfg.LogLevel,
		MaxSizeMB:  cfg.FileSize,
		MaxBackups: cfg.MaxBackupCount,
		MaxAge:     cfg.MaxBackupAge,
	}
	l := logger.NewZapLogger(logCfg)
	logger.SetLogger(l)
	return nil
}

// InitAPM creates the APM service from config.Cfg.APM and sets s.apm. Business metrics are in internal/transfer/apm/metrics.go.
func (s *Service) InitAPM() error {
	cfg := &config.Cfg.APM
	s.apm = apm.NewAPM(cfg.Enabled, cfg.Endpoint)
	return nil
}

// ReloadConfig re-reads config from ConfigFilePath and re-inits logger (for SIGHUP).
func (s *Service) ReloadConfig() error {
	if ConfigFilePath == "" {
		return nil
	}
	viper.SetConfigFile(ConfigFilePath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&config.Cfg, transferUnmarshalOpt); err != nil {
		return err
	}
	if err := s.InitLogger(); err != nil {
		return err
	}
	logger.Infof("config reloaded")
	return nil
}

// buildDiscoveryTLS builds *tls.Config from discovery config for etcd https endpoints.
// Returns (nil, nil) when TLS is not needed (no UseTLS, no cert paths, no InsecureSkipVerify, no https endpoint).
// If any endpoint starts with "https://" but TLS was not configured, a TLS config with InsecureSkipVerify is used.
func buildDiscoveryTLS(cfg *config.DiscoveryConfig, endpoints []string) (*tls.Config, error) {
	needTLS := cfg.UseTLS || cfg.InsecureSkipVerify || cfg.EtcdCACert != "" || cfg.EtcdCert != "" || cfg.EtcdKey != ""
	if !needTLS {
		for _, ep := range endpoints {
			if strings.HasPrefix(ep, "https://") {
				needTLS = true
				break
			}
		}
	}
	if !needTLS {
		return nil, nil
	}
	insecureSkip := cfg.InsecureSkipVerify
	if !insecureSkip && cfg.EtcdCACert == "" && cfg.EtcdCert == "" && cfg.EtcdKey == "" {
		for _, ep := range endpoints {
			if strings.HasPrefix(ep, "https://") {
				insecureSkip = true
				break
			}
		}
	}
	tlsCfg := &tls.Config{InsecureSkipVerify: insecureSkip}
	if cfg.EtcdCACert != "" {
		b, err := os.ReadFile(cfg.EtcdCACert)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("no valid CA certs in %s", cfg.EtcdCACert)
		}
		tlsCfg.RootCAs = pool
	}
	if cfg.EtcdCert != "" && cfg.EtcdKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.EtcdCert, cfg.EtcdKey)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return tlsCfg, nil
}

// getTransferListenAddr returns the first agent source endpoint from config as the listen address for discovery.
func getTransferListenAddr() string {
	for _, src := range config.Cfg.Source {
		if src.Type != "agent" {
			continue
		}
		if ep, ok := src.Config["endpoint"].(string); ok && ep != "" {
			return ep
		}
		if ep, ok := src.Config["endpoints"].(string); ok && ep != "" {
			return ep
		}
		break
	}
	return ""
}

// RegisterSelf registers the transfer service with the discovery service (etcd).
func (s *Service) RegisterSelf() error {
	cfg := &config.Cfg.Discovery
	endpoints := strings.Split(cfg.EtcdEndpoint, ",")
	for i, ep := range endpoints {
		endpoints[i] = strings.TrimSpace(ep)
	}
	if len(endpoints) == 0 || endpoints[0] == "" {
		logger.Warnf("discovery etcdEndpoint is empty, skip register")
		return nil
	}

	tlsCfg, err := buildDiscoveryTLS(cfg, endpoints)
	if err != nil {
		return err
	}

	serviceID := uuid.New().String()
	opts := []discovery.Option{
		discovery.OptionEndpoints(endpoints),
		discovery.OptionUser(cfg.EtcdUser),
		discovery.OptionPassword(cfg.EtcdPassword),
		discovery.OptionDialTimeout(cfg.DialTimeout),
		discovery.OptionAutoSyncInterval(cfg.AutoSyncInterval),
		discovery.OptionKeepAliveTime(cfg.DialKeepAliveTime),
		discovery.OptionKeepAliveTimeout(cfg.DialKeepAliveTimeout),
		discovery.OptionRegistryRootKeyPrefix(cfg.RegistryRootKeyPrefix),
		discovery.OptionServiceName("transfer"),
		discovery.OptionServiceID(serviceID),
		discovery.OptionTTL(int(cfg.RegistryTTL)),
		discovery.OptionLogger(logger.GetOriginLogger()),
	}
	if tlsCfg != nil {
		opts = append(opts, discovery.OptionTLS(tlsCfg))
	}
	cli, err := discovery.NewClientWithOptions(opts...)
	if err != nil {
		return err
	}

	s.discoveryClient = cli
	s.registry = cli.CreateRegistry()
	listenAddr := getTransferListenAddr()
	registerTimeout := max(2*cfg.DialTimeout, cfg.DialTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), registerTimeout)
	defer cancel()

	if err := s.registry.SetService(ctx, listenAddr); err != nil {
		s.registry.Close()
		s.registry = nil
		return err
	}

	logger.Infof("transfer registered to discovery, serviceID=%s, listenAddress=%s", serviceID, listenAddr)
	return nil
}

// Run starts the transfer service. It initializes logger and APM, then starts APM (if enabled) in a goroutine and runs the source.
func (s *Service) Run() error {
	if err := s.InitLogger(); err != nil {
		return err
	}
	if err := s.InitAPM(); err != nil {
		return err
	}
	if s.apm != nil && s.apm.IsEnabled() {
		go func() {
			_ = s.apm.Run()
		}()
	}

	if err := s.RegisterSelf(); err != nil {
		return err
	}

	return s.svr.Run(context.Background(), s.handler)
}

// Close stops the transfer service (registry, APM, then connection handler and sink).
func (s *Service) Close() error {
	if s.registry != nil {
		s.registry.Close()
		s.registry = nil
	}
	if s.apm != nil {
		_ = s.apm.Close()
	}
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
