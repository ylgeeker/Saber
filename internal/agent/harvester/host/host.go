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

package host

import (
	"context"
	"sync"
	"time"

	"os-artificer/saber/internal/agent/harvester/plugin"
	"os-artificer/saber/pkg/logger"
)

const hostPluginVersion = "1.0.0"

func init() {
	plugin.RegisterPlugin("host", newHostPlugin)
}

// HostPlugin collects host metrics/info.
type HostPlugin struct {
	plugin.UnimplementedPlugin

	stats *Stats
	wg    sync.WaitGroup
	done  chan struct{}
	opts  any
}

func newHostPlugin(ctx context.Context, opts any) (plugin.Plugin, error) {
	return &HostPlugin{
		stats: NewStats(),
		opts:  opts,
		done:  make(chan struct{}),
	}, nil
}

func (p *HostPlugin) Version() string {
	return hostPluginVersion
}

func (p *HostPlugin) Name() string {
	return "host"
}

func (p *HostPlugin) Run(ctx context.Context) (plugin.EventC, error) {
	eventC := make(plugin.EventC)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer close(eventC)

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-p.done:
				logger.Infof("host plugin run exited: %s", p.Name())
				return

			case <-ctx.Done():
				logger.Infof("host plugin run exited: %s", p.Name())
				return

			case <-timer.C:
				if err := p.stats.CollectStats(); err != nil {
					logger.Errorf("failed to collect host stats: %v", err)
					timer.Reset(5 * time.Second)
					continue
				}

				eventC <- &plugin.Event{
					PluginName: p.Name(),
					EventName:  "host",
					Data:       p.stats,
				}

				timer.Reset(5 * time.Second)
			}
		}
	}()

	return eventC, nil
}

func (p *HostPlugin) Close() error {
	if p.done != nil {
		close(p.done)
		p.done = nil
	}

	p.wg.Wait()
	return nil
}
