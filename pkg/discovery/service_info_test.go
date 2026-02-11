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
	"encoding/json"
	"testing"
	"time"
)

func TestServiceInfoCreation(t *testing.T) {
	now := time.Now()
	info := ServiceInfo{
		ID:        "test-id",
		Name:      "test-service",
		Nice:      0,
		IPs:       []string{"127.0.0.1", "127.0.0.2"},
		StartTime: now,
		Uptime:    "1h30m",
		UpdatedAt: now,
	}

	if info.ID != "test-id" {
		t.Fatalf("ServiceInfo.ID = %v, want test-id", info.ID)
	}
	if info.Name != "test-service" {
		t.Fatalf("ServiceInfo.Name = %v, want test-service", info.Name)
	}
	if len(info.IPs) != 2 {
		t.Fatalf("ServiceInfo.IPs length = %v, want 2", len(info.IPs))
	}
	t.Logf("ServiceInfo created: %+v", info)
}

func TestServiceInfoWithEndpoints(t *testing.T) {
	now := time.Now()
	listenAddr := &Endpoint{Host: "0.0.0.0", Port: 8080}
	probeAddr := &Endpoint{Host: "0.0.0.0", Port: 8081}

	info := ServiceInfo{
		ID:            "test-id",
		Name:          "test-service",
		Nice:          10,
		IPs:           []string{"127.0.0.1"},
		ListenAddress: listenAddr,
		ProbeEndpoint: probeAddr,
		StartTime:     now,
		Uptime:        "2h",
		UpdatedAt:     now,
	}

	if info.ListenAddress == nil {
		t.Fatal("ServiceInfo.ListenAddress should not be nil")
	}
	if info.ProbeEndpoint == nil {
		t.Fatal("ServiceInfo.ProbeEndpoint should not be nil")
	}
	if info.ListenAddress.Port != 8080 {
		t.Fatalf("ServiceInfo.ListenAddress.Port = %v, want 8080", info.ListenAddress.Port)
	}
	t.Logf("ServiceInfo with endpoints: %+v", info)
}

func TestServiceInfoJSONSerialization(t *testing.T) {
	now := time.Now()
	info := ServiceInfo{
		ID:        "json-test-id",
		Name:      "json-test-service",
		Nice:      5,
		IPs:       []string{"127.0.0.1"},
		StartTime: now,
		Uptime:    "30m",
		UpdatedAt: now,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	t.Logf("ServiceInfo JSON: %s", string(data))

	var decoded ServiceInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.ID != info.ID {
		t.Fatalf("decoded.ID = %v, want %v", decoded.ID, info.ID)
	}
	if decoded.Name != info.Name {
		t.Fatalf("decoded.Name = %v, want %v", decoded.Name, info.Name)
	}
	t.Logf("ServiceInfo deserialized: %+v", decoded)
}

func TestServiceInfoNilEndpoints(t *testing.T) {
	info := ServiceInfo{
		ID:   "nil-endpoint-test",
		Name: "test",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded ServiceInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if decoded.ListenAddress != nil {
		t.Fatal("decoded.ListenAddress should be nil")
	}
	if decoded.ProbeEndpoint != nil {
		t.Fatal("decoded.ProbeEndpoint should be nil")
	}
	t.Logf("ServiceInfo with nil endpoints handled correctly")
}
