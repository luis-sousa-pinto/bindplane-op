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

package broadcast

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func TestOptionsParseFunc(t *testing.T) {
	msg := testMessage{
		Value: 10,
	}

	rawData, err := jsoniter.Marshal(&msg)
	require.NoError(t, err)

	opts := &Options[testMessage]{}

	_, err = opts.ParseTo(rawData)
	require.ErrorContains(t, err, "no parse func specified")

	WithParseFunc[testMessage](func(data []byte) (testMessage, error) {
		var newUpdates testMessage
		err := jsoniter.Unmarshal(data, &newUpdates)
		return newUpdates, err
	})(opts)

	newMsg, err := opts.ParseTo(rawData)
	require.NoError(t, err)
	require.Equal(t, msg, newMsg)
}
