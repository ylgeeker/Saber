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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"os-artificer/saber/internal/controller/apm"
	"os-artificer/saber/internal/controller/config"
	"os-artificer/saber/internal/controller/server"
	"os-artificer/saber/pkg/discovery"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/sbnet"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// controllerUnmarshalOpt composes default viper hooks with string->Endpoint so
// service.listenAddress (string) unmarshals into sbnet.Endpoint.
var controllerUnmarshalOpt = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	sbnet.StringToEndpointHookFunc(),
))

// Service is the controller service.
type Service struct {
	svr             *server.AgentServer
	apm             *apm.APM
	discoveryClient *discovery.Client
	registry        *discovery.Registry
}

// CreateService creates a new controller service. APM is initialized later in Run() via InitAPM().
func CreateService(ctx context.Context, address sbnet.Endpoint, serviceID string) *Service {
	svr := server.New(ctx, address, serviceID)
	return &Service{
		svr:      svr,
		apm:      nil,
		registry: nil,
	}
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

// InitAPM creates the APM service from config.Cfg.APM and sets s.apm. Business metrics are in internal/controller/apm/metrics.go.
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
	if err := viper.Unmarshal(&config.Cfg, controllerUnmarshalOpt); err != nil {
		return err
	}
	if err := s.InitLogger(); err != nil {
		return err
	}
	logger.Infof("config reloaded")
	return nil
}

// buildDiscoveryTLS builds *tls.Config from discovery config for etcd https endpoints.
// Returns (nil, nil) when TLS is not needed (no UseTLS, no cert paths, no InsecureSkipVerify).
func buildDiscoveryTLS(cfg *config.DiscoveryConfig) (*tls.Config, error) {
	if !cfg.UseTLS && !cfg.InsecureSkipVerify && cfg.EtcdCACert == "" && cfg.EtcdCert == "" && cfg.EtcdKey == "" {
		return nil, nil
	}

	tlsCfg := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}
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

// RegisterSelf registers the controller service with the discovery service (etcd).
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

	tlsCfg, err := buildDiscoveryTLS(cfg)
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
		discovery.OptionServiceName("controller"),
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
	listenAddr := config.Cfg.Service.ListenAddress.String()
	// Use longer timeout for first-time connect + auth + grant + put (2x DialTimeout).
	registerTimeout := max(2*cfg.DialTimeout, cfg.DialTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), registerTimeout)
	defer cancel()

	if err := s.registry.SetService(ctx, listenAddr); err != nil {
		s.registry.Close()
		s.registry = nil
		return err
	}

	logger.Infof("controller registered to discovery, serviceID=%s, listenAddress=%s", serviceID, listenAddr)
	return nil
}

// Run starts the controller service. It initializes logger and APM, then starts APM (if enabled) in a goroutine and runs the gRPC server.
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

	return s.svr.Run()
}

// Close stops the controller service (registry, APM, then the gRPC server).
func (s *Service) Close() error {
	if s.registry != nil {
		s.registry.Close()
		s.registry = nil
	}
	if s.apm != nil {
		_ = s.apm.Close()
	}
	return s.svr.Close()
}

// loadControllerConfig reads config from ConfigFilePath and returns the listen address.
func loadControllerConfig() {
	if ConfigFilePath == "" {
		return
	}

	viper.SetConfigFile(ConfigFilePath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Failed to read config file: %v", err)
		return
	}

	if err := viper.Unmarshal(&config.Cfg, controllerUnmarshalOpt); err != nil {
		logger.Errorf("Failed to unmarshal config: %v", err)
		return
	}

	logger.Infof("Loaded controller config: %+v", config.Cfg)
}
