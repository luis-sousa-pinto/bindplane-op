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

// Package version contains the version logic for models in BindPlane
package version

const (
	// V1Alpha is the version for resources during the alpha of BindPlane. There were no breaking changes made before V1,
	// so the V1 models will be used.
	V1Alpha = "bindplane.observiq.com/v1alpha"

	// V1Beta is the version for resources during the beta of BindPlane. There were no breaking changes made before V1,
	// so the V1 models will be used.
	V1Beta = "bindplane.observiq.com/v1beta"

	// V1 is the version for the initial resources defined for BindPlane
	V1 = "bindplane.observiq.com/v1"
)
