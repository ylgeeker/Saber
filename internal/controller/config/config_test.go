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

package config

import (
	"testing"

	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/sbnet"

	"gopkg.in/yaml.v3"
)

func TestDefaultCfg(t *testing.T) {
	if Cfg.Name != "Controller" {
		t.Errorf("Cfg.Name = %q, want %q", Cfg.Name, "Controller")
	}
	if Cfg.Version != "v1.0.0" {
		t.Errorf("Cfg.Version = %q, want %q", Cfg.Version, "v1.0.0")
	}
	if Cfg.Service.ListenAddress.Protocol != "tcp" || Cfg.Service.ListenAddress.Host != "127.0.0.1" || Cfg.Service.ListenAddress.Port != 26689 {
		t.Errorf("Cfg.Service.ListenAddress = %+v, want tcp://127.0.0.1/26689", Cfg.Service.ListenAddress)
	}
	if Cfg.Discovery.EtcdEndpoint != "http://etcd-server:2379" {
		t.Errorf("Cfg.Discovery.EtcdEndpoint = %q, want %q", Cfg.Discovery.EtcdEndpoint, "http://etcd-server:2379")
	}
	if Cfg.Discovery.RegistryRootKeyPrefix != "os-artificer/saber" {
		t.Errorf("Cfg.Discovery.RegistryRootKeyPrefix = %q, want %q", Cfg.Discovery.RegistryRootKeyPrefix, "os-artificer/saber")
	}
	if Cfg.Discovery.RegistryTTL != 60 {
		t.Errorf("Cfg.Discovery.RegistryTTL = %d, want 60", Cfg.Discovery.RegistryTTL)
	}
	if Cfg.Log.FileName != "./logs/controller.log" {
		t.Errorf("Cfg.Log.FileName = %q, want %q", Cfg.Log.FileName, "./logs/controller.log")
	}
	if Cfg.Log.LogLevel != logger.DebugLevel {
		t.Errorf("Cfg.Log.LogLevel = %q, want %q", Cfg.Log.LogLevel, logger.DebugLevel)
	}
	if Cfg.Log.FileSize != 100 {
		t.Errorf("Cfg.Log.FileSize = %d, want 100", Cfg.Log.FileSize)
	}
	if Cfg.Log.MaxBackupCount != 10 {
		t.Errorf("Cfg.Log.MaxBackupCount = %d, want 10", Cfg.Log.MaxBackupCount)
	}
	if Cfg.Log.MaxBackupAge != 10 {
		t.Errorf("Cfg.Log.MaxBackupAge = %d, want 10", Cfg.Log.MaxBackupAge)
	}
}

func TestUnmarshalConfiguration(t *testing.T) {
	yamlBytes := []byte(`
name: TestController
version: v2.0.0
discovery:
  etcdEndpoint: "http://etcd:2379"
  registryRootKeyPrefix: "test/prefix"
  registryTTL: 120
service:
  listenAddress: "tcp://0.0.0.0:26689"
log:
  fileName: "/var/log/controller.log"
  logLevel: "info"
  fileSize: 200
  maxBackupCount: 5
  maxBackupAge: 7
`)
	var c Configuration
	if err := yaml.Unmarshal(yamlBytes, &c); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if c.Name != "TestController" {
		t.Errorf("Name = %q, want TestController", c.Name)
	}
	if c.Version != "v2.0.0" {
		t.Errorf("Version = %q, want v2.0.0", c.Version)
	}
	if c.Discovery.EtcdEndpoint != "http://etcd:2379" {
		t.Errorf("Discovery.EtcdEndpoint = %q, want http://etcd:2379", c.Discovery.EtcdEndpoint)
	}
	if c.Discovery.RegistryRootKeyPrefix != "test/prefix" {
		t.Errorf("Discovery.RegistryRootKeyPrefix = %q, want test/prefix", c.Discovery.RegistryRootKeyPrefix)
	}
	if c.Discovery.RegistryTTL != 120 {
		t.Errorf("Discovery.RegistryTTL = %d, want 120", c.Discovery.RegistryTTL)
	}
	wantEp := sbnet.Endpoint{Protocol: "tcp", Host: "0.0.0.0", Port: 26689}
	if c.Service.ListenAddress != wantEp {
		t.Errorf("Service.ListenAddress = %+v, want %+v", c.Service.ListenAddress, wantEp)
	}
	if c.Log.FileName != "/var/log/controller.log" {
		t.Errorf("Log.FileName = %q, want /var/log/controller.log", c.Log.FileName)
	}
	if c.Log.LogLevel != logger.InfoLevel {
		t.Errorf("Log.LogLevel = %q, want info", c.Log.LogLevel)
	}
	if c.Log.FileSize != 200 {
		t.Errorf("Log.FileSize = %d, want 200", c.Log.FileSize)
	}
	if c.Log.MaxBackupCount != 5 {
		t.Errorf("Log.MaxBackupCount = %d, want 5", c.Log.MaxBackupCount)
	}
	if c.Log.MaxBackupAge != 7 {
		t.Errorf("Log.MaxBackupAge = %d, want 7", c.Log.MaxBackupAge)
	}
}

func TestUnmarshalConfiguration_listenAddressOnly(t *testing.T) {
	yamlBytes := []byte(`
service:
  listenAddress: "tcp://127.0.0.1:26688"
`)
	var c Configuration
	if err := yaml.Unmarshal(yamlBytes, &c); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	wantEp := sbnet.Endpoint{Protocol: "tcp", Host: "127.0.0.1", Port: 26688}
	if c.Service.ListenAddress != wantEp {
		t.Errorf("Service.ListenAddress = %+v, want %+v", c.Service.ListenAddress, wantEp)
	}
}

func TestUnmarshalConfiguration_invalidListenAddress(t *testing.T) {
	yamlBytes := []byte(`
service:
  listenAddress: "no-scheme"
`)
	var c Configuration
	err := yaml.Unmarshal(yamlBytes, &c)
	if err == nil {
		t.Error("yaml.Unmarshal expected error for invalid listenAddress, got nil")
	}
}
