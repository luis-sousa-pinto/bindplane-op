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

package rest

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestMemoryFirstResourceStore(t *testing.T) {
	t.Run("Source", func(t *testing.T) {
		testSource1 := model.NewSource("testSource1", "source-type", []model.Parameter{})
		testSource2 := model.NewSource("testSource2", "source-type", []model.Parameter{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("Source", ctx, "testSource2").Return(testSource2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testSource1}, mockResStore)
		source, err := mfrs.Source(ctx, "testSource1")
		require.NoError(t, err)
		require.Equal(t, source, testSource1)

		source2, err := mfrs.Source(ctx, "testSource2")
		require.NoError(t, err)
		require.Equal(t, source2, testSource2)

		mockResStore.AssertExpectations(t)
	})

	t.Run("SourceType", func(t *testing.T) {
		testSourceType1 := model.NewSourceType("testSourceType1", []model.ParameterDefinition{}, []string{})
		testSourceType2 := model.NewSourceType("testSourceType2", []model.ParameterDefinition{}, []string{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("SourceType", ctx, "testSourceType2").Return(testSourceType2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testSourceType1}, mockResStore)
		sourceType, err := mfrs.SourceType(ctx, "testSourceType1")
		require.NoError(t, err)
		require.Equal(t, sourceType, testSourceType1)

		sourceType2, err := mfrs.SourceType(ctx, "testSourceType2")
		require.NoError(t, err)
		require.Equal(t, sourceType2, testSourceType2)

		mockResStore.AssertExpectations(t)
	})

	t.Run("Processor", func(t *testing.T) {
		testProcessor1 := model.NewProcessor("testProcessor1", "processor-type", []model.Parameter{})
		testProcessor2 := model.NewProcessor("testProcessor2", "processor-type", []model.Parameter{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("Processor", ctx, "testProcessor2").Return(testProcessor2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testProcessor1}, mockResStore)
		processor, err := mfrs.Processor(ctx, "testProcessor1")
		require.NoError(t, err)
		require.Equal(t, processor, testProcessor1)

		processor2, err := mfrs.Processor(ctx, "testProcessor2")
		require.NoError(t, err)
		require.Equal(t, processor2, testProcessor2)

		mockResStore.AssertExpectations(t)
	})

	t.Run("ProcessorType", func(t *testing.T) {
		testProcessorType1 := model.NewProcessorType("testProcessorType1", []model.ParameterDefinition{})
		testProcessorType2 := model.NewProcessorType("testProcessorType2", []model.ParameterDefinition{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("ProcessorType", ctx, "testProcessorType2").Return(testProcessorType2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testProcessorType1}, mockResStore)
		processorType, err := mfrs.ProcessorType(ctx, "testProcessorType1")
		require.NoError(t, err)
		require.Equal(t, processorType, testProcessorType1)

		processorType2, err := mfrs.ProcessorType(ctx, "testProcessorType2")
		require.NoError(t, err)
		require.Equal(t, processorType2, testProcessorType2)

		mockResStore.AssertExpectations(t)
	})

	t.Run("Destination", func(t *testing.T) {
		testDestination1 := model.NewDestination("testDestination1", "destination-type", []model.Parameter{})
		testDestination2 := model.NewDestination("testDestination2", "destination-type", []model.Parameter{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("Destination", ctx, "testDestination2").Return(testDestination2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testDestination1}, mockResStore)
		destination, err := mfrs.Destination(ctx, "testDestination1")
		require.NoError(t, err)
		require.Equal(t, destination, testDestination1)

		destination2, err := mfrs.Destination(ctx, "testDestination2")
		require.NoError(t, err)
		require.Equal(t, destination2, testDestination2)

		mockResStore.AssertExpectations(t)
	})

	t.Run("DestinationType", func(t *testing.T) {
		testProcessorType1 := model.NewDestinationType("testDestinationType1", []model.ParameterDefinition{})
		testProcessorType2 := model.NewDestinationType("testDestinationType2", []model.ParameterDefinition{})
		ctx := context.Background()

		mockResStore := model.NewMockResourceStore(t)
		mockResStore.On("DestinationType", ctx, "testDestinationType2").Return(testProcessorType2, nil)

		mfrs := newMemoryFirstResourceStore([]model.Resource{testProcessorType1}, mockResStore)
		destinationType, err := mfrs.DestinationType(ctx, "testDestinationType1")
		require.NoError(t, err)
		require.Equal(t, destinationType, testProcessorType1)

		destinationType2, err := mfrs.DestinationType(ctx, "testDestinationType2")
		require.NoError(t, err)
		require.Equal(t, destinationType2, testProcessorType2)

		mockResStore.AssertExpectations(t)
	})
}
