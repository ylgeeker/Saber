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
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"os-artificer/saber/pkg/app"

	"github.com/spf13/cobra"
)

const supervisorEnv = "SABER_AGENT_SUPERVISOR"

func runStart(cmd *cobra.Command, args []string) error {
	if os.Getenv(supervisorEnv) == "1" {
		return RunSupervisor()
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	var childArgs []string
	if ConfigFilePath != "" {
		childArgs = []string{"-c", ConfigFilePath}
	}

	runCmd := exec.Command(executable, childArgs...)
	runCmd.Env = append(os.Environ(), supervisorEnv+"=1")
	runCmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("start supervisor: %w", err)
	}

	return nil
}

// RunSupervisor runs the daemon loop in the current process (worker executable and
// args from os.Executable() and os.Args[1:]). Called when SABER_AGENT_SUPERVISOR=1.
func RunSupervisor() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	childArgs := os.Args[1:]

	workerEnv := map[string]string{supervisorEnv: ""}
	daemon := app.NewDaemon(ctx, workerEnv, executable, childArgs...)
	if err := daemon.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	<-ctx.Done()
	daemon.Stop()
	return nil
}
