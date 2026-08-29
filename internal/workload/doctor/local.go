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
	"errors"
	"fmt"
	"path/filepath"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/workload/wapi"
)

// Stable check identifiers. They surface in reports and JSON output, so they
// must not change between releases.
const (
	CheckIDPresence   = "wapi.presence"
	CheckIDConfig     = "wapi.config"
	CheckIDManifest   = "wapi.manifest"
	CheckIDDivergence = "wapi.config-manifest-divergence"
	CheckIDRollback   = "wapi.rollback"
	CheckIDLock       = "wapi.lock"

	// M4 extra checks — appended after the remote block in the pinned order.
	CheckIDLegacyUnmigrated  = "wapi.legacy-unmigrated"
	CheckIDDrignore          = "wapi.drignore"
	CheckIDHistory           = "wapi.history"
	CheckIDNoCodeRef         = "remote.no-coderef"
	CheckIDCheckoutsOrphaned = "wapi.checkouts-orphaned"
)

// Checks returns the complete doctor check suite in the pinned fixed order:
// the six local checks, the four remote checks, then the five M4 extra
// checks appended after the remote block. The remote checks and the
// remote.no-coderef extra share one artifact snapshot fetched through store
// (exactly one GetArtifact per run).
//
// Each check resolves projectDir independently at Run time, so the returned
// checks stay correct even if the directory's state changes between
// construction and execution.
func Checks(projectDir string, store ArtifactGetter) []core.Check {
	snapshot := &remoteSnapshot{store: store}

	checks := append(LocalChecks(projectDir), remoteChecksWithSnapshot(projectDir, snapshot)...)

	return append(checks, ExtraChecks(projectDir, snapshot)...)
}

// ExtraChecks returns the five M4 extra checks appended after the remote
// block in the pinned fixed order: legacy-unmigrated, drignore, history,
// remote.no-coderef, wapi.checkouts-orphaned. The remote.no-coderef check
// shares the same artifact snapshot as the ticket-scope remote checks (one
// fetch per run); pass nil when no remote checks are needed (local-only test
// runs — the no-coderef check will SKIP).
//
// Each check resolves projectDir independently at Run time, so the returned
// checks stay correct even if the directory's state changes between
// construction and execution.
func ExtraChecks(projectDir string, snapshot *remoteSnapshot) []core.Check {
	// A nil snapshot (local-only test runs) is replaced with an empty one
	// whose nil store makes the no-coderef check SKIP cleanly.
	if snapshot == nil {
		snapshot = &remoteSnapshot{}
	}

	return []core.Check{
		&legacyUnmigratedCheck{projectDir: projectDir},
		&drignoreCheck{projectDir: projectDir},
		&historyCheck{projectDir: projectDir},
		&noCodeRefCheck{remoteBase{projectDir: projectDir, snapshot: snapshot}},
		&checkoutsOrphanedCheck{projectDir: projectDir},
	}
}

// LocalChecks returns the six local sync-state checks in the fixed report
// order: presence, config, manifest, divergence, rollback, lock. The remote
// checks (defined by their own feature) append after these.
//
// Each check resolves projectDir independently at Run time, so the returned
// checks stay correct even if the directory's state changes between
// construction and execution.
func LocalChecks(projectDir string) []core.Check {
	return []core.Check{
		&presenceCheck{projectDir: projectDir},
		&configCheck{projectDir: projectDir},
		&manifestCheck{projectDir: projectDir},
		&divergenceCheck{projectDir: projectDir},
		&rollbackCheck{projectDir: projectDir},
		newLockCheck(projectDir),
	}
}

// skipIfUnlinked implements the presence-FAIL cascade: a check that needs
// linked state SKIPs with an honest "no linked state" summary rather than
// reporting a misleading FAIL of its own. The second return value reports
// whether the caller should skip.
func skipIfUnlinked(projectDir string) (core.Result, bool) {
	if wapi.Exists(projectDir) {
		return core.Result{}, false
	}

	return core.Result{
		Status:  core.StatusSKIP,
		Summary: "no linked state; nothing to check",
	}, true
}

// corruptFileResult builds a FAIL result for a missing or unreadable state
// file. The absolute file path appears both in the human-readable summary
// and in details.path for JSON consumers.
func corruptFileResult(summary, path, remedy string, fixable bool) core.Result {
	abs := absPath(path)

	return core.Result{
		Status:  core.StatusFAIL,
		Summary: fmt.Sprintf("%s (%s)", summary, abs),
		Remedy:  remedy,
		Details: map[string]string{"path": abs},
		Fixable: fixable,
	}
}

// stateErrPath extracts the file path carried by a wapi.CorruptedError,
// falling back to fallbackPath when err is not one.
func stateErrPath(err error, fallbackPath string) string {
	var corruptErr *wapi.CorruptedError

	if errors.As(err, &corruptErr) {
		return corruptErr.Path
	}

	return fallbackPath
}

// corruptReason returns the most specific message for a corrupted state
// file: the underlying cause of a wapi.CorruptedError (whose own Error text
// already embeds the path, which corruptFileResult appends once), or the
// error itself for anything else.
func corruptReason(err error) string {
	var corruptErr *wapi.CorruptedError

	if errors.As(err, &corruptErr) && corruptErr.Err != nil {
		return corruptErr.Err.Error()
	}

	return err.Error()
}

// absPath converts p to its absolute form. On the (non-representable) error
// path it returns p unchanged: callers already pass an absolute project dir
// in normal wiring, where the command resolves --dir with filepath.Abs.
func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}

	return abs
}

// normalizeStringPtr treats an empty string as absent, so a pointer to ""
// compares equal to nil everywhere the doctor reasons about config/manifest
// pointer fields (the "empty ≈ nil" normalization pinned for all checks).
func normalizeStringPtr(p *string) *string {
	if p == nil || *p == "" {
		return nil
	}

	return p
}

// ptrDisplay renders an optional pointer for human/JSON summaries: the value,
// or "null" when absent (after empty-string normalization).
func ptrDisplay(p *string) string {
	if normalizeStringPtr(p) == nil {
		return "null"
	}

	return *p
}
