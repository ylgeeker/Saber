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
	"runtime"
	"strings"
	"testing"
)

func TestRun_ReturnsExitCode(t *testing.T) {
	ctx := context.Background()
	code, err := Run(ctx, "true")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if code != 0 {
		t.Errorf("Run(true) exit code = %d, want 0", code)
	}

	code, err = Run(ctx, "false")
	if err != nil {
		t.Fatalf("Run(false): %v", err)
	}
	if code != 1 {
		t.Errorf("Run(false) exit code = %d, want 1", code)
	}
}

func TestRunWithEnv_ReturnsExitCode(t *testing.T) {
	ctx := context.Background()
	code, err := RunWithEnv(ctx, nil, "true")
	if err != nil {
		t.Fatalf("RunWithEnv: %v", err)
	}
	if code != 0 {
		t.Errorf("RunWithEnv exit code = %d, want 0", code)
	}
}

func TestRunWithStdout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	stdout, err := RunWithStdout(ctx, "sh", "-c", "echo hello")
	if err != nil {
		t.Fatalf("RunWithStdout: %v", err)
	}
	if got := strings.TrimSpace(string(stdout)); got != "hello" {
		t.Errorf("RunWithStdout = %q, want hello", got)
	}
}

func TestRunWithStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	stderr, err := RunWithStderr(ctx, "sh", "-c", "echo err >&2")
	if err != nil {
		t.Fatalf("RunWithStderr: %v", err)
	}
	if got := strings.TrimSpace(string(stderr)); got != "err" {
		t.Errorf("RunWithStderr = %q, want err", got)
	}
}

func TestRunWithStdoutAndStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	stdout, stderr, err := RunWithStdoutAndStderr(ctx, "sh", "-c", "echo out; echo err >&2")
	if err != nil {
		t.Fatalf("RunWithStdoutAndStderr: %v", err)
	}
	if got := strings.TrimSpace(string(stdout)); got != "out" {
		t.Errorf("stdout = %q, want out", got)
	}
	if got := strings.TrimSpace(string(stderr)); got != "err" {
		t.Errorf("stderr = %q, want err", got)
	}
}
