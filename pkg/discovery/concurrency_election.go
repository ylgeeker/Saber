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
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type ConcurrencyElection interface {
	// Campaign During the election process,
	// the Campagin will remain blocked until it's elected as the leader.
	Campaign(ctx context.Context) error
	Close()
	Done() <-chan struct{}
}

type concurrencyElection struct {
	etcdCli  *clientv3.Client
	session  *concurrency.Session
	election *concurrency.Election
	key      string
}

func (ce *concurrencyElection) Campaign(ctx context.Context) error {
	return ce.election.Campaign(ctx, ce.key)
}

func (ce *concurrencyElection) Close() {
	ce.session.Close()
	ce.etcdCli.Close()
}

func (ce *concurrencyElection) Done() <-chan struct{} {
	return ce.session.Done()
}
