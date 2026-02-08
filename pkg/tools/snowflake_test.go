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
	"testing"
	"time"

	"os-artificer/saber/pkg/tools"
)

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
