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

package ui

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/observiq/bindplane-op/version"
	"github.com/stretchr/testify/require"
)

func TestGenerateGlobalJS(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"renders template",
			fmt.Sprintf("var __BINDPLANE_VERSION__ = \"%s\";\n", version.NewVersion().String()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateGlobalJS()

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_templateStringParses(t *testing.T) {
	_, err := template.New("globals").Parse(templateStr)
	require.NoError(t, err)
}

func Test_templateExecute(t *testing.T) {
	tmp, err := template.New("globals").Parse(templateStr)
	require.NoError(t, err)

	opts := newConfigOptions()

	w := bytes.NewBufferString("")
	err = tmp.Execute(w, opts)
	require.NoError(t, err)
}
