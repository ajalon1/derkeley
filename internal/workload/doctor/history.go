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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	core "github.com/datarobot/cli/internal/doctor"
	"github.com/datarobot/cli/internal/workload/wapi"
)

// historyCheck reports whether history.log contains unparseable JSONL lines
// (crash truncation). A missing history.log is OK (append-only, lazily
// created on first write); an unreadable one WARNs with RemedyHistoryUnreadable
// (nothing can be parsed from a file that cannot be opened, so the remedy is
// about restoring access, not fixing lines). The check is informational and
// not fixable: the doctor never writes history.log in read-only mode.
type historyCheck struct {
	projectDir string
}

func (c *historyCheck) ID() string {
	return CheckIDHistory
}

func (c *historyCheck) Name() string {
	return "History log"
}

// Run reads history.log line by line and attempts to parse each non-empty
// line as a JSON object. Any unparseable line yields a WARN with fixable=false.
// A file that cannot be read at all (permission or I/O error) WARNs with
// RemedyHistoryUnreadable instead — the mismatched "fix unparseable lines"
// remedy would tell a user who cannot even open the file to edit it. A missing
// file is OK (not an error). The file is never modified.
func (c *historyCheck) Run(_ context.Context) core.Result {
	if res, skip := skipIfUnlinked(c.projectDir); skip {
		return res
	}

	path := filepath.Join(wapi.Dir(c.projectDir), wapi.HistoryFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return core.Result{
				Status:  core.StatusOK,
				Summary: "no history log (lazily created on first write)",
			}
		}

		return core.Result{
			Status:  core.StatusWARN,
			Summary: fmt.Sprintf("cannot read history log (%s)", absPath(path)),
			Remedy:  RemedyHistoryUnreadable,
			Fixable: false,
		}
	}

	// Split into lines and parse each non-empty one as JSON. A trailing
	// newline produces an empty final element which is skipped.
	lines := strings.Split(string(data), "\n")

	var badCount int

	var firstBad int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal([]byte(trimmed), &entry); err != nil {
			badCount++

			if firstBad == 0 {
				firstBad = i + 1
			}
		}
	}

	if badCount > 0 {
		return core.Result{
			Status: core.StatusWARN,
			Summary: fmt.Sprintf(
				"history.log has %d unparseable line(s) (first at line %d)",
				badCount, firstBad,
			),
			Remedy:  RemedyHistory,
			Fixable: false,
		}
	}

	return core.Result{
		Status:  core.StatusOK,
		Summary: "history log is valid",
	}
}
