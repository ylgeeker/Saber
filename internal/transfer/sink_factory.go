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

package transfer

import (
	"fmt"

	"os-artificer/saber/internal/transfer/config"
	"os-artificer/saber/internal/transfer/sink"
	"os-artificer/saber/internal/transfer/sink/kafka"
	"os-artificer/saber/pkg/logger"

	kafkago "github.com/segmentio/kafka-go"
)

const sinkTypeKafka = "kafka"

// NewSinkFromConfig builds a sink.Sink from the given sink configs.
// Returns nil, nil when configs is empty or all entries are skipped (e.g. unknown type);
// the caller may treat nil as "no sink" (requests dropped).
// Returns an error when a configured sink fails to build (e.g. invalid kafka config).
func NewSinkFromConfig(configs []config.SinkConfig) (sink.Sink, error) {
	if len(configs) == 0 {
		return nil, nil
	}

	var sinks []sink.Sink
	for i, sc := range configs {
		s, err := buildSink(&sc)
		if err != nil {
			return nil, fmt.Errorf("sink[%d] type %q: %w", i, sc.Type, err)
		}
		if s != nil {
			sinks = append(sinks, s)
		}
	}

	if len(sinks) == 0 {
		logger.Warnf("no sink created from %d config(s), transfer requests will be dropped", len(configs))
		return nil, nil
	}

	if len(sinks) == 1 {
		return sinks[0], nil
	}
	return sink.NewMultiSink(sinks), nil
}

func buildSink(sc *config.SinkConfig) (sink.Sink, error) {
	if sc == nil {
		return nil, nil
	}

	switch sc.Type {
	case sinkTypeKafka:
		return buildKafkaSink(sc.Config)

	default:
		logger.Warnf("unknown sink type %q, skipping", sc.Type)
		return nil, nil
	}
}

func buildKafkaSink(cfg map[string]any) (sink.Sink, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config is empty")
	}

	brokers, err := parseStringSlice(cfg, "brokers")
	if err != nil || len(brokers) == 0 {
		return nil, fmt.Errorf("brokers: %w", err)
	}

	topic, err := parseString(cfg, "topic")
	if err != nil || topic == "" {
		return nil, fmt.Errorf("topic: %w", err)
	}

	writer := &kafkago.Writer{
		Addr:  kafkago.TCP(brokers...),
		Topic: topic,
	}

	return kafka.New(writer), nil
}

func parseString(m map[string]any, key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("missing %q", key)
	}

	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%q must be a string", key)
	}
	return s, nil
}

func parseStringSlice(m map[string]any, key string) ([]string, error) {
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing %q", key)
	}
	switch raw := v.(type) {
	case []string:
		return raw, nil

	case []any:
		out := make([]string, 0, len(raw))
		for i, item := range raw {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%q[%d] must be a string", key, i)
			}
			out = append(out, s)
		}

		return out, nil

	default:
		return nil, fmt.Errorf("%q must be a string slice", key)
	}
}
