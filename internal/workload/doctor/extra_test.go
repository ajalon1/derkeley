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

// runExtraChecks runs the extra check suite through the framework Runner and
// returns the results in the fixed check order.
func runExtraChecks(t *testing.T, projectDir string) []core.Result {
	t.Helper()

	return core.NewRunner(ExtraChecks(projectDir, nil)...).Run(context.Background())
}

// legacyOnlyProject writes a valid state dir at the legacy .wapi/ location
// with no .datarobot/workload/ present.
func legacyOnlyProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	legacy := filepath.Join(dir, wapi.LegacyDirName)

	require.NoError(t, os.MkdirAll(legacy, 0o755))

	require.NoError(t, os.WriteFile(
		filepath.Join(legacy, "config.json"),
		[]byte(validConfigJSON()), 0o644,
	))

	require.NoError(t, os.WriteFile(
		filepath.Join(legacy, "manifest.json"),
		[]byte(`{"version":1,"files":{}}`), 0o644,
	))

	return dir
}

// linkedWithDrignore creates a linked project with a .drignore at the root.
func linkedWithDrignore(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	require.NoError(t, wapi.SaveManifest(dir, validManifest("")))

	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ignore.FileName), wapi.IgnoreTemplate(), 0o644,
	))

	return dir
}

// linkedWithoutDrignore creates a linked project with no ignore file at the root.
func linkedWithoutDrignore(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	require.NoError(t, wapi.SaveManifest(dir, validManifest("")))

	return dir
}

// writeHistoryLog writes the given raw content to history.log inside the state dir.
func writeHistoryLog(t *testing.T, projectDir, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(
		filepath.Join(wapi.Dir(projectDir), wapi.HistoryFile),
		[]byte(content), 0o644,
	))
}

// --- ExtraChecks fixed order ---

func TestExtraChecks_FixedOrder(t *testing.T) {
	ids := make([]string, 0, 5)

	for _, c := range ExtraChecks(t.TempDir(), nil) {
		ids = append(ids, c.ID())
	}

	assert.Equal(t, []string{
		CheckIDLegacyUnmigrated,
		CheckIDDrignore,
		CheckIDHistory,
		CheckIDNoCodeRef,
		CheckIDCheckoutsOrphaned,
	}, ids)
}

// --- wapi.legacy-unmigrated (VAL-EXTRA-001) ---

func TestLegacyUnmigrated_LegacyOnly_WARN(t *testing.T) {
	dir := legacyOnlyProject(t)

	results := runExtraChecks(t, dir)

	require.Len(t, results, 5)

	assert.Equal(t, core.StatusWARN, results[0].Status)
	assert.Contains(t, results[0].Summary, "legacy")
	assert.Contains(t, results[0].Summary, ".wapi")
	assert.Equal(t, RemedyLegacyUnmigrated, results[0].Remedy)
	assert.True(t, results[0].Fixable, "legacy migration is fixable")

	// Read-only: no migration occurs.
	assert.True(t, fsutil.DirExists(filepath.Join(dir, wapi.LegacyDirName)),
		"legacy dir must survive a read-only run")
	assert.False(t, fsutil.DirExists(filepath.Join(dir, wapi.RootDirName, wapi.StateDirName)),
		"current dir must not be created during a read-only run")
}

func TestLegacyUnmigrated_CurrentExists_OK(t *testing.T) {
	dir := healthyProject(t)

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[0].Status)
	assert.Contains(t, results[0].Summary, "current location")
}

func TestLegacyUnmigrated_Unlinked_SKIP(t *testing.T) {
	results := runExtraChecks(t, t.TempDir())

	assert.Equal(t, core.StatusSKIP, results[0].Status)
	assert.Contains(t, results[0].Summary, "no linked state")
}

// --- wapi.drignore (VAL-EXTRA-003) ---

func TestDrignore_Missing_WARN(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusWARN, results[1].Status)
	assert.Contains(t, results[1].Summary, "no ignore file")
	assert.Equal(t, RemedyDrignore, results[1].Remedy)
	assert.True(t, results[1].Fixable, "drignore reseed is fixable")

	// Read-only: no file created.
	_, err := os.Stat(filepath.Join(dir, ignore.FileName))
	assert.ErrorIs(t, err, os.ErrNotExist, "read-only run must not create .drignore")
}

func TestDrignore_Present_OK(t *testing.T) {
	dir := linkedWithDrignore(t)

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[1].Status)
}

