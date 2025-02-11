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

// Package ui provides the `globals` object for the UI.
// js/_globals.js is used to fetch the server version without authentication.
package ui

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/version"
)

const templateStr = `var __BINDPLANE_VERSION__ = "{{.Version}}";
`

type configOptions struct {
	Version string
}

func newConfigOptions() *configOptions {
	return &configOptions{
		Version: version.NewVersion().String(),
	}
}

// generateGlobalJS generates the static javascript file for the UI.
func generateGlobalJS() string {
	tmp, _ := template.New("globals").Parse(templateStr)

	opts := newConfigOptions()

	w := bytes.NewBufferString("")
	_ = tmp.Execute(w, opts)

	return w.String()
}

func globalJS(ctx *gin.Context) {
	js := generateGlobalJS()

	ctx.Header("Content-Type", "application/javascript")
	ctx.String(http.StatusOK, js)
}
