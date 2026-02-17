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
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewEndpointFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Endpoint
		wantErr bool
		errSub  string
	}{
		{
			name:  "valid tcp",
			input: "tcp://127.0.0.1:8080",
			want:  &Endpoint{Protocol: "tcp", Host: "127.0.0.1", Port: 8080},
		},
		{
			name:  "valid http localhost",
			input: "http://localhost:80",
			want:  &Endpoint{Protocol: "http", Host: "localhost", Port: 80},
		},
		{
			name:  "valid grpc",
			input: "grpc://0.0.0.0:9090",
			want:  &Endpoint{Protocol: "grpc", Host: "0.0.0.0", Port: 9090},
		},
		{
			name:  "valid port min",
			input: "tcp://host:1",
			want:  &Endpoint{Protocol: "tcp", Host: "host", Port: 1},
		},
		{
			name:  "valid port max",
			input: "tcp://host:65535",
			want:  &Endpoint{Protocol: "tcp", Host: "host", Port: 65535},
		},
		{
			name:  "valid IPv6 loopback",
			input: "tcp://[::1]:8080",
			want:  &Endpoint{Protocol: "tcp", Host: "::1", Port: 8080},
		},
		{
			name:  "valid IPv6 with scope",
			input: "tcp://[2001:db8::1]:80",
			want:  &Endpoint{Protocol: "tcp", Host: "2001:db8::1", Port: 80},
		},
		{
			name:    "missing scheme",
			input:   "127.0.0.1:8080",
			wantErr: true,
			errSub:  "missing '://'",
		},
		{
			name:    "empty protocol",
			input:   "://127.0.0.1:8080",
			wantErr: true,
			errSub:  "empty protocol",
		},
		{
			name:    "empty host",
			input:   "tcp://:8080",
			wantErr: true,
			errSub:  "empty host",
		},
		{
			name:    "missing port",
			input:   "tcp://127.0.0.1",
			wantErr: true,
			errSub:  "missing port",
		},
		{
			name:    "empty port",
			input:   "tcp://127.0.0.1:",
			wantErr: true,
			errSub:  "empty port",
		},
		{
			name:    "invalid port non-numeric",
			input:   "tcp://127.0.0.1:abc",
			wantErr: true,
			errSub:  "invalid port",
		},
		{
			name:    "port zero",
			input:   "tcp://127.0.0.1:0",
			wantErr: true,
			errSub:  "out of range",
		},
		{
			name:    "port out of range high",
			input:   "tcp://127.0.0.1:65536",
			wantErr: true,
			errSub:  "out of range",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errSub:  "missing '://'",
		},
		{
			name:    "IPv6 missing ]",
			input:   "tcp://[::1:8080",
			wantErr: true,
			errSub:  "missing ']'",
		},
		{
			name:    "IPv6 no port",
			input:   "tcp://[::1]",
			wantErr: true,
			errSub:  "missing port",
		},
		{
			name:    "IPv6 empty port",
			input:   "tcp://[::1]:",
			wantErr: true,
			errSub:  "empty port",
		},
		{
			name:    "IPv6 empty host in brackets",
			input:   "tcp://[]:8080",
			wantErr: true,
			errSub:  "empty host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEndpointFromString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEndpointFromString() expected error, got nil")
					return
				}
				if tt.errSub != "" && !strings.Contains(err.Error(), tt.errSub) {
					t.Errorf("NewEndpointFromString() error = %v, want substring %q", err, tt.errSub)
				}
				return
			}
			if err != nil {
				t.Errorf("NewEndpointFromString() unexpected error: %v", err)
				return
			}
			if got.Protocol != tt.want.Protocol || got.Host != tt.want.Host || got.Port != tt.want.Port {
				t.Errorf("NewEndpointFromString() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEndpoint_String(t *testing.T) {
	tests := []struct {
		name string
		e    *Endpoint
		want string
	}{
		{"tcp", &Endpoint{Protocol: "tcp", Host: "127.0.0.1", Port: 8080}, "tcp://127.0.0.1:8080"},
		{"http", &Endpoint{Protocol: "http", Host: "localhost", Port: 80}, "http://localhost:80"},
		{"grpc", &Endpoint{Protocol: "grpc", Host: "0.0.0.0", Port: 9090}, "grpc://0.0.0.0:9090"},
		{"IPv6 loopback", &Endpoint{Protocol: "tcp", Host: "::1", Port: 8080}, "tcp://[::1]:8080"},
		{"IPv6", &Endpoint{Protocol: "tcp", Host: "2001:db8::1", Port: 80}, "tcp://[2001:db8::1]:80"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("(*Endpoint).String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEndpoint_HostPort(t *testing.T) {
	tests := []struct {
		name string
		e    *Endpoint
		want string
	}{
		{"IPv4", &Endpoint{Host: "127.0.0.1", Port: 8080}, "127.0.0.1:8080"},
		{"localhost", &Endpoint{Host: "localhost", Port: 80}, "localhost:80"},
		{"IPv6 loopback", &Endpoint{Host: "::1", Port: 8080}, "[::1]:8080"},
		{"IPv6", &Endpoint{Host: "2001:db8::1", Port: 80}, "[2001:db8::1]:80"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.HostPort(); got != tt.want {
				t.Errorf("(*Endpoint).HostPort() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewEndpointFromString_roundTrip(t *testing.T) {
	for _, input := range []string{"tcp://127.0.0.1:8080", "tcp://[::1]:8080"} {
		ep, err := NewEndpointFromString(input)
		if err != nil {
			t.Fatalf("NewEndpointFromString(%q) error: %v", input, err)
		}
		got := ep.String()
		if got != input {
			t.Errorf("round-trip: String() = %q, want %q", got, input)
		}
		ep2, err := NewEndpointFromString(got)
		if err != nil {
			t.Fatalf("NewEndpointFromString(round-trip %q) error: %v", got, err)
		}
		if ep2.Protocol != ep.Protocol || ep2.Host != ep.Host || ep2.Port != ep.Port {
			t.Errorf("round-trip parse = %+v, want %+v", ep2, ep)
		}
	}
}

func TestEndpoint_UnmarshalYAML(t *testing.T) {
	type config struct {
		ListenAddress Endpoint `yaml:"listenAddress"`
	}
	tests := []struct {
		name    string
		yaml    string
		want    Endpoint
		wantErr bool
	}{
		{
			name: "valid tcp",
			yaml: `listenAddress: "tcp://127.0.0.1:26689"`,
			want: Endpoint{Protocol: "tcp", Host: "127.0.0.1", Port: 26689},
		},
		{
			name: "valid http",
			yaml: `listenAddress: "http://localhost:80"`,
			want: Endpoint{Protocol: "http", Host: "localhost", Port: 80},
		},
		{
			name:    "invalid",
			yaml:    `listenAddress: "no-scheme"`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c config
			err := yaml.Unmarshal([]byte(tt.yaml), &c)
			if tt.wantErr {
				if err == nil {
					t.Errorf("yaml.Unmarshal() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("yaml.Unmarshal() error: %v", err)
				return
			}
			got := c.ListenAddress
			if got.Protocol != tt.want.Protocol || got.Host != tt.want.Host || got.Port != tt.want.Port {
				t.Errorf("ListenAddress = %+v, want %+v", got, tt.want)
			}
		})
	}
}
