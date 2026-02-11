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
	"os-artificer/saber/pkg/gerrors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	etcdKeySegmentSelf     = "self"
	etcdKeySegmentMutex    = "mutex"
	etcdKeySegmentElection = "election/leader"
)

// Client etcd client
type Client struct {
	opts             options
	createEtcdClient func() (*clientv3.Client, error)
}

// NewClientWithOptions create etcd client with option
func NewClientWithOptions(opts ...Option) (*Client, error) {
	cli := &Client{
		opts: defaultOptions,
	}

	for _, opt := range opts {
		if err := opt.apply(&cli.opts); err != nil {
			return nil, err
		}
	}

	if cli.opts.serviceName != "" {
		cli.opts.registryRootKeyPrefix += "/" + cli.opts.serviceName
	}

	cli.createEtcdClient = func() (*clientv3.Client, error) {
		etcdCli, err := clientv3.New(cli.opts.Config())
		if err != nil {
			return nil, gerrors.Newf(gerrors.ComponentFailure, "%v", err)
		}

		return etcdCli, nil
	}

	return cli, nil
}

// OriginClient return the origin etcd client
func (c Client) OriginClient() (*clientv3.Client, error) {
	return c.createEtcdClient()
}

// GetRegistryPrefix returns the etcd key prefix under which same-module instances register.
// Use with Discovery.GetWithPrefix / WatchWithPrefix to list or watch analysis instances.
func (c Client) GetRegistryPrefix() string {
	return c.opts.registryRootKeyPrefix
}

// GetSelfPrefix returns the etcd key prefix under which same-module instances register (self nodes).
// Full key for one instance is GetSelfPrefix() + "/" + serviceID.
// Use with GetWithPrefix to list all instances.
func (c Client) GetSelfPrefix() string {
	return c.opts.registryRootKeyPrefix + "/" + etcdKeySegmentSelf
}

// GetElectionPrefix returns the etcd key prefix for leader election.
//
//	Full key for one election is GetElectionPrefix() + "/" + name.
func (c Client) GetElectionPrefix() string {
	return c.opts.registryRootKeyPrefix + "/" + etcdKeySegmentElection
}

// CreateRegistry create new etcd registry
func (c Client) CreateRegistry() *Registry {
	rootKey := c.opts.registryRootKeyPrefix
	if c.opts.serviceID != "" {
		rootKey += "/" + etcdKeySegmentSelf + "/" + c.opts.serviceID
	}

	registry := &Registry{
		serviceID:        c.opts.serviceID,
		rootKey:          rootKey,
		ttl:              defaultTTL,
		createEtcdClient: c.createEtcdClient,
	}

	return registry
}

// CreateDiscovery create etcd discovery
func (c Client) CreateDiscovery() (*Discovery, error) {
	discovery := &Discovery{
		quit:             make(chan struct{}),
		createEtcdClient: c.createEtcdClient,
	}

	return discovery, nil
}

// CreateMutex returns concurrency mutex.
func (c Client) CreateMutex(key string) (ConcurrencyMutex, error) {
	etcdCli, err := c.createEtcdClient()
	if err != nil {
		return nil, err
	}

	session, err := concurrency.NewSession(etcdCli)
	if err != nil {
		return nil, gerrors.NewE(gerrors.ComponentFailure, err)
	}

	muKey := c.opts.registryRootKeyPrefix + "/" + etcdKeySegmentMutex + "/" + key
	mu := &concurrencyMutex{
		etcdCli: etcdCli,
		session: session,
		key:     muKey,
		mutex:   concurrency.NewMutex(session, muKey),
	}

	return mu, nil
}

// CreateElection returns concurrency election.
func (c Client) CreateElection(name string) (ConcurrencyElection, error) {
	etcdCli, err := c.createEtcdClient()
	if err != nil {
		return nil, err
	}

	session, err := concurrency.NewSession(etcdCli)
	if err != nil {
		return nil, gerrors.NewE(gerrors.ComponentFailure, err)
	}

	electionKey := c.opts.registryRootKeyPrefix + "/" + etcdKeySegmentElection + "/" + name

	election := &concurrencyElection{
		etcdCli:  etcdCli,
		session:  session,
		election: concurrency.NewElection(session, electionKey),
		key:      c.opts.serviceID,
	}

	return election, nil
}
