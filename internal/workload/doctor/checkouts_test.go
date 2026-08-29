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
	"github.com/datarobot/cli/internal/workload/wapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCheckoutsCheck runs only the checkouts-orphaned check.
func runCheckoutsCheck(t *testing.T, projectDir string) core.Result {
	t.Helper()

	check := &checkoutsOrphanedCheck{projectDir: projectDir}

	return check.Run(context.Background())
}

// --- VAL-EXTRA-011: .checkouts/ with snapshot dirs → WARN with disk usage ---

func TestCheckoutsOrphaned_WithSnapshots_WARN_WithDiskUsage(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	// Create two snapshot dirs with junk files.
	checkoutsDir := wapi.CheckoutsDir(dir)

	require.NoError(t, os.MkdirAll(filepath.Join(checkoutsDir, "v123"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(checkoutsDir, "v123", "app.go"), []byte("package main\n"), 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(checkoutsDir, "v456"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(checkoutsDir, "v456", "app.go"), []byte("package main\nfunc main() {}\n"), 0o644))

	result := runCheckoutsCheck(t, dir)

	assert.Equal(t, core.StatusWARN, result.Status)
	assert.Contains(t, result.Summary, "2 snapshot")
	assert.Contains(t, result.Summary, "disk", "summary must include a disk-usage figure")
	assert.False(t, result.Fixable, "checkouts-orphaned is informational, not fixable")
	assert.Equal(t, RemedyCheckoutsOrphaned, result.Remedy)

	// Read-only: the dirs are untouched.
	assert.True(t, dirExists(filepath.Join(checkoutsDir, "v123")))
	assert.True(t, dirExists(filepath.Join(checkoutsDir, "v456")))
}

// --- VAL-EXTRA-012: .checkouts/ absent → OK, not created ---

func TestCheckoutsOrphaned_Absent_OK(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	result := runCheckoutsCheck(t, dir)

	assert.Equal(t, core.StatusOK, result.Status)
	assert.Contains(t, result.Summary, "no .checkouts")

	// Read-only: .checkouts/ is NOT created.
	_, err := os.Stat(wapi.CheckoutsDir(dir))
	assert.ErrorIs(t, err, os.ErrNotExist, "doctor must not create .checkouts/")
}

func TestCheckoutsOrphaned_EmptyDir_OK(t *testing.T) {
	// An empty .checkouts/ directory (no snapshot dirs) is OK.
	dir := linkedWithoutDrignore(t)

	require.NoError(t, os.MkdirAll(wapi.CheckoutsDir(dir), 0o755))

	result := runCheckoutsCheck(t, dir)

	assert.Equal(t, core.StatusOK, result.Status)
	assert.Contains(t, result.Summary, "no .checkouts/ snapshots")
}

func TestCheckoutsOrphaned_WithFilesButNoDirs_OK(t *testing.T) {
	// .checkouts/ with loose files but no subdirs is OK (no snapshot dirs).
	dir := linkedWithoutDrignore(t)

	checkoutsDir := wapi.CheckoutsDir(dir)

	require.NoError(t, os.MkdirAll(checkoutsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(checkoutsDir, "loose.txt"), []byte("junk\n"), 0o644))

	result := runCheckoutsCheck(t, dir)

	assert.Equal(t, core.StatusOK, result.Status)
}

func TestCheckoutsOrphaned_Unlinked_SKIP(t *testing.T) {
	result := runCheckoutsCheck(t, t.TempDir())

	assert.Equal(t, core.StatusSKIP, result.Status)
	assert.Contains(t, result.Summary, "no linked state")
}

// --- Read-only guarantee ---

func TestCheckoutsOrphaned_ReadOnly_DirsUntouched(t *testing.T) {
	dir := linkedWithoutDrignore(t)

	checkoutsDir := wapi.CheckoutsDir(dir)

	require.NoError(t, os.MkdirAll(filepath.Join(checkoutsDir, "v123"), 0o755))

	content := []byte("package main\n")
	require.NoError(t, os.WriteFile(filepath.Join(checkoutsDir, "v123", "app.go"), content, 0o644))

	before := stateFileHashes(t, dir)

	runCheckoutsCheck(t, dir)

	assert.Equal(t, before, stateFileHashes(t, dir), "checkouts-orphaned must not modify anything")

	// Content is byte-identical.
	data, err := os.ReadFile(filepath.Join(checkoutsDir, "v123", "app.go"))
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

// dirExists is a small helper for test assertions.
func dirExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}
