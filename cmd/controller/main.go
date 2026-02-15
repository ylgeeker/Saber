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

package main

import (
	"os-artificer/saber/internal/controller"
	"os-artificer/saber/pkg/logger"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "Controller",
		Short:        "Saber Controller Server",
		SilenceUsage: true,
		RunE:         controller.Run,
	}

	rootCmd.PersistentFlags().StringVarP(&controller.ConfigFilePath, "config", "c", "./etc/controller.yaml", "")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(controller.StartCmd)
	rootCmd.AddCommand(controller.StopCmd)
	rootCmd.AddCommand(controller.RestartCmd)
	rootCmd.AddCommand(controller.ReloadCmd)
	rootCmd.AddCommand(controller.HealthCheckCmd)
	rootCmd.AddCommand(controller.VersionCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("failed to start controller server. errmsg:%s", err.Error())
		return
	}

}
