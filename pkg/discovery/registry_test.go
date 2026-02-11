//go:build integration

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

package discovery_test

import (
	"context"
	"testing"
	"time"

	"os-artificer/saber/pkg/gerrors"
)

func TestRegsitrySetService(t *testing.T) {
	ctx := context.Background()

	err := reg.SetService(ctx, "registry-test-value")
	if err != nil {
		t.Errorf("failed to set service, errmsg: %v", err)
	}

	resp, err := etcdClient.Get(ctx, reg.GetRootKey())
	if err != nil {
		t.Errorf("failed to get service value, errmsg: %v", err)
	}

	if len(resp.Kvs) == 0 {
		t.Errorf("undefined service value")
	}
	if string(resp.Kvs[0].Value) != "registry-test-value" {
		t.Errorf("service value does not match. Expected: registry-test-value, Got: %s", string(resp.Kvs[0].Value))
	}

	if resp.Kvs[0].Lease == 0 {
		t.Errorf("service has no lease attached")
	}

}

func TestRegistrySet(t *testing.T) {
	ctx := context.Background()

	err := reg.Set(ctx, "registry-test-key", "registry-test-value")
	if err != nil {
		t.Errorf("failed to set key, errmsg: %v", err)
	}

	time.Sleep(10 * time.Second)

	expectedKey := reg.GetRootKey() + "/registry-test-key"
	resp, err := etcdClient.Get(ctx, expectedKey)

	if err != nil {
		t.Errorf("failed to get child node value, errmsg: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Errorf("missing child node value, errmsg: %v", err)
	}
	if string(resp.Kvs[0].Value) != "registry-test-value" {
		t.Errorf("service value does not match. Expected: registry-test-value, Got: %s", string(resp.Kvs[0].Value))
	}
}

func TestRegistryLeaseManagement(t *testing.T) {
	ctx := context.Background()

	err := reg.SetService(ctx, "registry-test-value")
	if err != nil {
		t.Errorf("failed to set service, errmsg: %v", err)
	}

	resp, err := etcdClient.Get(ctx, reg.GetRootKey())
	if err != nil {
		t.Errorf("failed to get service value, errmsg: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Errorf("service value expired, errmsg: %v", err)
	}

	initialLeaseID := resp.Kvs[0].Lease

	time.Sleep(10 * time.Second)

	resp, err = etcdClient.Get(ctx, reg.GetRootKey())
	if err != nil {
		t.Errorf("failed to get service value, errmsg: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Errorf("service value expired, errmsg: %v", err)
	}

	if resp.Kvs[0].Lease != initialLeaseID {
		t.Errorf("The lease ID has changed, likely due to lease renewal.")
	}
}

func TestRegistryInvalidParameters(t *testing.T) {
	ctx := context.Background()

	err := reg.Set(ctx, "", "registry-test-value")
	if err == nil {
		t.Errorf("expect an error to be returned for empty keys")
	}
	if err.(*gerrors.GError).Code() != gerrors.InvalidParameter {
		t.Errorf("expected error code: InvalidParameter, actual: %v", err)
	}
}

func TestRegistryClose(t *testing.T) {
	newReg := client.CreateRegistry()

	ctx := context.Background()

	err := newReg.SetService(ctx, "registry-test-value")
	if err != nil {
		t.Errorf("failed to set service, errmsg: %v", err)
	}

	newReg.Close()
}
