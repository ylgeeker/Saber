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
	"time"

	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/sbnet"
)

var Cfg = Configuration{
	Name:    "Controller",
	Version: "v1.0.0",

	Discovery: DiscoveryConfig{
		EtcdEndpoint:          "http://etcd-server:2379",
		EtcdUser:              "root",
		EtcdPassword:          "wktest",
		DialTimeout:           5 * time.Second,
		AutoSyncInterval:      10 * time.Second,
		DialKeepAliveTime:     5 * time.Second,
		DialKeepAliveTimeout:  5 * time.Second,
		RegistryRootKeyPrefix: "os-artificer/saber",
		RegistryTTL:           60,
	},

	Service: ServiceConfig{
		ListenAddress: sbnet.Endpoint{Protocol: "tcp", Host: "127.0.0.1", Port: 26689},
	},
	Log: LogConfig{
		FileName:       "./logs/controller.log",
		LogLevel:       logger.DebugLevel,
		FileSize:       100,
		MaxBackupCount: 10,
		MaxBackupAge:   10,
	},
}

// DiscoveryConfig discovery's config
type DiscoveryConfig struct {
	EtcdEndpoint          string        `yaml:"etcdEndpoint"`
	EtcdUser              string        `yaml:"etcdUser"`
	EtcdPassword          string        `yaml:"etcdPassword"`
	DialTimeout           time.Duration `yaml:"dialTimeout"`
	AutoSyncInterval      time.Duration `yaml:"autoSyncInterval"`
	DialKeepAliveTime     time.Duration `yaml:"dialKeepAliveTime"`
	DialKeepAliveTimeout  time.Duration `yaml:"dialKeepAliveTimeout"`
	RegistryRootKeyPrefix string        `yaml:"registryRootKeyPrefix"`
	RegistryTTL           int64         `yaml:"registryTTL"` // in seconds
}

// ServiceConfig service local config
type ServiceConfig struct {
	ListenAddress sbnet.Endpoint `yaml:"listenAddress"`
}

// LogConfig log config
type LogConfig struct {
	FileName       string       `yaml:"fileName"`
	LogLevel       logger.Level `yaml:"logLevel"`
	FileSize       int          `yaml:"fileSize"`
	MaxBackupCount int          `yaml:"maxBackupCount"`
	MaxBackupAge   int          `yaml:"maxBackupAge"`
}

// Configuration controller's configuration
type Configuration struct {
	Name      string          `yaml:"name"`
	Version   string          `yaml:"version"`
	Discovery DiscoveryConfig `yaml:"discovery"`
	Service   ServiceConfig   `yaml:"service"`
	Log       LogConfig       `yaml:"log"`
}
