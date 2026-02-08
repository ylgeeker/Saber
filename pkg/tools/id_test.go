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

	"os-artificer/saber/pkg/tools"
)

func TestID(t *testing.T) {
	id, err := tools.MachineID("saber-agent")
	if err != nil {
		t.Fatalf("failed to generate machine-id:%v", err)
	}

	id2, err := tools.MachineID("saber-agent")
	if err != nil {
		t.Fatalf("failed to generate machine-id:%v", err)
	}

	if id != id2 {
		t.Fatalf("machine-id is invalid, id(%s), id2(%s)", id, id2)
	}

	t.Logf("machine-id:%s", id)
}

func TestSequenceID(t *testing.T) {
	for i := 0; i < 10; i++ {
		id, err := tools.NewSequenceID()
		if err != nil {
			t.Fatalf("failed to generate sequence-id:%v", err)
		}
		t.Logf("sequence-id:%d", id)
	}
}
