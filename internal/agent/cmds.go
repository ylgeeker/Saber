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
	"os-artificer/saber/pkg/version"

	"github.com/spf13/cobra"
)

// VersionCmd prints the version information of the agent
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Version Information",
	Run: func(cmd *cobra.Command, args []string) {
		version.Print("saber Agent")
	},
}

// RestartCmd restarts the agent
var RestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Agent",
	RunE:  runRestart,
}

// ReloadCmd reloads the agent
var ReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload Agent",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// StopCmd stops the agent
var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Agent",
	RunE:  runStop,
}

// StartCmd starts the agent
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Agent",
	RunE:  runStart,
}

// HealthCheckCmd checks the health of the agent
var HealthCheckCmd = &cobra.Command{
	Use:   "health",
	Short: "Health Check Agent",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
