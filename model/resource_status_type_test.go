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

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceSetLatest(t *testing.T) {
	t.Run("configuration supports latest", func(t *testing.T) {
		config := &Configuration{}
		require.False(t, config.Status.Latest)
		require.False(t, config.IsLatest())
		config.SetLatest(true)
		require.True(t, config.Status.Latest)
		require.True(t, config.IsLatest())
	})
	t.Run("sourceType supports latest", func(t *testing.T) {
		sourceType := &SourceType{}
		require.False(t, sourceType.Status.Latest)
		require.False(t, sourceType.IsLatest())
		sourceType.SetLatest(true)
		require.True(t, sourceType.Status.Latest)
		require.True(t, sourceType.IsLatest())
	})
	t.Run("AnyResource uses a map", func(t *testing.T) {
		anyResource := &AnyResource{}
		require.Nil(t, anyResource.Status["latest"])
		require.False(t, anyResource.IsLatest())
		anyResource.SetLatest(true)
		require.NotNil(t, anyResource.Status)
		require.True(t, anyResource.Status["latest"].(bool))
		require.True(t, anyResource.IsLatest())
	})
}

func TestResourceSetPending(t *testing.T) {
	t.Run("configuration supports pending", func(t *testing.T) {
		config := &Configuration{}
		require.False(t, config.Status.Pending)
		require.False(t, config.IsPending())
		config.SetPending(true)
		require.True(t, config.Status.Pending)
		require.True(t, config.IsPending())
	})
	t.Run("sourceType doesn't support pending", func(t *testing.T) {
		sourceType := &SourceType{}
		require.False(t, sourceType.IsPending())
		sourceType.SetPending(true)
		require.False(t, sourceType.IsPending())
	})
	t.Run("AnyResource uses a map", func(t *testing.T) {
		anyResource := &AnyResource{}
		require.Nil(t, anyResource.Status["pending"])
		require.False(t, anyResource.IsPending())
		anyResource.SetPending(true)
		require.NotNil(t, anyResource.Status)
		require.True(t, anyResource.Status["pending"].(bool))
		require.True(t, anyResource.IsPending())
	})
}

func TestResourceSetCurrent(t *testing.T) {
	t.Run("configuration supports current", func(t *testing.T) {
		config := &Configuration{}
		require.False(t, config.Status.Current)
		require.False(t, config.IsCurrent())
		config.SetCurrent(true)
		require.True(t, config.Status.Current)
		require.True(t, config.IsCurrent())
	})
	t.Run("sourceType doesn't support current", func(t *testing.T) {
		sourceType := &SourceType{}
		require.False(t, sourceType.IsCurrent())
		sourceType.SetCurrent(true)
		require.False(t, sourceType.IsCurrent())
	})
	t.Run("AnyResource uses a map", func(t *testing.T) {
		anyResource := &AnyResource{}
		require.Nil(t, anyResource.Status["current"])
		require.False(t, anyResource.IsCurrent())
		anyResource.SetCurrent(true)
		require.NotNil(t, anyResource.Status)
		require.True(t, anyResource.Status["current"].(bool))
		require.True(t, anyResource.IsCurrent())
	})
}
