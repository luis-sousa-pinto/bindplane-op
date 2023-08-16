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

package printer

import (
	"encoding/csv"
	"io"

	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// CSVPrinter logs CSV to an io.Writer
type CSVPrinter struct {
	writer *csv.Writer
	logger *zap.Logger
}

var _ Printer = (*CSVPrinter)(nil)

// NewCSVPrinter returns a new *CSVPrinter
func NewCSVPrinter(writer io.Writer, logger *zap.Logger) *CSVPrinter {
	return &CSVPrinter{
		writer: csv.NewWriter(writer),
		logger: logger,
	}
}

// PrintResource prints a generic model that implements the printable interface
func (cp *CSVPrinter) PrintResource(item model.Printable) {
	if item == nil {
		return
	}

	cp.PrintResources([]model.Printable{item})
}

// PrintResources prints a generic model that implements the model.Printable interface
func (cp *CSVPrinter) PrintResources(list []model.Printable) {
	if len(list) == 0 {
		return
	}

	titles := list[0].PrintableFieldTitles()
	fieldsAndTitles := [][]string{}
	fieldsAndTitles = append(fieldsAndTitles, titles)
	for _, item := range list {
		resource := []string{}
		for _, title := range titles {
			resource = append(resource, item.PrintableFieldValue(title))
		}
		fieldsAndTitles = append(fieldsAndTitles, resource)
	}
	err := cp.writer.WriteAll(fieldsAndTitles)
	if err != nil {
		cp.logger.Error("failed to write CSV", zap.Error(err))
	}
}
