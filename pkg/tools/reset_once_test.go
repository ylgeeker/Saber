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

package tools

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestResetOnce_Do_runsOnce(t *testing.T) {
	var o ResetOnce
	var runCount int32
	f := func() error {
		atomic.AddInt32(&runCount, 1)
		return nil
	}
	if err := o.Do(f); err != nil {
		t.Fatalf("first Do: %v", err)
	}
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("runCount = %d, want 1", runCount)
	}
	if err := o.Do(f); err != nil {
		t.Fatalf("second Do: %v", err)
	}
	if err := o.Do(f); err != nil {
		t.Fatalf("third Do: %v", err)
	}
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("runCount = %d after repeated Do, want 1", runCount)
	}
}

func TestResetOnce_Do_returnsErrorWhenFuncFails(t *testing.T) {
	var o ResetOnce
	wantErr := errors.New("f failed")
	f := func() error { return wantErr }
	err := o.Do(f)
	if err != wantErr {
		t.Fatalf("Do(f) = %v, want %v", err, wantErr)
	}
	// done not set, so next Do runs f again
	var runCount int32
	f2 := func() error {
		atomic.AddInt32(&runCount, 1)
		return nil
	}
	if err := o.Do(f2); err != nil {
		t.Fatalf("Do after failed f: %v", err)
	}
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("runCount = %d, want 1", runCount)
	}
}

func TestResetOnce_Reset_allowsDoAgain(t *testing.T) {
	var o ResetOnce
	var runCount int32
	f := func() error {
		atomic.AddInt32(&runCount, 1)
		return nil
	}
	o.Do(f)
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("after first Do runCount = %d, want 1", runCount)
	}
	o.Do(f)
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("after second Do runCount = %d, want 1", runCount)
	}
	o.Reset()
	o.Do(f)
	if atomic.LoadInt32(&runCount) != 2 {
		t.Fatalf("after Reset and Do runCount = %d, want 2", runCount)
	}
}

func TestResetOnce_Do_concurrent(t *testing.T) {
	var o ResetOnce
	var runCount int32
	start := make(chan struct{})
	f := func() error {
		atomic.AddInt32(&runCount, 1)
		return nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if err := o.Do(f); err != nil {
				t.Errorf("Do: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("concurrent Do: runCount = %d, want 1", runCount)
	}
	// After Reset, concurrent Do again should run f exactly once
	o.Reset()
	start2 := make(chan struct{})
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start2
			o.Do(f)
		}()
	}
	close(start2)
	wg.Wait()
	if atomic.LoadInt32(&runCount) != 2 {
		t.Fatalf("after Reset concurrent Do: runCount = %d, want 2", runCount)
	}
}
