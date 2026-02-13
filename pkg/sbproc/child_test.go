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
	"testing"
)

func TestStart_ReturnsPIDAndIgnoresIO(t *testing.T) {
	ctx := context.Background()
	child, err := Start(ctx, "true")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	pid := child.PID()
	if pid <= 0 {
		t.Errorf("PID() = %d, want positive", pid)
	}

	code, err := child.Wait()
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestStart_WaitReturnsExitCode(t *testing.T) {
	ctx := context.Background()
	child, err := Start(ctx, "false")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	code, err := child.Wait()
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestStart_ContextCancelKillsProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sleep")
	}
	ctx, cancel := context.WithCancel(context.Background())
	child, err := Start(ctx, "sleep", "10")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	cancel()
	code, err := child.Wait()
	if err == nil && code == 0 {
		t.Errorf("expected error or non-zero exit after context cancel, got code=%d err=%v", code, err)
	}
}
