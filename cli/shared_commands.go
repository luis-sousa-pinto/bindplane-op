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

package cli

import (
	"github.com/observiq/bindplane-op/cli/commands/apply"
	"github.com/observiq/bindplane-op/cli/commands/delete"
	"github.com/observiq/bindplane-op/cli/commands/get"
	"github.com/observiq/bindplane-op/cli/commands/initialize"
	"github.com/observiq/bindplane-op/cli/commands/install"
	"github.com/observiq/bindplane-op/cli/commands/label"
	"github.com/observiq/bindplane-op/cli/commands/profile"
	"github.com/observiq/bindplane-op/cli/commands/rollout"
	"github.com/observiq/bindplane-op/cli/commands/sync"
	"github.com/observiq/bindplane-op/cli/commands/update"
	"github.com/observiq/bindplane-op/cli/commands/version"
	"github.com/spf13/cobra"
)

// SharedCommands returns a list of commands that should be shared between server and ctl
func SharedCommands(factory *Factory) []*cobra.Command {
	return []*cobra.Command{
		AddPrerunsToExistingCmd(apply.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(get.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(label.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(delete.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(profile.Command(factory), factory, AddLoadConfigPrerun),
		AddPrerunsToExistingCmd(version.Command(factory), factory, AddLoadConfigPrerun),
		AddPrerunsToExistingCmd(initialize.Command(factory), factory, AddLoadConfigPrerun),
		AddPrerunsToExistingCmd(install.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(sync.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(update.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
		AddPrerunsToExistingCmd(rollout.Command(factory), factory, AddLoadConfigPrerun, AddValidationPrerun),
	}
}
