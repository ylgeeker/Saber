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

package discovery

import (
	"time"

	"os-artificer/saber/pkg/gerrors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	defaultChannelBuffMaxSize    = 1024
	defaultTTL                   = 6
	defaultMaxUnaryRetries       = 3
	defaultDialTimeout           = 5 * time.Second
	defaultAutoSyncInterval      = 60 * time.Second
	defaultKeepAliveTime         = 30 * time.Second
	defaultKeepAliveTimeout      = 10 * time.Second
	defaultRegistryRootKeyPrefix = "/saber/registry"
)

type Option interface {
	apply(*options) error
}

var defaultOptions = options{
	bufferMaxSize:         defaultChannelBuffMaxSize,
	ttl:                   defaultTTL,
	dialTimeout:           defaultDialTimeout,
	autoSyncInterval:      defaultAutoSyncInterval,
	keepAliveTime:         defaultKeepAliveTime,
	keepAliveTimeout:      defaultKeepAliveTimeout,
	registryRootKeyPrefix: defaultRegistryRootKeyPrefix,
	maxUnaryRetries:       defaultMaxUnaryRetries,
}

type options struct {
	user                  string
	password              string
	bufferMaxSize         int
	ttl                   int
	serviceID             string
	serviceName           string
	endpoints             []string
	dialTimeout           time.Duration
	autoSyncInterval      time.Duration
	keepAliveTime         time.Duration
	keepAliveTimeout      time.Duration
	registryRootKeyPrefix string
	maxUnaryRetries       uint
	Logger                *zap.Logger
}

func (o options) Config() clientv3.Config {
	return clientv3.Config{
		Username:  o.user,
		Password:  o.password,
		Endpoints: o.endpoints,

		// DialTimeout is the timeout for failing to establish a connection.
		DialTimeout: o.dialTimeout,

		// AutoSyncInterval is the interval to update endpoints with its latest members.
		// 0 disables auto-sync. By default auto-sync is disabled.
		AutoSyncInterval: o.autoSyncInterval,

		// DialKeepAliveTime is the time after which client pings the server to see if
		// transport is alive.
		DialKeepAliveTime: o.keepAliveTime,

		// DialKeepAliveTimeout is the time that the client waits for a response for the
		// keep-alive probe. If the response is not received in this time, the connection is closed.
		DialKeepAliveTimeout: o.keepAliveTimeout,

		// MaxUnaryRetries is the maximum number of retries for unary RPCs.
		MaxUnaryRetries: 5,

		// Logger export the gRPC log into the custom.
		Logger: o.Logger,
	}
}

type funcOptions struct {
	f func(opt *options) error
}

func (fdo *funcOptions) apply(opt *options) error {
	return fdo.f(opt)
}

func OptionUser(val string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.user = val
			return nil
		},
	}
}

func OptionPassword(val string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.password = val
			return nil
		},
	}
}

func OptionBufferMaxSize(val int) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.bufferMaxSize = val
			return nil
		},
	}
}

func OptionTTL(val int) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			if val < defaultTTL {
				opt.ttl = defaultTTL
				return nil
			}

			opt.ttl = val
			return nil
		},
	}
}

func OptionServiceID(val string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.serviceID = val
			if opt.serviceID == "" {
				return gerrors.New(gerrors.InvalidParameter, "service-id is required")
			}

			return nil
		},
	}
}

func OptionServiceName(val string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.serviceName = val
			return nil
		},
	}
}

func OptionEndpoints(epoints []string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.endpoints = append(opt.endpoints, epoints...)
			return nil
		},
	}
}

func OptionDialTimeout(val time.Duration) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.dialTimeout = val
			return nil
		},
	}
}

func OptionAutoSyncInterval(val time.Duration) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.autoSyncInterval = val
			return nil
		},
	}
}

func OptionKeepAliveTime(val time.Duration) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.keepAliveTime = val
			return nil
		},
	}
}

func OptionKeepAliveTimeout(val time.Duration) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.keepAliveTimeout = val
			return nil
		},
	}
}

func OptionRegistryRootKeyPrefix(val string) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.registryRootKeyPrefix = val
			return nil
		},
	}
}

func OptionMaxUnaryRetries(val uint) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.maxUnaryRetries = val
			return nil
		},
	}
}

func OptionLogger(val *zap.Logger) *funcOptions {
	return &funcOptions{
		f: func(opt *options) error {
			opt.Logger = val
			return nil
		},
	}
}
