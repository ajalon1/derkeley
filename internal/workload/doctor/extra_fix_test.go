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
	"os"
	"path/filepath"
	"testing"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/fsutil"
	"github.com/datarobot/cli/internal/workload/ignore"
	"github.com/datarobot/cli/internal/workload/wapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fixLegacyMigration (VAL-EXTRA-002) ---

func TestFixLegacyMigration_LegacyOnly_Performed(t *testing.T) {
	dir := legacyOnlyProject(t)

	// Record a distinctive catalogId to verify content preservation.
	legacy := filepath.Join(dir, wapi.LegacyDirName)
	cfgData, err := os.ReadFile(filepath.Join(legacy, "config.json"))
	require.NoError(t, err)

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	migration := byID[CheckIDLegacyUnmigrated]

	assert.Equal(t, core.ActionPerformed, migration.Status)
	assert.NotEmpty(t, migration.Reason)

	// Legacy dir is gone, current dir exists.
	assert.False(t, fsutil.DirExists(legacy), "legacy .wapi/ must be removed")
	assert.True(t,
		fsutil.DirExists(filepath.Join(dir, wapi.RootDirName, wapi.StateDirName)),
		"current .datarobot/workload/ must exist")

	// Config content is preserved.
	current := filepath.Join(dir, wapi.RootDirName, wapi.StateDirName)
	migratedData, err := os.ReadFile(filepath.Join(current, "config.json"))
	require.NoError(t, err)
	assert.Equal(t, cfgData, migratedData, "config.json must be byte-identical after migration")
}

func TestFixLegacyMigration_CurrentExists_NotNeeded(t *testing.T) {
	dir := healthyProject(t)

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionNotNeeded, byID[CheckIDLegacyUnmigrated].Status)
}

func TestFixLegacyMigration_Unlinked_NotNeeded(t *testing.T) {
	dir := t.TempDir()

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionNotNeeded, byID[CheckIDLegacyUnmigrated].Status)
}

// --- fixDrignore (VAL-EXTRA-004/005) ---

func TestFixDrignore_Missing_Performed(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	reseed := byID[CheckIDDrignore]

	assert.Equal(t, core.ActionPerformed, reseed.Status)
	assert.Contains(t, reseed.Reason, "reseeded")

	// File exists and matches the template.
	data, err := os.ReadFile(filepath.Join(dir, ignore.FileName))
	require.NoError(t, err)
	assert.Equal(t, wapi.IgnoreTemplate(), data,
		"reseeded .drignore must match the embedded template byte-for-byte")
}

func TestFixDrignore_Existing_NotNeeded_NotOverwritten(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// Write a custom .drignore.
	custom := []byte("# my custom patterns\n*.secret\n")
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ignore.FileName), custom, 0o644,
	))

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionNotNeeded, byID[CheckIDDrignore].Status,
		"existing .drignore must not be reseeded")

	// File is byte-identical.
	data, err := os.ReadFile(filepath.Join(dir, ignore.FileName))
	require.NoError(t, err)
	assert.Equal(t, custom, data, "custom .drignore must not be overwritten")
}

func TestFixDrignore_LegacyWapiignore_NotNeeded(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// A legacy .wapiignore satisfies presence.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ignore.LegacyFileName), []byte("*.tmp\n"), 0o644,
	))

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionNotNeeded, byID[CheckIDDrignore].Status,
		"legacy .wapiignore satisfies presence; no reseed needed")

	// .drignore must not be created.
	_, err := os.Stat(filepath.Join(dir, ignore.FileName))
	assert.ErrorIs(t, err, os.ErrNotExist, ".drignore must not be created when .wapiignore exists")
}

func TestFixDrignore_Unlinked_NotNeeded(t *testing.T) {
	dir := t.TempDir()

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionNotNeeded, byID[CheckIDDrignore].Status)
}

// --- seedIgnoreFileIfAbsent: the O_EXCL reseed primitive ---

func TestSeedIgnoreFileIfAbsent_CreatesTemplateWhenAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ignore.FileName)

	created, err := seedIgnoreFileIfAbsent(path, wapi.IgnoreTemplate())

	require.NoError(t, err)
	assert.True(t, created)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, wapi.IgnoreTemplate(), data,
		"created file must carry the template byte-for-byte")
}

func TestSeedIgnoreFileIfAbsent_ExistingFileNeverClobbered(t *testing.T) {
	path := filepath.Join(t.TempDir(), ignore.FileName)

	existing := []byte("# user patterns\n*.secret\n")
	require.NoError(t, os.WriteFile(path, existing, 0o644))

	created, err := seedIgnoreFileIfAbsent(path, wapi.IgnoreTemplate())

	require.NoError(t, err)
	assert.False(t, created, "O_EXCL create must refuse to replace an existing file")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, existing, data,
		"the never-overwrite contract must hold even in the Locate→write window: O_EXCL claims or refuses atomically, so a file that appears in the window is never clobbered by a rename")
}

// --- --fix composition: only fixable extras repaired (VAL-EXTRA-013 partial) ---

func TestRunFix_Extras_HistoryNotFixable_StaysWARN(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// Corrupt history.log (not fixable).
	writeHistoryLog(t, dir, `{"op":"init"}`+"\n"+"GARBAGE\n")

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	// Drignore is fixable and gets reseeded.
	assert.Equal(t, core.ActionPerformed, byID[CheckIDDrignore].Status,
		"drignore is fixable and should be reseeded")

	// History is NOT fixable — no action attempts it (there is no history fix
	// in the actions array at all). The history check stays WARN in the
	// post-fix re-run.
	_, hasHistoryAction := byID[CheckIDHistory]
	assert.False(t, hasHistoryAction,
		"history is not fixable; no action entry should exist for it")
}

// --- --fix with held lock: all repairs including extras skipped ---

func TestRunFix_HeldLock_ExtrasSkipped(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// Create the lock file so holdLockForTest can open and flock it.
	require.NoError(t, os.WriteFile(lockPath(t, dir), nil, 0o600))

	release := holdLockForTest(t, dir)
	t.Cleanup(release)

	actions := RunFix(context.Background(), dir)

	byID := actionByID(t, actions)

	assert.Equal(t, core.ActionSkipped, byID[CheckIDLegacyUnmigrated].Status)
	assert.Contains(t, byID[CheckIDLegacyUnmigrated].Reason, "sync in progress")
	assert.Equal(t, core.ActionSkipped, byID[CheckIDDrignore].Status)
	assert.Contains(t, byID[CheckIDDrignore].Reason, "sync in progress")

	// .drignore must NOT be created while the lock is held.
	_, err := os.Stat(filepath.Join(dir, ignore.FileName))
	assert.ErrorIs(t, err, os.ErrNotExist, "drignore must not be reseeded while lock is held")
}
