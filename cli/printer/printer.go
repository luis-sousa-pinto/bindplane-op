// Copyright  observIQ, Inc.
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

// Package printer provides an interface for printing resources, and several implementations.
package printer

import (
	"io"

	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// Printer formats and prints printable resources
//
//go:generate mockery --name Printer --filename mock_printer.go --structname MockPrinter
type Printer interface {
	// PrintResource prints a generic model that implements the printable interface
	PrintResource(model.Printable)
	// PrintResources prints a list of generic models that implements the printable interface
	PrintResources([]model.Printable)
}

// PrintResource prints a single resource. It only exists to match the syntax of PrintResource.
func PrintResource(printer Printer, resource model.Printable) {
	printer.PrintResource(resource)
}

// PrintResources allows an array of something that implements model.Printable to be printed. It's a little extra
// compute in exchange for simpler code.
func PrintResources[T model.Printable](printer Printer, resources []T) {
	printables := make([]model.Printable, len(resources))
	for i, resource := range resources {
		printables[i] = resource
	}
	printer.PrintResources(printables)
}

// Build builds a printer based on the given format.
func Build(format string, writer io.Writer, logger *zap.Logger) Printer {
	switch format {
	case "json":
		return NewJSONPrinter(writer, logger)
	case "yaml":
		return NewYamlPrinter(writer, logger)
	case "csv":
		return NewCSVPrinter(writer, logger)
	default:
		return NewTablePrinter(writer)
	}
}
