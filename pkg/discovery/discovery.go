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

	"os-artificer/saber/pkg/gerrors"
	"os-artificer/saber/pkg/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatchedEventType int

const (
	WatchedEventPut WatchedEventType = iota
	WatchedEventDelete
	WatchedEventUnknown
)

var (
	ErrEmptyWatchedKey = gerrors.New(gerrors.InvalidParameter, "watched key is required but got empty string")
)

// WatchEvent This event will be generated
// when the watch method detects that event has occured.
type WatchEvent struct {
	EventType WatchedEventType
	Key       string
	Value     []byte
}

// Discovery service discovery
type Discovery struct {
	quit             chan struct{}
	cliMu            sync.RWMutex
	client           *clientv3.Client
	createEtcdClient func() (*clientv3.Client, error)
	wg               sync.WaitGroup
}

// Watch Subscribe to target key events and receive data from the watch channel.
func (d *Discovery) Watch(ctx context.Context, key string) (<-chan *WatchEvent, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrEmptyWatchedKey
	}

	return d.watchCommon(ctx, key)
}

// WatchWithPrefix Subscribe to prefix key events with prefix and receive data from the watch channel.
func (d *Discovery) WatchWithPrefix(ctx context.Context, key string) (<-chan *WatchEvent, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrEmptyWatchedKey
	}

	return d.watchCommon(ctx, key, clientv3.WithPrefix())
}

// Get Only get the value of the key.
func (d *Discovery) Get(ctx context.Context, key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	d.cliMu.RLock()
	defer d.cliMu.RUnlock()

	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	if len(resp.Kvs) == 0 {
		errmsg := fmt.Sprintf("the key not exists, key:%s", key)
		return nil, gerrors.New(gerrors.NotFound, errmsg)
	}

	return resp.Kvs[0].Value, nil
}

// GetWithPrefix Get all values that start with the specified key prefix.
func (d *Discovery) GetWithPrefix(ctx context.Context, key string) (map[string][]byte, error) {
	key = strings.TrimSpace(key)
	d.cliMu.RLock()
	defer d.cliMu.RUnlock()

	resp, err := d.client.Get(ctx, key, clientv3.WithPrefix())

	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	kvs := map[string][]byte{}
	for _, kv := range resp.Kvs {
		kvs[string(kv.Key)] = []byte(kv.Value)
	}

	return kvs, nil
}

// Close Discovery instance
func (d *Discovery) Close() {
	if d.quit != nil {
		close(d.quit) // NOTE: Notify all goroutines by closing this channel.
	}

	d.wg.Wait()
	d.quit = nil

	d.cliMu.Lock()
	defer d.cliMu.Unlock()

	if d.client != nil {
		d.client.Close()
		d.client = nil
	}
}

func (d *Discovery) watchCommon(ctx context.Context, key string, opts ...clientv3.OpOption) (<-chan *WatchEvent, error) {
	if d.quit == nil {
		d.quit = make(chan struct{})
	}

	d.cliMu.Lock()
	if d.client == nil {
		etcdCli, err := d.createEtcdClient()
		if err != nil {
			d.cliMu.Unlock()
			return nil, err
		}
		d.client = etcdCli
	}

	// set watcher
	watchChan := d.client.Watch(ctx, key, opts...)
	watchEventChan := make(chan *WatchEvent, defaultChannelBuffMaxSize)

	d.cliMu.Unlock()

	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		defer close(watchEventChan)

		for {
			select {
			case <-d.quit:
				logger.Infof("exit watcher. key prefix:%s", key)
				return

			case <-ctx.Done():
				logger.Infof("exit watcher. key prefix:%s", key)
				return

			case watchResp := <-watchChan:
				if err := watchResp.Err(); err != nil {
					d.cliMu.Lock()
					d.client.Close()
					d.client = nil
					d.cliMu.Unlock()
					logger.Errorf("failed to read watch event, errmsg: %v", err)
					return
				}

				for _, event := range watchResp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						event := &WatchEvent{
							EventType: WatchedEventPut,
							Key:       string(event.Kv.Key),
							Value:     []byte(event.Kv.Value),
						}
						watchEventChan <- event

					case clientv3.EventTypeDelete:
						event := &WatchEvent{
							EventType: WatchedEventDelete,
							Key:       string(event.Kv.Key),
							Value:     []byte(event.Kv.Value),
						}
						watchEventChan <- event

					}
				}
			}
		}
	}()

	return watchEventChan, nil
}
