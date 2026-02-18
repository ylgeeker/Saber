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

package controller

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"os-artificer/saber/internal/controller/config"
	"os-artificer/saber/pkg/logger"

	"github.com/spf13/cobra"
)

func setupGracefulShutdown(svr *Service) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigC
		svr.Close()
		os.Exit(0)
	}()
}

func Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	loadControllerConfig()

	svr := CreateService(ctx, config.Cfg.Service.ListenAddress, "")

	setupGracefulShutdown(svr)

	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)

	go func() {
		for range reloadCh {
			if err := svr.ReloadConfig(); err != nil {
				logger.Warnf("reload config: %v", err)
			}
		}
	}()

	return svr.Run()
}
