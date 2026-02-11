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

	"os-artificer/saber/pkg/gerrors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type ConcurrencyMutex interface {
	TryLock(ctx context.Context) error
	Unlock(ctx context.Context) error
	Close()
}

type concurrencyMutex struct {
	etcdCli *clientv3.Client
	session *concurrency.Session
	mutex   *concurrency.Mutex
	key     string
}

func (c *concurrencyMutex) TryLock(ctx context.Context) error {
	if err := c.mutex.TryLock(context.Background()); err != nil {
		return gerrors.Newf(gerrors.Failure, "%v", err)
	}
	return nil
}

func (c *concurrencyMutex) Unlock(ctx context.Context) error {
	if err := c.mutex.Unlock(context.Background()); err != nil {
		return gerrors.Newf(gerrors.Failure, "%v", err)
	}
	return nil
}

func (c *concurrencyMutex) Close() {
	c.session.Close()
	c.etcdCli.Close()
}
