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

package client

import (
	"os-artificer/saber/pkg/version"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Version Information",
	Run: func(cmd *cobra.Command, args []string) {
		version.Print("saber Client")
	},
}

var HealthCheckCmd = &cobra.Command{
	Use:   "health",
	Short: "Health Check Client Server",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var (
	listHost bool
	listNum  int
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources (e.g. hosts)",
	RunE:  runList,
}

func init() {
	// -h is reserved by Cobra for help; use -H/--host for "query host information".
	ListCmd.Flags().BoolVarP(&listHost, "host", "H", false, "query host information")
	ListCmd.Flags().IntVarP(&listNum, "num", "n", 10, "number of items to display")
}

func runList(cmd *cobra.Command, args []string) error {
	// -h: 查询主机信息
	// -n: 显示数量
	_ = listHost
	_ = listNum
	return nil
}
