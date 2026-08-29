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
	"github.com/datarobot/cli/internal/workload"
)

// noCodeRefCheck reports whether the linked artifact has a usable code
// reference. An artifact that exists but was never synced from this project
// has no codeRef (ExtractCodeRef returns nil, or both CatalogID and
// CatalogVersionID are empty). This is distinct from drift: drift compares a
// populated local pin against a remote version, while no-coderef means the
// artifact simply has no codeRef at all. After a real sync establishes the
// codeRef, this check reports OK.
//
// The check shares the same artifact snapshot as the ticket-scope remote
// checks (one fetch per run). A 404 SKIPs (artifact-exists owns that
// finding); any other fetch failure SKIPs with the connectivity remedy.
type noCodeRefCheck struct {
	remoteBase
}

func (c *noCodeRefCheck) ID() string {
	return CheckIDNoCodeRef
}

func (c *noCodeRefCheck) Name() string {
	return "Artifact code reference"
}

// Run fetches the shared snapshot and checks whether the artifact carries a
// usable codeRef. A nil codeRef or one whose CatalogID and CatalogVersionID
// are both empty (empty ≈ nil normalization) means the artifact was never
// synced → WARN. The WARN is informational and not fixable: the user must
// run a sync to establish the codeRef. A 404 or other fetch failure SKIPs
// (the remote checks own those findings).
func (c *noCodeRefCheck) Run(_ context.Context) core.Result {
	cfg, res, ok := c.linkedConfig()
	if !ok {
		return res
	}

	art, res, _, ok := c.fetchedArtifact(cfg)
	if !ok {
		return res
	}

	if hasUsableCodeRef(art) {
		return core.Result{
			Status:  core.StatusOK,
			Summary: "artifact has a usable code reference",
		}
	}

	return core.Result{
		Status:  core.StatusWARN,
		Summary: "artifact has no usable code reference (never synced from this project)",
		Remedy:  RemedyNoCodeRef,
		Fixable: false,
	}
}

// hasUsableCodeRef reports whether the artifact carries a codeRef with at
// least one populated field. "No usable codeRef" = nil OR both
// CatalogID/CatalogVersionID empty (empty ≈ nil normalization), matching the
// pinned semantics shared with the drift and catalog-mismatch checks.
func hasUsableCodeRef(art *workload.Artifact) bool {
	codeRef := workload.ExtractCodeRef(*art)
	if codeRef == nil {
		return false
	}

	// Both fields empty = no usable codeRef (empty ≈ nil normalization).
	return codeRef.CatalogID != "" || codeRef.CatalogVersionID != ""
}