func TestDrignore_LegacyWapiignore_OK(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// A legacy .wapiignore satisfies presence per ignore.Locate semantics.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ignore.LegacyFileName), []byte("*.tmp\n"), 0o644,
	))

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[1].Status, "legacy .wapiignore satisfies presence")
}

func TestDrignore_Unlinked_SKIP(t *testing.T) {
	results := runExtraChecks(t, t.TempDir())

	assert.Equal(t, core.StatusSKIP, results[1].Status)
}

// --- wapi.history (VAL-EXTRA-006/007/008) ---

func TestHistory_ValidLog_OK(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	writeHistoryLog(t, dir, `{"ts":"2026-01-01T00:00:00Z","op":"init","artifact":"abc123"}`+"\n"+
		`{"ts":"2026-01-02T00:00:00Z","op":"sync","duration":"1.0s"}`+"\n")

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[2].Status)
	assert.Contains(t, results[2].Summary, "valid")
}

func TestHistory_MissingLog_OK(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[2].Status)
	assert.Contains(t, results[2].Summary, "no history log")

	// Read-only: history.log is NOT created.
	_, err := os.Stat(filepath.Join(wapi.Dir(dir), wapi.HistoryFile))
	assert.ErrorIs(t, err, os.ErrNotExist, "doctor must not create history.log")
}

func TestHistory_CorruptLog_WARN_NotFixable(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// One valid line then a truncated/garbage line.
	writeHistoryLog(t, dir, `{"ts":"2026-01-01T00:00:00Z","op":"init"}`+"\n"+
		`{"ts":"2026-01-02T00:00:00Z","op":"syn`)

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusWARN, results[2].Status)
	assert.Contains(t, results[2].Summary, "unparseable")
	assert.False(t, results[2].Fixable, "history corruption is informational, not fixable")
	assert.NotContains(t, results[2].Remedy, "--fix",
		"history remedy must not mention --fix")

	// File is untouched (mtime/hash unchanged — verified by content still corrupt).
	data, err := os.ReadFile(filepath.Join(wapi.Dir(dir), wapi.HistoryFile))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"op":"syn`,
		"corrupt line must survive the read-only run byte-for-byte")
}

func TestHistory_AllGarbage_WARN(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	writeHistoryLog(t, dir, "this is not json\nneither is this\n")

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusWARN, results[2].Status)
	assert.Contains(t, results[2].Summary, "2 unparseable")
	assert.False(t, results[2].Fixable)
}

func TestHistory_EmptyFile_OK(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	writeHistoryLog(t, dir, "")

	results := runExtraChecks(t, dir)

	assert.Equal(t, core.StatusOK, results[2].Status)
}

func TestHistory_Unlinked_SKIP(t *testing.T) {
	results := runExtraChecks(t, t.TempDir())

	assert.Equal(t, core.StatusSKIP, results[2].Status)
}

// --- Full Checks() order (VAL-EXTRA: checks after remote block) ---

func TestChecks_FullOrder_WithExtras(t *testing.T) {
	ids := make([]string, 0, 15)

	for _, c := range Checks(t.TempDir(), &fakeArtifactStore{}) {
		ids = append(ids, c.ID())
	}

	assert.Equal(t, []string{
		// 6 local
		CheckIDPresence,
		CheckIDConfig,
		CheckIDManifest,
		CheckIDDivergence,
		CheckIDRollback,
		CheckIDLock,
		// 4 remote
		CheckIDArtifactExists,
		CheckIDArtifactLocked,
		CheckIDCatalogMismatch,
		CheckIDDrift,
		// 5 M4 extras
		CheckIDLegacyUnmigrated,
		CheckIDDrignore,
		CheckIDHistory,
		CheckIDNoCodeRef,
		CheckIDCheckoutsOrphaned,
	}, ids)
}

// --- Read-only guarantee for extra checks ---

func TestExtraChecks_ReadOnly_ZeroWrites(t *testing.T) {
	dir := legacyOnlyProject(t)

	// Add a history.log with valid content.
	writeHistoryLog(t, dir, `{"ts":"2026-01-01T00:00:00Z","op":"init"}`+"\n")

	before := stateFileHashes(t, dir)

	runExtraChecks(t, dir)

	assert.Equal(t, before, stateFileHashes(t, dir),
		"extra checks must not write anything in read-only mode")
}

// --- Extra checks cascade with presence FAIL ---

func TestExtraChecks_PresenceFail_AllSkip(t *testing.T) {
	results := runExtraChecks(t, t.TempDir())

	for _, res := range results {
		assert.Equal(t, core.StatusSKIP, res.Status, "check %s should SKIP", res.CheckID)
		assert.Contains(t, res.Summary, "no linked state")
	}
}
