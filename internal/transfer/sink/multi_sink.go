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

package sink

import (
	"context"

	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"
)

var _ Sink = (*MultiSink)(nil)

// MultiSink writes each TransferRequest to all underlying sinks in order.
// Close closes all sinks.
type MultiSink struct {
	sinks []Sink
}

// NewMultiSink returns a Sink that forwards Write to each sink and Close to all.
func NewMultiSink(sinks []Sink) *MultiSink {
	return &MultiSink{sinks: sinks}
}

// Write implements Sink.
func (m *MultiSink) Write(ctx context.Context, req *proto.TransferRequest) error {
	for _, s := range m.sinks {
		if err := s.Write(ctx, req); err != nil {
			logger.Warnf("multi-sink write failed: %v", err)
			// continue to other sinks
		}
	}
	return nil
}

// Close implements Sink.
func (m *MultiSink) Close() error {
	for _, s := range m.sinks {
		if err := s.Close(); err != nil {
			logger.Warnf("multi-sink close failed: %v", err)
		}
	}
	return nil
}
