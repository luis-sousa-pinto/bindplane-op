// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"testing"
)

// TestSequence is a wrapper around *testing.T that ensures that tests are run in sequence. If a test fails, all
// subsequent tests are skipped.
type TestSequence interface {
	// Run runs the test function f as a subtest of t called name. It returns whether the test succeeded or failed.
	Run(name string, f func(t *testing.T)) bool
}

type testSequence struct {
	failed bool
	count  int
	t      *testing.T
}

// NewTestSequence creates a new test sequence.
func NewTestSequence(t *testing.T) TestSequence {
	return &testSequence{
		t: t,
	}
}

func (ts *testSequence) Run(name string, f func(t *testing.T)) bool {
	ts.count++
	result := ts.t.Run(fmt.Sprintf("%03d %s", ts.count, name), func(t *testing.T) {
		if ts.failed {
			t.SkipNow()
			return
		}
		f(t)
	})
	if !result {
		ts.failed = true
	}
	return result
}
