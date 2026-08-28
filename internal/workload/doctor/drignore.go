// Copyright 2026 DataRobot, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doctor

import (
	"context"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/workload/ignore"
)

// drignoreCheck reports whether an ignore file is present at the project root.
// Presence uses ignore.Locate semantics: .drignore or legacy .wapiignore
// satisfies it. A linked project without either should be reseeded.
type drignoreCheck struct {
	projectDir string
}

func (c *drignoreCheck) ID() string {
	return CheckIDDrignore
}

func (c *drignoreCheck) Name() string {
	return "Ignore file"
}

// Run delegates to ignore.Locate, which checks .drignore first and falls back
// to .wapiignore. An empty result means neither file exists (WARN, fixable).
// The check is presence-only: it never polices the file's contents.
func (c *drignoreCheck) Run(_ context.Context) core.Result {
	if res, skip := skipIfUnlinked(c.projectDir); skip {
		return res
	}

	if ignore.Locate(c.projectDir) != "" {
		return core.Result{
			Status:  core.StatusOK,
			Summary: "ignore file present at project root",
		}
	}

	return core.Result{
		Status:  core.StatusWARN,
		Summary: "no ignore file found at project root (.drignore or .wapiignore)",
		Remedy:  RemedyDrignore,
		Fixable: true,
	}
}
