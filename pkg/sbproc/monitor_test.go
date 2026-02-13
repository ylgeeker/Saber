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
	"testing"
)

func TestIsRunning(t *testing.T) {
	if IsRunning(0) {
		t.Errorf("IsRunning(0) = true, want false")
	}

	ctx := context.Background()
	child, err := Start(ctx, "true")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	pid := child.PID()
	if !IsRunning(pid) {
		t.Errorf("IsRunning(%d) = false during run, want true", pid)
	}
	_, _ = child.Wait()
}
