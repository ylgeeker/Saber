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
	"runtime"
	"testing"
	"time"
)

func TestDaemon_StartAndStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	d := NewDaemon(ctx, nil, "sh", "-c", "exit 0")
	if err := d.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	d.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestDaemon_StartWithEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on sh")
	}
	ctx := context.Background()
	env := map[string]string{"DAEMON_TEST": "1"}
	d := NewDaemon(ctx, env, "sh", "-c", "exit 0")
	if err := d.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	d.Stop()
}

func TestDaemon_StartFailsInvalidCommand(t *testing.T) {
	ctx := context.Background()
	d := NewDaemon(ctx, nil, "/nonexistent/binary", "arg")
	err := d.Start()
	if err == nil {
		t.Fatal("Start expected to fail")
	}
}
