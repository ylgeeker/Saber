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
	"os-artificer/saber/internal/transfer"
	"os-artificer/saber/pkg/logger"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "Transfer",
		Short:        "Saber Transfer Server",
		SilenceUsage: true,
		RunE:         transfer.Run,
	}

	rootCmd.PersistentFlags().StringVarP(&transfer.ConfigFilePath, "config", "c", "./etc/transfer.yaml", "")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(transfer.StartCmd)
	rootCmd.AddCommand(transfer.StopCmd)
	rootCmd.AddCommand(transfer.RestartCmd)
	rootCmd.AddCommand(transfer.ReloadCmd)
	rootCmd.AddCommand(transfer.HealthCheckCmd)
	rootCmd.AddCommand(transfer.VersionCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("failed to start transfer server. errmsg:%s", err.Error())
		return
	}

}
