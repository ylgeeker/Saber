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

package agent

import (
	"context"
	"fmt"
	"sync"

	"os-artificer/saber/internal/agent/config"
	"os-artificer/saber/internal/agent/controller"
	"os-artificer/saber/internal/agent/harvester"
	"os-artificer/saber/internal/agent/harvester/plugin"
	"os-artificer/saber/internal/agent/reporter"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"
	"os-artificer/saber/pkg/tools"
)

type Service struct {
	ctx       context.Context
	cancel    context.CancelFunc
	reporter  reporter.Reporter
	harvester *harvester.Harvester
	ctrl      *controller.ControllerClient
}

// NewService builds a service from a reporter, harvester, and optional controller client (used by CreateService).
// If ctrl is nil, no controller connection is used.
func NewService(
	ctx context.Context,
	rep reporter.Reporter,
	h *harvester.Harvester,
	ctrl *controller.ControllerClient) *Service {
	ctx, cancel := context.WithCancel(ctx)
	return &Service{ctx: ctx, cancel: cancel, reporter: rep, harvester: h, ctrl: ctrl}
}

// Run starts reporter, harvester, and optional controller client, then blocks until context is cancelled.
func (s *Service) Run() error {
	var runWg sync.WaitGroup

	if s.ctrl != nil {
		runWg.Add(1)
		tools.Go(func() {
			defer runWg.Done()
			if err := s.ctrl.Run(); err != nil && s.ctx.Err() == nil {
				logger.Warnf("controller client exited: %v", err)
			}
		})
	}

	runWg.Add(1)
	tools.Go(func() {
		defer runWg.Done()

		if err := s.reporter.Run(); err != nil {
			logger.Warnf("reporter exited: %v", err)
		}
	})

	runWg.Add(1)
	tools.Go(func() {
		defer runWg.Done()

		if err := s.harvester.Run(s.ctx); err != nil && s.ctx.Err() == nil {
			logger.Warnf("harvester exited: %v", err)
		}
	})

	<-s.ctx.Done()

	if s.ctrl != nil {
		if err := s.ctrl.Close(); err != nil {
			logger.Warnf("controller client close: %v", err)
		}
	}

	if err := s.reporter.Close(); err != nil {
		logger.Warnf("reporter close: %v", err)
	}

	if err := s.harvester.Close(); err != nil {
		logger.Warnf("harvester close: %v", err)
	}

	runWg.Wait()
	return s.ctx.Err()
}

// Close cancels the service context so Run returns.
func (s *Service) Close() error {
	if s.ctrl != nil {
		if err := s.ctrl.Close(); err != nil {
			logger.Warnf("controller client close: %v", err)
		}
		s.ctrl = nil
	}

	if s.harvester != nil {
		if err := s.harvester.Close(); err != nil {
			logger.Warnf("harvester close: %v", err)
		}
		s.harvester = nil
	}

	if s.reporter != nil {
		if err := s.reporter.Close(); err != nil {
			logger.Warnf("reporter close: %v", err)
		}
		s.reporter = nil
	}

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	return nil
}

// CreateService creates a Service by loading reporter and harvester from config (dynamic proxy).
func CreateService(ctx context.Context, cfg *config.Configuration) (*Service, error) {
	if len(cfg.Reporters) == 0 {
		return nil, fmt.Errorf("no reporters configured")
	}

	entry := cfg.Reporters[0]
	opts := &config.ReporterOpts{
		Config:       entry.Config,
		AgentName:    cfg.Name,
		AgentVersion: cfg.Version,
	}

	rep, err := reporter.CreateReporter(ctx, entry.Type, opts)
	if err != nil {
		return nil, err
	}

	pluginConfigs := make([]plugin.PluginConfig, 0, len(cfg.Harvester.Plugins))
	for _, e := range cfg.Harvester.Plugins {
		pluginConfigs = append(pluginConfigs, plugin.PluginConfig{Name: e.Name, Options: e.Options})
	}

	plugins, err := plugin.CreatePlugins(ctx, pluginConfigs)
	if err != nil {
		_ = rep.Close()
		return nil, err
	}

	h := harvester.NewHarvester(rep, plugins)

	var ctrl *controller.ControllerClient
	if cfg.Controller.Endpoints != "" {
		clientID, err := tools.MachineID("saber-agent")
		if err != nil {
			_ = rep.Close()
			return nil, fmt.Errorf("controller client requires machine-id: %w", err)
		}
		ctrl = controller.NewControllerClient(ctx, cfg.Controller.Endpoints, clientID)
		ctrl.OnResponse(func(resp *proto.AgentResponse) {
			if resp == nil {
				return
			}
			if resp.Code != 0 {
				logger.Warnf("controller response error: code=%d errmsg=%s", resp.Code, resp.GetErrmsg())
				return
			}
			if len(resp.Payload) > 0 {
				logger.Debugf("controller response: payload len=%d", len(resp.Payload))
			}
		})
	}

	return NewService(ctx, rep, h, ctrl), nil
}
