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

package metrics

import (
	"context"
)

// Nop is a Provider that does nothing
type Nop struct{}

// NewNop creates a new Nop Provider
func NewNop() *Nop {
	return &Nop{}
}

// Start starts the Provider
func (n *Nop) Start(_ context.Context) error {
	return nil
}

// Shutdown shuts down the Provider
func (n *Nop) Shutdown(_ context.Context) error {
	return nil
}
