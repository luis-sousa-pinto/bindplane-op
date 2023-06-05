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

	"github.com/observiq/bindplane-op/model"
)

// memoryFirstResourceStore is a model.ResourceStore that first attempts to return in-memory resources, before
// falling back to an underlying model.ResourceStore.
type memoryFirstResourceStore struct {
	resourceStore model.ResourceStore
	sources map[string]*model.Source
	sourceTypes map[string]*model.SourceType
	processors map[string]*model.Processor
	processorTypes map[string]*model.ProcessorType
	destinations map[string]*model.Destination
	destinationTypes map[string]*model.DestinationType
}

// newMemoryFirstResourceStore returns a new MemoryFirstResourceStore, which first looks for the resource in
// the provided resource slice, then looks into the model.ResourceStore.
func newMemoryFirstResourceStore(resources []model.Resource, store model.ResourceStore) *memoryFirstResourceStore {
	rt := &memoryFirstResourceStore{
		resourceStore: store,
		sources: map[string]*model.Source{},
		sourceTypes: map[string]*model.SourceType{},
		processors: map[string]*model.Processor{},
		processorTypes: map[string]*model.ProcessorType{},
		destinations: map[string]*model.Destination{},
		destinationTypes: map[string]*model.DestinationType{},
	}

	for _, res := range resources {
		switch typedRes := res.(type) {
		case *model.Source:
			rt.sources[typedRes.Name()] = typedRes
		case *model.SourceType:
			rt.sourceTypes[typedRes.Name()] = typedRes
		case *model.Processor:
			rt.processors[typedRes.Name()] = typedRes
		case *model.ProcessorType:
			rt.processorTypes[typedRes.Name()] = typedRes
		case *model.Destination:
			rt.destinations[typedRes.Name()] = typedRes
		case *model.DestinationType:
			rt.destinationTypes[typedRes.Name()] = typedRes
		}
	}

	return rt
}

func (t memoryFirstResourceStore) Source(ctx context.Context, name string) (*model.Source, error) {
	if source, ok := t.sources[name]; ok {
		return source, nil
	}

	return t.resourceStore.Source(ctx, name)
}

func (t memoryFirstResourceStore) SourceType(ctx context.Context, name string) (*model.SourceType, error) {
	if sourceType, ok := t.sourceTypes[name]; ok {
		return sourceType, nil
	}
	
	return t.resourceStore.SourceType(ctx, name)
}

func (t memoryFirstResourceStore) Processor(ctx context.Context, name string) (*model.Processor, error) {
	if processors, ok := t.processors[name]; ok {
		return processors, nil
	}
	
	return t.resourceStore.Processor(ctx, name)
}

func (t memoryFirstResourceStore) ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	if processorType, ok := t.processorTypes[name]; ok {
		return processorType, nil
	}
	
	return t.resourceStore.ProcessorType(ctx, name)
}

func (t memoryFirstResourceStore) Destination(ctx context.Context, name string) (*model.Destination, error) {
	if destination, ok := t.destinations[name]; ok {
		return destination, nil
	}
	
	return t.resourceStore.Destination(ctx, name)
}

func (t memoryFirstResourceStore) DestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	if destinationType, ok := t.destinationTypes[name]; ok {
		return destinationType, nil
	}
	
	return t.resourceStore.DestinationType(ctx, name)
}
