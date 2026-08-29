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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/workload/wapi"
)

// checkoutsOrphanedCheck reports whether the .checkouts/ directory holds
// orphaned version snapshots from `dr artifact code checkout`. The check is
// informational: it reports the snapshot-dir count and total disk usage, and
// is not fixable (the doctor never deletes .checkouts/ snapshots). An absent
// .checkouts/ directory is OK (the healthy steady state).
type checkoutsOrphanedCheck struct {
	projectDir string
}

func (c *checkoutsOrphanedCheck) ID() string {
	return CheckIDCheckoutsOrphaned
}

func (c *checkoutsOrphanedCheck) Name() string {
	return "Orphaned checkouts"
}

// Run reads the .checkouts/ directory and counts snapshot subdirectories. A
// missing directory is OK. Zero snapshots is OK. One or more snapshots is a
// WARN with the count and total disk usage. The directory is never modified.
func (c *checkoutsOrphanedCheck) Run(_ context.Context) core.Result {
	if res, skip := skipIfUnlinked(c.projectDir); skip {
		return res
	}

	checkoutsDir := wapi.CheckoutsDir(c.projectDir)

	entries, err := os.ReadDir(checkoutsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return core.Result{
				Status:  core.StatusOK,
				Summary: "no .checkouts/ directory",
			}
		}

		// An unreadable directory is not an orphaned-snapshots condition;
		// report OK so a permission issue does not maslead the user into
		// thinking snapshots are orphaned.
		return core.Result{
			Status:  core.StatusOK,
			Summary: "cannot read .checkouts/ directory (permission issue?)",
		}
	}

	// Count snapshot subdirectories (version snapshots are dirs).
	var count int

	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}

	if count == 0 {
		return core.Result{
			Status:  core.StatusOK,
			Summary: "no .checkouts/ snapshots",
		}
	}

	usage := dirDiskUsage(checkoutsDir)

	return core.Result{
		Status:  core.StatusWARN,
		Summary: fmt.Sprintf(".checkouts/ has %d snapshot dir(s) using %s of disk", count, humanSize(usage)),
		Remedy:  RemedyCheckoutsOrphaned,
		Fixable: false,
	}
}

// dirDiskUsage walks dir and sums the sizes of all regular files. It is
// best-effort: unreadable files are silently skipped (the check is
// informational, not a byte-exact audit).
func dirDiskUsage(dir string) int64 {
	var total int64

	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		total += info.Size()

		return nil
	})

	return total
}

// humanSize formats a byte count as a human-readable string (B, KiB, MiB,
// GiB). It mirrors the format used by the sync display plan so disk-usage
// figures are consistent across the CLI.
func humanSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KiB", float64(n)/1024.0)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MiB", float64(n)/1024.0/1024.0)
	default:
		return fmt.Sprintf("%.1f GiB", float64(n)/1024.0/1024.0/1024.0)
	}
}
