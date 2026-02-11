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
	"fmt"
	"strings"
	"sync"
	"time"

	"os-artificer/saber/pkg/gerrors"
	"os-artificer/saber/pkg/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry  service registry
type Registry struct {
	serviceID        string
	rootKey          string
	ttl              int64
	wg               sync.WaitGroup
	cliMu            sync.RWMutex
	client           *clientv3.Client
	leaseID          clientv3.LeaseID
	keepAliveCancel  context.CancelFunc
	createEtcdClient func() (*clientv3.Client, error)
}

func (r *Registry) grant(ctx context.Context) error {
	r.cliMu.Lock()
	defer r.cliMu.Unlock()

	cli, err := r.createEtcdClient()
	if err != nil {
		return err
	}

	r.client = cli

	leaseResp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return gerrors.NewE(gerrors.ComponentFailure, err)
	}
	r.leaseID = leaseResp.ID

	logger.Debugf("registry start keepalive, leaseID: %d", r.leaseID)

	_, err = r.client.Put(ctx, r.rootKey, "", clientv3.WithLease(r.leaseID))
	if err != nil {
		r.client.Close()
		return gerrors.NewE(gerrors.ComponentFailure, err)
	}

	return r.createKeepAlive()
}

func (r *Registry) handleKeepalive(keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for respd := range keepAliveChan {
		if respd == nil {
			logger.Warnf("registry keepalive response failure.")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := r.recoverConnection(ctx); err != nil {
				logger.Warnf("registry keepalive response failure, failed to recover, errmsg: %s", err)
				return
			}

			logger.Warnf("registry keepalive response failure, recovered")
			return
		}
	}
}

// SetService Create or set the registry root key.
func (r *Registry) SetService(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)

	if r.isInvalidClient() {
		logger.Warnf("registry set service trigger to recover.")
		if err := r.recoverConnection(ctx); err != nil {
			return err
		}
	}

	r.cliMu.RLock()
	defer r.cliMu.RUnlock()

	_, err := r.client.Put(ctx, r.rootKey, value, clientv3.WithLease(r.leaseID))
	if err != nil {
		logger.Warnf("registry set service put failed, lease-id: %v errmsg: %v", r.leaseID, err)
		return gerrors.Newf(gerrors.ComponentFailure,
			"registry set service put failed, lease-id: %d, errmsg: %s", r.leaseID, err)
	}

	return nil
}

// Set Create or set the child node with this registry root key.
// If the key has the root key prefix, the key will be applied
// and then it will be appended to the registry root key.
func (r *Registry) Set(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return gerrors.New(gerrors.InvalidParameter, "key is required")
	}

	if !strings.HasPrefix(key, r.rootKey) {
		key = fmt.Sprintf("%s/%s", r.rootKey, key)
	}

	if r.isInvalidClient() {
		logger.Debugf("registry set trigger to recover.")
		if err := r.recoverConnection(ctx); err != nil {
			return err
		}
	}

	r.cliMu.RLock()
	defer r.cliMu.RUnlock()

	_, err := r.client.Put(ctx, key, value)
	if err != nil {
		return gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	return nil
}

// Close Registry instance
func (r *Registry) Close() {
	r.cliMu.Lock()
	defer r.cliMu.Unlock()

	if r.keepAliveCancel != nil {
		r.keepAliveCancel()
	}

	if r.leaseID != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.client.Revoke(ctx, r.leaseID)
	}

	r.wg.Wait()
	logger.Debugf("registry closed")
}

// GetRootKey returns the root key of the registry
func (r *Registry) GetRootKey() string {
	return r.rootKey
}

func (r *Registry) createKeepAlive() error {
	// NOTE: keepAlive must use the context without timeout.
	keepAliveCtx, cancel := context.WithCancel(context.Background())
	r.keepAliveCancel = cancel
	keepAliveResp, err := r.client.KeepAlive(keepAliveCtx, r.leaseID)
	if err != nil {
		r.client.Close()
		return gerrors.NewE(gerrors.ComponentFailure, err)
	}

	logger.Debugf("registry start keepalive monitor, leaseID: %d", r.leaseID)

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.handleKeepalive(keepAliveResp)
	}()

	return nil
}

func (r *Registry) isInvalidClient() bool {
	r.cliMu.RLock()
	defer r.cliMu.RUnlock()
	return r.client == nil || r.leaseID == 0
}

func (r *Registry) renewalLease(ctx context.Context) error {
	r.cliMu.Lock()
	defer r.cliMu.Unlock()

	resp, err := r.client.TimeToLive(ctx, r.leaseID)
	if err != nil {
		logger.Warnf("failed to retrieve the lease, need to grant a new lease, errmsg: %s", err)
		return gerrors.New(gerrors.ComponentFailure, "failed to retrieve the lease")
	}

	if resp.TTL <= 0 {
		logger.Warnf("registry lease expired, need to grant a new lease, TTL:%d, errmsg: %s", resp.TTL, err)
		return gerrors.New(gerrors.ComponentFailure, "need to grant a new lease")
	}

	return r.createKeepAlive()
}

func (r *Registry) recoverConnection(ctx context.Context) error {
	if r.isInvalidClient() {
		logger.Warnf("registry invalid connection, need to grant a new lease.")
		return r.grant(ctx)
	}

	// renewal lease
	if err := r.renewalLease(ctx); err != nil {
		logger.Warnf("failed to renewal lease, errmsg: %s", err)
		// create new lease
		return r.grant(ctx)
	}

	return nil
}
