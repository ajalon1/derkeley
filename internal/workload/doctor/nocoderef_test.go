// Copyright 2026 DataRobot, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed under an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doctor

import (
	"context"
	"net/http"
	"testing"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/drapi"
	"github.com/datarobot/cli/internal/workload"
	"github.com/datarobot/cli/internal/workload/wapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runNoCodeRefCheck runs only the no-coderef check against the fixture store.
func runNoCodeRefCheck(t *testing.T, projectDir string, store ArtifactGetter) core.Result {
	t.Helper()

	snapshot := &remoteSnapshot{store: store}
	check := &noCodeRefCheck{remoteBase{projectDir: projectDir, snapshot: snapshot}}

	return check.Run(context.Background())
}

// --- VAL-EXTRA-009: remote.no-coderef WARN distinct from drift ---

func TestNoCodeRef_NeverSyncedNilCodeRef_WARN(t *testing.T) {
	// Fresh draft: no codeRef at all (ExtractCodeRef returns nil). The
	// no-coderef check must WARN, and the summary must NOT mention drift.
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	store := &fakeArtifactStore{artifact: testArtifact("draft", nil)}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusWARN, result.Status)
	assert.Contains(t, result.Summary, "no usable code reference")
	assert.Contains(t, result.Summary, "never synced")
	assert.NotContains(t, result.Summary, "drift", "no-coderef summary must not conflate with drift")
	assert.NotContains(t, result.Summary, "CatalogVersionID mismatch")
	assert.False(t, result.Fixable, "no-coderef is not fixable by --fix")
	assert.Equal(t, RemedyNoCodeRef, result.Remedy)
	assert.Contains(t, result.Remedy, "sync")
}

func TestNoCodeRef_BothEmptyFieldsCodeRef_WARN(t *testing.T) {
	// ExtractCodeRef may return a non-nil pointer with empty fields; both
	// empty = no usable codeRef → same WARN as nil.
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	store := &fakeArtifactStore{artifact: testArtifact("draft", &workload.DatarobotCodeRef{
		CatalogID:        "",
		CatalogVersionID: "",
	})}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusWARN, result.Status)
	assert.Contains(t, result.Summary, "no usable code reference")
	assert.False(t, result.Fixable)
}

func TestNoCodeRef_WithCodeRef_OK(t *testing.T) {
	// After a real sync, the artifact has a populated codeRef → OK.
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig(testCatalogID, testVersionID)))

	store := &fakeArtifactStore{artifact: testArtifact("draft", &workload.DatarobotCodeRef{
		CatalogID:        testCatalogID,
		CatalogVersionID: testVersionID,
	})}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusOK, result.Status)
	assert.Contains(t, result.Summary, "usable code reference")
}

func TestNoCodeRef_PartialCodeRef_OK(t *testing.T) {
	// A codeRef with only CatalogID set (CatalogVersionID empty) is still
	// "usable" — both-empty is the no-coderef threshold, not either-empty.
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig(testCatalogID, "")))

	store := &fakeArtifactStore{artifact: testArtifact("draft", &workload.DatarobotCodeRef{
		CatalogID:        testCatalogID,
		CatalogVersionID: "",
	})}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusOK, result.Status, "partial codeRef (catalog set, version empty) is usable")
}

// --- SKIP cascades ---

func TestNoCodeRef_Unlinked_SKIP(t *testing.T) {
	store := &fakeArtifactStore{}

	result := runNoCodeRefCheck(t, t.TempDir(), store)

	assert.Equal(t, core.StatusSKIP, result.Status)
	assert.Contains(t, result.Summary, "no linked state")
	assert.Zero(t, store.getCalls, "no fetch without linked state")
}

func TestNoCodeRef_ConfigMissing_SKIP(t *testing.T) {
	dir := t.TempDir()

	initStateDir(t, dir)

	store := &fakeArtifactStore{}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusSKIP, result.Status)
	assert.Zero(t, store.getCalls)
}

func TestNoCodeRef_404_SKIP(t *testing.T) {
	// A 404 is artifact-exists' finding; no-coderef SKIPs.
	dir := healthyProject(t)

	store := &fakeArtifactStore{err: &drapi.HTTPError{StatusCode: http.StatusNotFound, URL: "https://test/"}}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusSKIP, result.Status)
	assert.Contains(t, result.Summary, "not found")
}

func TestNoCodeRef_Non404Error_SKIP(t *testing.T) {
	// Any non-404 error → SKIP with connectivity remedy, never FAIL.
	dir := healthyProject(t)

	store := &fakeArtifactStore{err: &drapi.HTTPError{StatusCode: http.StatusInternalServerError, URL: "https://test/"}}

	result := runNoCodeRefCheck(t, dir, store)

	assert.Equal(t, core.StatusSKIP, result.Status)
	assert.NotContains(t, result.Remedy, "--relink")
}

// --- Single-fetch: no-coderef shares the snapshot with remote checks ---

func TestNoCodeRef_SharesSnapshotWithRemoteChecks(t *testing.T) {
	// When run through the full Checks suite, the no-coderef check must
	// share the same snapshot as the four remote checks (one fetch total).
	dir := healthyProject(t)

	store := &fakeArtifactStore{artifact: testArtifact("draft", nil)}

	results := core.NewRunner(Checks(dir, store)...).Run(context.Background())

	// Exactly one fetch for all five remote checks.
	assert.Equal(t, 1, store.getCalls, "no-coderef shares the single-fetch snapshot")

	// The no-coderef result is in the extras block.
	for _, res := range results {
		if res.CheckID == CheckIDNoCodeRef {
			assert.Equal(t, core.StatusWARN, res.Status)

			return
		}
	}

	t.Fatal("no-coderef check not found in results")
}

// --- Drift and catalog-mismatch stay OK for the same shape (VAL-EXTRA-009) ---

func TestNoCodeRef_DriftAndMismatchStayOK_ForNilCodeRef(t *testing.T) {
	// A never-synced draft (nil codeRef, nil local pins): no-coderef WARNs
	// while drift and catalog-mismatch both report OK (both-nil branch).
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	store := &fakeArtifactStore{artifact: testArtifact("draft", nil)}

	results := core.NewRunner(Checks(dir, store)...).Run(context.Background())

	res := byID(results)

	assert.Equal(t, core.StatusWARN, res[CheckIDNoCodeRef].Status, "no-coderef WARNs")
	assert.Equal(t, core.StatusOK, res[CheckIDDrift].Status, "drift stays OK (both-nil)")
	assert.Equal(t, core.StatusOK, res[CheckIDCatalogMismatch].Status, "catalog-mismatch stays OK (both-nil)")
}

func TestNoCodeRef_DriftAndMismatchStayOK_ForBothEmptyCodeRef(t *testing.T) {
	// Same as above but with both-empty-fields codeRef shape.
	dir := t.TempDir()

	initStateDir(t, dir)

	require.NoError(t, wapi.SaveConfig(dir, validConfig("", "")))

	store := &fakeArtifactStore{artifact: testArtifact("draft", &workload.DatarobotCodeRef{
		CatalogID:        "",
		CatalogVersionID: "",
	})}

	results := core.NewRunner(Checks(dir, store)...).Run(context.Background())

	res := byID(results)

	assert.Equal(t, core.StatusWARN, res[CheckIDNoCodeRef].Status, "no-coderef WARNs (both-empty shape)")
	assert.Equal(t, core.StatusOK, res[CheckIDDrift].Status, "drift stays OK (both-empty normalized to nil)")
	assert.Equal(t, core.StatusOK, res[CheckIDCatalogMismatch].Status, "catalog-mismatch stays OK (both-empty)")
}
