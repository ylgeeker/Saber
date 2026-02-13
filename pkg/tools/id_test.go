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

func TestHash(t *testing.T) {
	// Same input and bits => same output (deterministic)
	const s = "hello"
	h1 := tools.Hash(s, 10)
	h2 := tools.Hash(s, 10)
	if h1 != h2 {
		t.Errorf("Hash(%q, 10) = %d, %d; want same value", s, h1, h2)
	}
	// Different input => different output (with high probability)
	hOther := tools.Hash("world", 10)
	if h1 == hOther {
		t.Errorf("Hash(%q, 10) == Hash(\"world\", 10) = %d; expect different", s, h1)
	}
	// Different bits => different mask; bits=8 => result < 256
	h8 := tools.Hash(s, 8)
	if h8 >= 256 {
		t.Errorf("Hash(%q, 8) = %d; want < 256", s, h8)
	}
	// bits=1 => result 0 or 1
	h1bit := tools.Hash(s, 1)
	if h1bit != 0 && h1bit != 1 {
		t.Errorf("Hash(%q, 1) = %d; want 0 or 1", s, h1bit)
	}
}
