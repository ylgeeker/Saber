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

package discovery

import (
	"context"
	"sync"
	"testing"

	"os-artificer/saber/pkg/gerrors"
)

// mockMutex implements discovery.ConcurrencyMutex for testing
type mockMutex struct {
	locked     bool
	tryLockErr error
	unlockErr  error
	mu         sync.Mutex
	closed     bool
}

func newMockMutex() *mockMutex {
	return &mockMutex{}
}

func (m *mockMutex) TryLock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tryLockErr != nil {
		return m.tryLockErr
	}

	if m.locked {
		return gerrors.New(gerrors.Failure, "already locked")
	}

	m.locked = true
	return nil
}

func (m *mockMutex) Unlock(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.unlockErr != nil {
		return m.unlockErr
	}

	if !m.locked {
		return gerrors.New(gerrors.Failure, "not locked")
	}

	m.locked = false
	return nil
}

func (m *mockMutex) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

// Verify mockMutex implements the interface
var _ ConcurrencyMutex = (*mockMutex)(nil)

func TestConcurrencyMutexInterface(t *testing.T) {
	t.Run("trylock_success", func(t *testing.T) {
		mock := newMockMutex()
		var mutex ConcurrencyMutex = mock

		err := mutex.TryLock(context.Background())
		if err != nil {
			t.Fatalf("TryLock() unexpected error: %v", err)
		}
		t.Logf("TryLock() succeeded")
	})

	t.Run("unlock_success", func(t *testing.T) {
		mock := newMockMutex()
		var mutex ConcurrencyMutex = mock

		_ = mutex.TryLock(context.Background())
		err := mutex.Unlock(context.Background())
		if err != nil {
			t.Fatalf("Unlock() unexpected error: %v", err)
		}
		t.Logf("Unlock() succeeded")
	})

	t.Run("double_lock_fails", func(t *testing.T) {
		mock := newMockMutex()
		var mutex ConcurrencyMutex = mock

		_ = mutex.TryLock(context.Background())
		err := mutex.TryLock(context.Background())
		if err == nil {
			t.Fatal("TryLock() expected error on double lock, got nil")
		}
		t.Logf("TryLock() returned expected error: %v", err)
	})

	t.Run("unlock_without_lock_fails", func(t *testing.T) {
		mock := newMockMutex()
		var mutex ConcurrencyMutex = mock

		err := mutex.Unlock(context.Background())
		if err == nil {
			t.Fatal("Unlock() expected error without lock, got nil")
		}
		t.Logf("Unlock() returned expected error: %v", err)
	})

	t.Run("concurrent_access", func(t *testing.T) {
		mock := newMockMutex()
		var mutex ConcurrencyMutex = mock

		var wg sync.WaitGroup
		successCount := 0
		var countMu sync.Mutex

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := mutex.TryLock(context.Background()); err == nil {
					countMu.Lock()
					successCount++
					countMu.Unlock()
					_ = mutex.Unlock(context.Background())
				}
			}()
		}

		wg.Wait()
		if successCount == 0 {
			t.Fatal("concurrent TryLock: at least one goroutine should succeed")
		}
		t.Logf("concurrent TryLock: %d succeeded", successCount)
	})
}
