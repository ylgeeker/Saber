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

package tools_test

import (
	"strings"
	"testing"
	"time"

	"os-artificer/saber/pkg/gerrors"
	"os-artificer/saber/pkg/tools"
)

func TestNewSnowflake_invalidMachineID(t *testing.T) {
	// maxMachineID is 1023 (10 bits)
	_, err := tools.NewSnowflake(1024, time.Now())
	if err == nil {
		t.Fatal("NewSnowflake(1024, ...) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "machine-id out of range") {
		t.Errorf("err = %v; want message containing 'machine-id out of range'", err)
	}
	var ge *gerrors.Error
	if !gerrors.As(err, &ge) || ge.Code() != gerrors.InvalidParameter {
		t.Errorf("expected InvalidParameter code, got %v", err)
	}
}

func TestNewSnowflake_ok(t *testing.T) {
	epoch := time.UnixMilli(0)
	sf, err := tools.NewSnowflake(0, epoch)
	if err != nil {
		t.Fatalf("NewSnowflake(0, epoch): %v", err)
	}
	if sf == nil {
		t.Fatal("NewSnowflake returned nil *Snowflake")
	}
	// machineID at max is also valid
	sf2, err := tools.NewSnowflake(1023, epoch)
	if err != nil {
		t.Fatalf("NewSnowflake(1023, epoch): %v", err)
	}
	if sf2 == nil {
		t.Fatal("NewSnowflake(1023, ...) returned nil *Snowflake")
	}
}

func TestSnowflake_ParseID(t *testing.T) {
	epoch := time.UnixMilli(0)
	const wantMid uint64 = 7
	sf, err := tools.NewSnowflake(wantMid, epoch)
	if err != nil {
		t.Fatalf("NewSnowflake: %v", err)
	}
	id, err := sf.NextID()
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	ts, mid, seq := sf.ParseID(id)
	if mid != wantMid {
		t.Errorf("ParseID(id) machineID = %d, want %d", mid, wantMid)
	}
	if ts < uint64(epoch.UnixMilli()) {
		t.Errorf("ParseID(id) timestamp = %d, want >= %d", ts, epoch.UnixMilli())
	}
	// sequence is 12 bits: 0..4095
	if seq > 4095 {
		t.Errorf("ParseID(id) sequence = %d, want <= 4095", seq)
	}
}

func TestSnowflake(t *testing.T) {
	id, err := tools.MachineID("saber-agent")
	if err != nil {
		t.Fatalf("failed to generate machine-id, %v", err)
	}

	idHash := tools.Hash(id, 10)

	t.Logf("machine-id:%s machine-id hash:%d", id, idHash)

	sf, err := tools.NewSnowflake(idHash, time.Now())
	if err != nil {
		t.Fatalf("failed to create snowflake, %v", err)
	}

	for i := 0; i < 10; i++ {
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("failed to create snowflake id, %v", err)
		}

		ts, mid, seq := sf.ParseID(id)
		t.Logf("id:%d timestamp:%d, machine-id:%d, seq:%d", id, ts, mid, seq)
	}
}
