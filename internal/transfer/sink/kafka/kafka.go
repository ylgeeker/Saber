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

package kafka

import (
	"context"

	"os-artificer/saber/internal/transfer/sink"
	"os-artificer/saber/pkg/logger"
	"os-artificer/saber/pkg/proto"

	"github.com/segmentio/kafka-go"
	kafkago "github.com/segmentio/kafka-go"
)

var _ sink.Sink = (*KafkaSink)(nil)

// KafkaSink implements sink.Sink by writing TransferRequest to Kafka.
type KafkaSink struct {
	writer *kafkago.Writer
}

// New returns a Sink that writes to the given Kafka writer.
func New(writer *kafkago.Writer) *KafkaSink {
	return &KafkaSink{writer: writer}
}

// Write implements sink.Sink.
func (k *KafkaSink) Write(ctx context.Context, req *proto.TransferRequest) error {
	if k.writer == nil || req == nil || len(req.Payload) == 0 {
		return nil
	}

	key := []byte(req.ClientID)
	if err := k.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: req.Payload,
	}); err != nil {
		logger.Warnf("write to kafka failed: %v", err)
		return err
	}
	return nil
}

// Close implements sink.Sink.
func (k *KafkaSink) Close() error {
	if k.writer == nil {
		return nil
	}

	if err := k.writer.Close(); err != nil {
		logger.Warnf("close kafka writer failed: %v", err)
		return err
	}

	return nil
}
