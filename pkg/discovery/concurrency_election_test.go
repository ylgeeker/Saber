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
)

// mockElection implements discovery.ConcurrencyElection for testing
type mockElection struct {
	campaignErr error
	done        chan struct{}
	closed      bool
	mu          sync.Mutex
}

func newMockElection(campaignErr error) *mockElection {
	return &mockElection{
		campaignErr: campaignErr,
		done:        make(chan struct{}),
	}
}

func (m *mockElection) Campaign(ctx context.Context) error {
	return m.campaignErr
}

func (m *mockElection) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		close(m.done)
		m.closed = true
	}
}

func (m *mockElection) Done() <-chan struct{} {
	return m.done
}

// Verify mockElection implements the interface
var _ ConcurrencyElection = (*mockElection)(nil)

func TestConcurrencyElectionInterface(t *testing.T) {
	t.Run("campaign_success", func(t *testing.T) {
		mock := newMockElection(nil)
		var election ConcurrencyElection = mock

		err := election.Campaign(context.Background())
		if err != nil {
			t.Fatalf("Campaign() unexpected error: %v", err)
		}
		t.Logf("Campaign() succeeded")
	})

	t.Run("campaign_failure", func(t *testing.T) {
		mock := newMockElection(context.DeadlineExceeded)
		var election ConcurrencyElection = mock

		err := election.Campaign(context.Background())
		if err == nil {
			t.Fatal("Campaign() expected error, got nil")
		}
		t.Logf("Campaign() returned expected error: %v", err)
	})

	t.Run("done_channel", func(t *testing.T) {
		mock := newMockElection(nil)
		var election ConcurrencyElection = mock

		done := election.Done()
		if done == nil {
			t.Fatal("Done() returned nil channel")
		}

		select {
		case <-done:
			t.Fatal("Done() channel should not be closed yet")
		default:
			t.Logf("Done() channel is open")
		}
	})

	t.Run("close_idempotent", func(t *testing.T) {
		mock := newMockElection(nil)
		var election ConcurrencyElection = mock

		election.Close()
		election.Close() // should not panic

		select {
		case <-election.Done():
			t.Logf("Done() channel closed after Close()")
		default:
			t.Fatal("Done() channel should be closed after Close()")
		}
	})
}
