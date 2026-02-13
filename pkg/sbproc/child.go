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

package sbproc

import (
	"io"
	"os/exec"
	"syscall"
)

const streamChanBuf = 64
const streamReadBufSize = 4096

// Child represents a child process started by this package. Call Wait() to wait for
// it to exit and avoid zombie processes.
type Child struct {
	cmd      *exec.Cmd
	stdoutCh <-chan []byte
	stderrCh <-chan []byte
}

// PID returns the child process ID. Valid after Start() has been called.
func (c *Child) PID() int {
	if c == nil || c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

// Wait waits for the child process to exit and returns its exit code (0 on success)
// and any error (e.g. from context cancel or failure to wait).
func (c *Child) Wait() (int, error) {
	if c == nil || c.cmd == nil {
		return -1, exec.ErrNotFound
	}

	err := c.cmd.Wait()
	if err == nil {
		return 0, nil
	}

	if ee, ok := err.(*exec.ExitError); ok {
		if status, ok := ee.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
		return 1, nil
	}

	return -1, err
}

// Stdout returns a receive-only channel of stdout data. Only non-nil when the child
// was started with StartWithStreams; otherwise returns nil (do not range over nil).
func (c *Child) Stdout() <-chan []byte {
	if c == nil {
		return nil
	}
	return c.stdoutCh
}

// Stderr returns a receive-only channel of stderr data. Only non-nil when the child
// was started with StartWithStreams; otherwise returns nil.
func (c *Child) Stderr() <-chan []byte {
	if c == nil {
		return nil
	}
	return c.stderrCh
}

// streamPipes reads from stdoutPipe and stderrPipe in two goroutines, sends data
// as []byte chunks on the returned channels, and closes each channel on EOF or error.
func streamPipes(stdoutPipe, stderrPipe io.Reader) (stdoutCh, stderrCh <-chan []byte) {
	outCh := make(chan []byte, streamChanBuf)
	errCh := make(chan []byte, streamChanBuf)

	go func() {
		defer close(outCh)
		buf := make([]byte, streamReadBufSize)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				cp := make([]byte, n)
				copy(cp, buf[:n])
				outCh <- cp
			}

			if err != nil {
				return
			}
		}
	}()

	go func() {
		defer close(errCh)
		buf := make([]byte, streamReadBufSize)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				cp := make([]byte, n)
				copy(cp, buf[:n])
				errCh <- cp
			}

			if err != nil {
				return
			}
		}
	}()

	return outCh, errCh
}

// newChildWithStreams builds a Child that has stdout/stderr channels fed by streamPipes.
func newChildWithStreams(cmd *exec.Cmd, stdoutPipe, stderrPipe io.Reader) *Child {
	stdoutCh, stderrCh := streamPipes(stdoutPipe, stderrPipe)
	return &Child{cmd: cmd, stdoutCh: stdoutCh, stderrCh: stderrCh}
}
