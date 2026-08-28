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
	"path/filepath"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/fsutil"
	"github.com/datarobot/cli/internal/workload/wapi"
)

// legacyUnmigratedCheck reports whether the project's state directory is still
// at the legacy .wapi/ location with no .datarobot/workload/ present. A
// legacy-only project works (wapi.Dir resolves there) but should be migrated
// so the state lives at the canonical location.
type legacyUnmigratedCheck struct {
	projectDir string
}

func (c *legacyUnmigratedCheck) ID() string {
	return CheckIDLegacyUnmigrated
}

func (c *legacyUnmigratedCheck) Name() string {
	return "Legacy state migration"
}

// Run checks both locations independently. If neither exists the project is
// not linked (SKIP via skipIfUnlinked). If the current directory exists the
// state is already at the canonical location (OK). If only the legacy
// directory exists the project should be migrated (WARN, fixable).
func (c *legacyUnmigratedCheck) Run(_ context.Context) core.Result {
	if res, skip := skipIfUnlinked(c.projectDir); skip {
		return res
	}

	legacy := filepath.Join(c.projectDir, wapi.LegacyDirName)
	current := filepath.Join(c.projectDir, wapi.RootDirName, wapi.StateDirName)

	if fsutil.DirExists(current) {
		return core.Result{
			Status:  core.StatusOK,
			Summary: "state directory is at the current location (.datarobot/workload/)",
		}
	}

	if !fsutil.DirExists(legacy) {
		// Should not happen: skipIfUnlinked passed, so one of the two must
		// exist. Treat as OK rather than crashing.
		return core.Result{
			Status:  core.StatusOK,
			Summary: "state directory found",
		}
	}

	return core.Result{
		Status:  core.StatusWARN,
		Summary: "legacy .wapi/ state directory exists but .datarobot/workload/ does not",
		Remedy:  RemedyLegacyUnmigrated,
		Fixable: true,
	}
}
