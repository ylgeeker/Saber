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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"os-artificer/saber/pkg/app"

	"github.com/spf13/cobra"
)

const supervisorEnv = "SABER_AGENT_SUPERVISOR"

func pidFilePath() string {
	return filepath.Join("pids", "agent.pid")
}

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

	pf := pidFilePath()
	if err := os.MkdirAll(filepath.Dir(pf), 0755); err != nil {
		return fmt.Errorf("create pids dir: %w", err)
	}
	if err := os.WriteFile(pf, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer os.Remove(pf)

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

func runStop(cmd *cobra.Command, args []string) error {
	pf := pidFilePath()
	data, err := os.ReadFile(pf)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not running")
		}
		return fmt.Errorf("read pid file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid pid file: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM: %w", err)
	}
	return nil
}

func runRestart(cmd *cobra.Command, args []string) error {
	if err := runStop(cmd, args); err != nil && err.Error() != "agent not running" {
		return err
	}
	time.Sleep(1 * time.Second)
	return runStart(cmd, args)
}
