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
	"encoding"
	"fmt"
	"net"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Endpoint represents a host:port endpoint.
type Endpoint struct {
	Protocol string
	Host     string
	Port     int
}

// NewEndpointFromString parses a string in the form "protocol://host:port" (e.g. "tcp://127.0.0.1:8080").
// Returns an error if the format is invalid.
func NewEndpointFromString(s string) (*Endpoint, error) {
	protocol, rest, err := splitSchemeAndAuthority(s)
	if err != nil {
		return nil, err
	}

	host, portStr, err := parseHostPort(rest, s)
	if err != nil {
		return nil, err
	}

	port, err := parsePort(portStr, s)
	if err != nil {
		return nil, err
	}

	return &Endpoint{
		Protocol: protocol,
		Host:     host,
		Port:     port,
	}, nil
}

// String returns the endpoint as a string.
// IPv6 hosts (containing ':') are emitted as protocol://[host]:port.
func (e *Endpoint) String() string {
	if strings.Contains(e.Host, ":") {
		return fmt.Sprintf("%s://[%s]:%d", e.Protocol, e.Host, e.Port)
	}
	return fmt.Sprintf("%s://%s:%d", e.Protocol, e.Host, e.Port)
}

// HostPort returns the endpoint as "host:port" or "[host]:port" for use with net.Listen / net.Dial.
func (e *Endpoint) HostPort() string {
	return net.JoinHostPort(e.Host, strconv.Itoa(e.Port))
}

// UnmarshalYAML implements yaml.Unmarshaler so that a YAML string (e.g. "tcp://127.0.0.1:26689") is parsed into Endpoint.
func (e *Endpoint) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("listenAddress: expected string (e.g. tcp://host:port): %w", err)
	}
	ep, err := NewEndpointFromString(s)
	if err != nil {
		return err
	}
	*e = *ep
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler so that viper/mapstructure can decode a string into Endpoint.
func (e *Endpoint) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	ep, err := NewEndpointFromString(s)
	if err != nil {
		return err
	}
	*e = *ep
	return nil
}

// Ensure *Endpoint implements encoding.TextUnmarshaler at compile time.
var _ encoding.TextUnmarshaler = (*Endpoint)(nil)

// splitSchemeAndAuthority splits s into "protocol" and "rest" (host:port) at "://" and validates both are non-empty.
func splitSchemeAndAuthority(s string) (protocol, rest string, err error) {
	if !strings.Contains(s, "://") {
		return "", "", fmt.Errorf("invalid endpoint: missing '://' (expected format protocol://host:port): %s", s)
	}
	parts := strings.SplitN(s, "://", 2)
	protocol, rest = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if protocol == "" {
		return "", "", fmt.Errorf("invalid endpoint: empty protocol: %s", s)
	}
	if rest == "" {
		return "", "", fmt.Errorf("invalid endpoint: empty host:port: %s", s)
	}
	return protocol, rest, nil
}

// parseHostPort extracts host and port string from rest (authority part after "://"). Handles IPv6 [host]:port and plain host:port.
func parseHostPort(rest, forErr string) (host, portStr string, err error) {
	if len(rest) > 0 && rest[0] == '[' {
		rb := strings.Index(rest, "]")
		if rb == -1 {
			return "", "", fmt.Errorf("invalid endpoint: missing ']' in IPv6 address: %s", forErr)
		}
		host = strings.TrimSpace(rest[1:rb])
		if host == "" {
			return "", "", fmt.Errorf("invalid endpoint: empty host in brackets: %s", forErr)
		}
		after := strings.TrimSpace(rest[rb+1:])
		if after == "" || after[0] != ':' {
			return "", "", fmt.Errorf("invalid endpoint: missing port after ']' (expected [host]:port): %s", forErr)
		}
		portStr = strings.TrimSpace(after[1:])
	} else {
		lastColon := strings.LastIndex(rest, ":")
		if lastColon == -1 {
			return "", "", fmt.Errorf("invalid endpoint: missing port (expected protocol://host:port): %s", forErr)
		}
		host = strings.TrimSpace(rest[:lastColon])
		portStr = strings.TrimSpace(rest[lastColon+1:])
	}

	if host == "" {
		return "", "", fmt.Errorf("invalid endpoint: empty host: %s", forErr)
	}

	if portStr == "" {
		return "", "", fmt.Errorf("invalid endpoint: empty port: %s", forErr)
	}

	return host, portStr, nil
}

// parsePort parses portStr as an integer and validates it is in 1-65535.
func parsePort(portStr, forErr string) (port int, err error) {
	port, err = strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid endpoint: invalid port %q: %w", portStr, err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid endpoint: port %d out of range 1-65535: %s", port, forErr)
	}
	return port, nil
}
