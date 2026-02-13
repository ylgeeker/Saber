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

package app

import (
	"context"
	"time"

	"os-artificer/saber/pkg/sbproc"
)

// Daemon starts a child process with the given parameters and restarts it
// automatically if it exits unexpectedly. Cancel the context passed to NewDaemon
// (or call Stop()) to stop the daemon and the child process.
type Daemon struct {
	env    map[string]string
	cmd    string
	args   []string
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDaemon returns a Daemon that will run the given command with args and optional env.
// env can be nil to use the current process environment. The daemon uses a child of ctx;
// cancel that context or call Stop() to stop the daemon and the child.
func NewDaemon(ctx context.Context, env map[string]string, cmd string, args ...string) *Daemon {
	ctx, cancel := context.WithCancel(ctx)
	return &Daemon{env: env, cmd: cmd, args: args, ctx: ctx, cancel: cancel}
}

// Start starts the child process and a goroutine that restarts it if it exits
// unexpectedly. Start returns nil once the first child has been started;
// call Stop or cancel the daemon's context to stop the child and the monitor.
func (d *Daemon) Start() error {
	child, err := d.startChild()
	if err != nil {
		return err
	}

	go d.monitor(child)
	return nil
}

func (d *Daemon) startChild() (*sbproc.Child, error) {
	return sbproc.StartWithEnv(d.ctx, d.env, d.cmd, d.args...)
}

func (d *Daemon) monitor(child *sbproc.Child) {
	for {
		_, _ = child.Wait()
		if d.ctx.Err() != nil {
			return
		}

		time.Sleep(time.Second)
		if d.ctx.Err() != nil {
			return
		}

		var err error
		child, err = d.startChild()
		if err != nil {
			return
		}
	}
}

// Stop stops the daemon and the child process. It is safe to call multiple times.
func (d *Daemon) Stop() {
	d.cancel()
}
