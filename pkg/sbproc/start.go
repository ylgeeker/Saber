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
	"context"
	"io"
	"maps"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

// Start starts the given command with args, ignoring stdin, stdout, and stderr.
// Returns a Child handle so the caller can get PID() and call Wait().
func Start(ctx context.Context, cmd string, args ...string) (*Child, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	closers, err := startWithIO(c)
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		for _, cl := range closers {
			_ = cl.Close()
		}
		return nil, err
	}

	for _, cl := range closers {
		_ = cl.Close()
	}
	return &Child{cmd: c}, nil
}

// StartWithEnv starts the given command with args and the provided environment
// (merged over os.Environ()). Stdin, stdout, and stderr are ignored.
func StartWithEnv(ctx context.Context, env map[string]string, cmd string, args ...string) (*Child, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = envSlice(env)
	closers, err := startWithIO(c)
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		for _, cl := range closers {
			_ = cl.Close()
		}
		return nil, err
	}

	for _, cl := range closers {
		_ = cl.Close()
	}
	return &Child{cmd: c}, nil
}

// StartWithOutput starts the command, waits for it to finish, and returns the child
// handle (for PID()), combined stdout and stderr bytes, and any error from Wait().
func StartWithOutput(ctx context.Context, cmd string, args ...string) (child *Child, stdout []byte, stderr []byte, err error) {
	c := exec.CommandContext(ctx, cmd, args...)
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	if err = c.Start(); err != nil {
		return nil, nil, nil, err
	}

	child = &Child{cmd: c}
	var outB, errB []byte
	var outErr, errErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		outB, outErr = io.ReadAll(stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		errB, errErr = io.ReadAll(stderrPipe)
	}()
	wg.Wait()

	code, waitErr := child.Wait()
	if waitErr != nil {
		return child, outB, errB, waitErr
	}
	if code != 0 {
		return child, outB, errB, &exitError{code: code}
	}
	if outErr != nil {
		return child, outB, errB, outErr
	}
	if errErr != nil {
		return child, outB, errB, errErr
	}
	return child, outB, errB, nil
}

// StartWithStreams starts the command and returns a Child whose stdout and stderr are
// available via child.Stdout() and child.Stderr() as receive-only channels. Data is
// sent as the process writes; each channel is closed when the stream reaches EOF.
// For long-running processes, consume from the channels (e.g. in goroutines) and call
// child.Wait() when done to reap the process.
func StartWithStreams(ctx context.Context, cmd string, args ...string) (*Child, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = c.Start(); err != nil {
		return nil, err
	}

	return newChildWithStreams(c, stdoutPipe, stderrPipe), nil
}

// startWithIO configures cmd to ignore stdin/stdout/stderr by wiring them to DevNull.
// Caller must call cmd.Start() and is responsible for closing any opened files.
func startWithIO(cmd *exec.Cmd) (closers []io.Closer, err error) {
	for _, name := range []string{os.DevNull, os.DevNull, os.DevNull} {
		f, err := os.Open(name)
		if err != nil {
			for _, c := range closers {
				_ = c.Close()
			}
			return nil, err
		}
		closers = append(closers, f)
	}

	cmd.Stdin = closers[0].(*os.File)
	cmd.Stdout = closers[1].(*os.File)
	cmd.Stderr = closers[2].(*os.File)
	return closers, nil
}

// envSlice merges base environment (os.Environ()) with overlay; overlay wins.
func envSlice(overlay map[string]string) []string {
	base := os.Environ()
	m := make(map[string]string)
	for _, s := range base {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				m[s[:i]] = s[i+1:]
				break
			}
		}
	}

	maps.Copy(m, overlay)

	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

type exitError struct{ code int }

func (e *exitError) Error() string {
	return "exit status " + strconv.Itoa(e.code)
}
