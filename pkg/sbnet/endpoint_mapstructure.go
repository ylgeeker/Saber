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

package sbnet

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

// StringToEndpointHookFunc returns a mapstructure DecodeHook that parses a string
// into Endpoint (e.g. "tcp://127.0.0.1:26689") for use with viper.Unmarshal.
// Use it in viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(..., sbnet.StringToEndpointHookFunc())).
func StringToEndpointHookFunc() mapstructure.DecodeHookFunc {
	return func(f, t reflect.Type, data any) (any, error) {
		if f != nil && f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(Endpoint{}) {
			return data, nil
		}
		s, ok := data.(string)
		if !ok {
			return data, nil
		}
		ep, err := NewEndpointFromString(s)
		if err != nil {
			return nil, err
		}
		return *ep, nil
	}
}
