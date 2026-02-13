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
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStartWithEnv_SetsEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.txt")
	ctx := context.Background()
	env := map[string]string{"SBPROC_TEST_VAR": "sbproc_value"}
	child, err := StartWithEnv(ctx, env, "sh", "-c", "echo $SBPROC_TEST_VAR > "+outFile)
	if err != nil {
		t.Fatalf("StartWithEnv: %v", err)
	}
	code, err := child.Wait()
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "sbproc_value" {
		t.Errorf("env output = %q, want %q", got, "sbproc_value")
	}
}

func TestStartWithOutput_CapturesStdoutAndStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	child, stdout, stderr, err := StartWithOutput(ctx, "sh", "-c", "echo out; echo err >&2")
	if err != nil {
		t.Fatalf("StartWithOutput: %v", err)
	}
	if child.PID() <= 0 {
		t.Errorf("PID() = %d, want positive", child.PID())
	}

	outStr := strings.TrimSpace(string(stdout))
	errStr := strings.TrimSpace(string(stderr))
	if outStr != "out" {
		t.Errorf("stdout = %q, want %q", outStr, "out")
	}
	if errStr != "err" {
		t.Errorf("stderr = %q, want %q", errStr, "err")
	}
}

func TestStartWithStreams_StreamsStdoutAndStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	child, err := StartWithStreams(ctx, "sh", "-c", "echo a; sleep 0.05; echo b; echo c >&2")
	if err != nil {
		t.Fatalf("StartWithStreams: %v", err)
	}
	if child.PID() <= 0 {
		t.Errorf("PID() = %d, want positive", child.PID())
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	doneOut := make(chan struct{})
	doneErr := make(chan struct{})
	go func() {
		for b := range child.Stdout() {
			stdoutBuf.Write(b)
		}
		close(doneOut)
	}()
	go func() {
		for b := range child.Stderr() {
			stderrBuf.Write(b)
		}
		close(doneErr)
	}()

	code, err := child.Wait()
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	<-doneOut
	<-doneErr

	stdoutStr := strings.TrimSpace(stdoutBuf.String())
	stderrStr := strings.TrimSpace(stderrBuf.String())
	if !strings.Contains(stdoutStr, "a") || !strings.Contains(stdoutStr, "b") {
		t.Errorf("stdout = %q, want to contain a and b", stdoutStr)
	}
	if !strings.Contains(stderrStr, "c") {
		t.Errorf("stderr = %q, want to contain c", stderrStr)
	}
}
