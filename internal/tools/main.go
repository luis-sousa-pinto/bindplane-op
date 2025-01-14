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

// Package main exists to provide imports for tools used in development
package main

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/client9/misspell"
	_ "github.com/google/addlicense"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/mgechev/revive"
	_ "github.com/nbutton23/zxcvbn-go" // required by gosec
	_ "github.com/ory/go-acc"
	_ "github.com/securego/gosec/v2"
	_ "github.com/securego/gosec/v2/report/sarif" // required by gosec
	_ "github.com/securego/gosec/v2/report/text"  // required by gosec
	_ "github.com/swaggo/swag/cmd/swag"
	_ "github.com/uw-labs/lichen"
	_ "github.com/vektra/mockery/v2"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
