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
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestOptionUser(t *testing.T) {
	user := "testuser"
	opt := OptionUser(user)
	if opt == nil {
		t.Fatal("OptionUser() returned nil")
	}
	t.Logf("OptionUser(%s) created successfully", user)
}

func TestOptionPassword(t *testing.T) {
	password := "testpassword"
	opt := OptionPassword(password)
	if opt == nil {
		t.Fatal("OptionPassword() returned nil")
	}
	t.Logf("OptionPassword() created successfully")
}

func TestOptionBufferMaxSize(t *testing.T) {
	size := 2048
	opt := OptionBufferMaxSize(size)
	if opt == nil {
		t.Fatal("OptionBufferMaxSize() returned nil")
	}
	t.Logf("OptionBufferMaxSize(%d) created successfully", size)
}

func TestOptionTTL(t *testing.T) {
	tests := []struct {
		name string
		ttl  int
	}{
		{"normal_ttl", 10},
		{"below_default_ttl", 3},
		{"zero_ttl", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := OptionTTL(tt.ttl)
			if opt == nil {
				t.Fatalf("OptionTTL(%d) returned nil", tt.ttl)
			}
			t.Logf("OptionTTL(%d) created successfully", tt.ttl)
		})
	}
}

func TestOptionServiceID(t *testing.T) {
	serviceID := "test-service-id"
	opt := OptionServiceID(serviceID)
	if opt == nil {
		t.Fatal("OptionServiceID() returned nil")
	}
	t.Logf("OptionServiceID(%s) created successfully", serviceID)
}

func TestOptionServiceName(t *testing.T) {
	serviceName := "test-service"
	opt := OptionServiceName(serviceName)
	if opt == nil {
		t.Fatal("OptionServiceName() returned nil")
	}
	t.Logf("OptionServiceName(%s) created successfully", serviceName)
}

func TestOptionEndpoints(t *testing.T) {
	endpoints := []string{"localhost:2379", "localhost:2380"}
	opt := OptionEndpoints(endpoints)
	if opt == nil {
		t.Fatal("OptionEndpoints() returned nil")
	}
	t.Logf("OptionEndpoints(%v) created successfully", endpoints)
}

func TestOptionDialTimeout(t *testing.T) {
	timeout := 10 * time.Second
	opt := OptionDialTimeout(timeout)
	if opt == nil {
		t.Fatal("OptionDialTimeout() returned nil")
	}
	t.Logf("OptionDialTimeout(%v) created successfully", timeout)
}

func TestOptionAutoSyncInterval(t *testing.T) {
	interval := 30 * time.Second
	opt := OptionAutoSyncInterval(interval)
	if opt == nil {
		t.Fatal("OptionAutoSyncInterval() returned nil")
	}
	t.Logf("OptionAutoSyncInterval(%v) created successfully", interval)
}

func TestOptionKeepAliveTime(t *testing.T) {
	keepAlive := 15 * time.Second
	opt := OptionKeepAliveTime(keepAlive)
	if opt == nil {
		t.Fatal("OptionKeepAliveTime() returned nil")
	}
	t.Logf("OptionKeepAliveTime(%v) created successfully", keepAlive)
}

func TestOptionKeepAliveTimeout(t *testing.T) {
	timeout := 5 * time.Second
	opt := OptionKeepAliveTimeout(timeout)
	if opt == nil {
		t.Fatal("OptionKeepAliveTimeout() returned nil")
	}
	t.Logf("OptionKeepAliveTimeout(%v) created successfully", timeout)
}

func TestOptionRegistryRootKeyPrefix(t *testing.T) {
	prefix := "/custom/registry"
	opt := OptionRegistryRootKeyPrefix(prefix)
	if opt == nil {
		t.Fatal("OptionRegistryRootKeyPrefix() returned nil")
	}
	t.Logf("OptionRegistryRootKeyPrefix(%s) created successfully", prefix)
}

func TestOptionMaxUnaryRetries(t *testing.T) {
	retries := uint(5)
	opt := OptionMaxUnaryRetries(retries)
	if opt == nil {
		t.Fatal("OptionMaxUnaryRetries() returned nil")
	}
	t.Logf("OptionMaxUnaryRetries(%d) created successfully", retries)
}

func TestOptionLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	opt := OptionLogger(logger)
	if opt == nil {
		t.Fatal("OptionLogger() returned nil")
	}
	t.Logf("OptionLogger() created successfully")
}
