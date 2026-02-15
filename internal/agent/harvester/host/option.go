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

package host

import (
	"encoding/json"
	"fmt"
	"time"

	"os-artificer/saber/pkg/logger"
)

// Duration supports JSON unmarshaling from string (e.g. "1s", "10m") or number (nanoseconds).
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch val := v.(type) {
	case string:
		parsed, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		*d = Duration(parsed)
		return nil

	case float64:
		*d = Duration(int64(val))
		return nil

	default:
		return fmt.Errorf("invalid duration: %v", v)
	}
}

// Duration returns the value as time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Options is the option for the host plugin.
type Options struct {
	Interval Duration `yaml:"interval" json:"interval"`
	Timeout  Duration `yaml:"timeout" json:"timeout"`
}

// OptionsFromAny converts opts (any) to Options. Supports nil, Options, and map[string]any (via JSON).
func OptionsFromAny(opts any) (Options, error) {
	if opts == nil {
		return Options{}, nil
	}

	if o, ok := opts.(Options); ok {
		return o, nil
	}

	if m, ok := opts.(map[string]any); ok {
		data, err := json.Marshal(m)
		if err != nil {
			return Options{}, fmt.Errorf("host options marshal: %w", err)
		}

		var out Options
		if err := json.Unmarshal(data, &out); err != nil {
			logger.Errorf("host options unmarshal: %v, %v", err, string(data))
			return Options{}, fmt.Errorf("host options unmarshal: %w", err)
		}

		return out, nil
	}

	return Options{}, fmt.Errorf("unsupported options type: %T", opts)
}
