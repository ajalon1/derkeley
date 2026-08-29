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

// Canonical remedy strings, one exact string per check condition. They are
// owned by this package and reused verbatim by both reporters (text and
// JSON), so the command layer must render them as-is rather than rewording.
const (
	// RemedyPresence is shown when the project has no state directory:
	// linking is the only way in.
	RemedyPresence = "dr artifact code init <artifact-id>"

	// RemedyConfig is shown when config.json is missing, corrupt, or fails
	// semantic validation. The config is the source of truth and cannot be
	// auto-rebuilt (a `--fix` manifest rebuild requires a valid config), so
	// recovery is re-initialization.
	RemedyConfig = "dr artifact code init <artifact-id>"

	// RemedyManifest is shown when manifest.json is missing, corrupt, or
	// fails semantic validation: `--fix` rebuilds an empty BASE from config.
	RemedyManifest = "dr artifact code doctor --fix (rebuilds an empty BASE from config)"

	// RemedyDivergence is shown when config lastSyncedVersionId and manifest
	// syncedVersionId disagree: `--fix` resets the manifest from config.
	RemedyDivergence = "dr artifact code doctor --fix (resets the manifest from config)"

	// RemedyRollback is shown when an interrupted rollback tree exists:
	// `--fix` restores the backed-up files and clears the tree.
	RemedyRollback = "dr artifact code doctor --fix (restores backed-up files and clears .rollback/)"

	// RemedyLockHeld is shown when a live process holds sync.lock. Quitting
	// the holder (never removing the lock file) is the only safe recovery.
	RemedyLockHeld = "identify and quit the process holding the sync lock (e.g. another 'dr artifact code sync'), then re-run this command"

	// RemedyLockInspect is shown when the lock file cannot be inspected
	// (permission or I/O error). This is NOT a held-lock condition.
	RemedyLockInspect = "check permissions on the sync state directory and sync.lock, then re-run this command"

	// RemedyRelink is shown when the linked artifact is gone (404): the only
	// recovery is repointing the project at a new artifact with a fresh BASE.
	RemedyRelink = "dr artifact code doctor --relink <new-artifact-id>"

	// RemedyArtifactLocked is shown when the linked artifact is locked
	// (locking is one-way). Sync execution is refused but preview works; work
	// against a draft instead, or relink to one.
	RemedyArtifactLocked = "work against a draft artifact, or relink to one: dr artifact code doctor --relink <new-artifact-id>"

	// RemedyCatalogMismatch is shown when the locally pinned catalog id no
	// longer matches the artifact's codeRef: the pin is stale server-side.
	RemedyCatalogMismatch = "dr artifact code doctor --relink <new-artifact-id> (or re-init against the intended artifact)"

	// RemedyDrift is shown when the artifact's codeRef version no longer
	// matches the last-synced version. Review what a sync would do first;
	// relink starts a fresh baseline instead.
	RemedyDrift = "review with 'dr artifact code sync --dry-run', or relink to start fresh: dr artifact code doctor --relink <new-artifact-id>"

	// RemedyRemoteConnectivity is shown when a remote check could not reach
	// the API for ANY reason other than a 404 (unauthenticated, 401/403,
	// 5xx, timeout, unreachable endpoint). It deliberately never mentions
	// --relink: a fetch failure is not evidence that the artifact is gone.
	RemedyRemoteConnectivity = "run 'dr auth login' or fix network connectivity, then re-run this command"

	// RemedyLegacyUnmigrated is shown when a legacy .wapi/ state directory
	// exists but the current .datarobot/workload/ does not: --fix migrates
	// the state in place, preserving all contents.
	RemedyLegacyUnmigrated = "dr artifact code doctor --fix (migrates .wapi/ to .datarobot/workload/ preserving contents)"

	// RemedyDrignore is shown when no ignore file is found at the project
	// root (.drignore or legacy .wapiignore): --fix reseeds .drignore from
	// the standard template. An existing file is never overwritten.
	RemedyDrignore = "dr artifact code doctor --fix (reseeds .drignore from the standard template)"

	// RemedyHistory is shown when history.log contains unparseable JSONL
	// lines (crash truncation). It is informational only: the doctor never
	// writes history.log, so the user must review and fix it manually.
	RemedyHistory = "review history.log and manually fix unparseable lines (doctor does not auto-repair history)"

	// RemedyHistoryUnreadable is shown when history.log cannot be read at
	// all (permission or I/O error). This is not an unparseable-lines
	// condition — nothing can be parsed from a file that cannot be opened —
	// so access must be restored first. The doctor never writes history.log,
	// so making the file inspectable is on the user.
	RemedyHistoryUnreadable = "check permissions and ownership on the state directory and history.log, then re-run this command"

	// RemedyNoCodeRef is shown when the linked artifact exists but has no
	// usable code reference (never synced from this project). It is distinct
	// from drift: the artifact simply has no codeRef yet. Running a sync
	// establishes it and clears the WARN.
	RemedyNoCodeRef = "run 'dr artifact code sync' to establish the code reference (the artifact has never been synced from this project)"

	// RemedyCheckoutsOrphaned is shown when .checkouts/ snapshot directories
	// are present. It is informational only: the doctor never deletes
	// .checkouts/ snapshots, so the user must review and remove them manually
	// if no longer needed.
	RemedyCheckoutsOrphaned = "review .checkouts/ snapshots and remove them manually if no longer needed (doctor does not auto-clean .checkouts/)"

	// RemedyCheckoutsUnreadable is shown when .checkouts/ exists but cannot
	// be read (permission or I/O error). Whether snapshots are present is
	// unverifiable from here, so the check SKIPs rather than assuming the
	// directory is healthy; access must be restored before it can run.
	RemedyCheckoutsUnreadable = "check permissions and ownership on the state directory and .checkouts/, then re-run this command"
)
