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

package list

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/datarobot/cli/internal/workload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()

	os.Stdout = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func makeArtifact(id, name, status, catalogID, catalogVersionID string) workload.Artifact {
	a := workload.Artifact{
		ID:        id,
		Name:      name,
		Status:    status,
		CreatedAt: time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 10, 14, 30, 0, 0, time.UTC),
	}

	if catalogID != "" {
		a.Spec = workload.Spec{
			ContainerGroups: []workload.ContainerGroup{
				{
					Containers: []workload.Container{
						{
							CodeRef: &workload.CodeRef{
								Datarobot: &workload.DatarobotCodeRef{
									CatalogID:        catalogID,
									CatalogVersionID: catalogVersionID,
								},
							},
						},
					},
				},
			},
		}
	}

	return a
}

func TestPrintTable_WithCodeRef(t *testing.T) {
	artifacts := []workload.Artifact{
		makeArtifact("art-abc-123", "my-agent", "DRAFT", "0123456789abcdef01234567", "fedcba0987654321fedcba09"),
	}

	output := captureStdout(t, func() {
		printTable(artifacts)
	})

	assert.Contains(t, output, "ARTIFACT ID")
	assert.Contains(t, output, "CATALOG ID")
	assert.Contains(t, output, "VERSION ID")
	assert.Contains(t, output, "art-abc-123")
	assert.Contains(t, output, "my-agent")
	assert.Contains(t, output, "DRAFT")
	assert.Contains(t, output, "0123456789abcdef01234567")
	assert.Contains(t, output, "fedcba0987654321fedcba09")
	assert.Contains(t, output, "2026-04-10 14:30 UTC")
}

func TestPrintTable_WithoutCodeRef(t *testing.T) {
	artifacts := []workload.Artifact{
		makeArtifact("art-abc-123", "my-agent", "DRAFT", "", ""),
	}

	output := captureStdout(t, func() {
		printTable(artifacts)
	})

	assert.Contains(t, output, "\u2014")
}

func TestPrintTable_Empty(t *testing.T) {
	output := captureStdout(t, func() {
		printTable([]workload.Artifact{})
	})

	assert.Equal(t, "No artifacts found.\n", output)
}

func TestPrintJSON_Array(t *testing.T) {
	artifacts := []workload.Artifact{
		makeArtifact("art-001", "agent-one", "DRAFT", "cat-001", "ver-001"),
		makeArtifact("art-002", "agent-two", "LOCKED", "", ""),
	}

	output := captureStdout(t, func() {
		err := printJSON(artifacts)
		require.NoError(t, err)
	})

	var parsed []map[string]interface{}

	err := json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "art-001", parsed[0]["id"])
	assert.Equal(t, "ver-001", parsed[0]["versionId"])
	assert.Equal(t, "art-002", parsed[1]["id"])
	assert.Empty(t, parsed[1]["versionId"])
}

func TestPrintJSON_EmptyArray(t *testing.T) {
	output := captureStdout(t, func() {
		err := printJSON([]workload.Artifact{})
		require.NoError(t, err)
	})

	var parsed []interface{}

	err := json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)
	assert.Empty(t, parsed)
}

func TestPrintJSON_FullVersion(t *testing.T) {
	artifacts := []workload.Artifact{
		makeArtifact("art-abc-123", "my-agent", "DRAFT", "cat-xyz-789", "fedcba0987654321fedcba09"),
	}

	output := captureStdout(t, func() {
		err := printJSON(artifacts)
		require.NoError(t, err)
	})

	var parsed []map[string]interface{}

	err := json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "fedcba0987654321fedcba09", parsed[0]["versionId"])
}

func TestCmd_InvalidOutputFormat(t *testing.T) {
	cmd := Cmd()
	cmd.PreRunE = nil
	cmd.SetArgs([]string{"--output", "xml"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format: xml")
}

func TestCmd_InvalidStatus(t *testing.T) {
	cmd := Cmd()
	cmd.PreRunE = nil
	cmd.SetArgs([]string{"--status", "UNKNOWN"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestCmd_InvalidLimit(t *testing.T) {
	for _, v := range []string{"-1", "0"} {
		cmd := Cmd()
		cmd.PreRunE = nil
		cmd.SetArgs([]string{"--limit", v})

		err := cmd.Execute()
		require.Error(t, err, "limit %s", v)
		assert.Contains(t, err.Error(), "must be positive")
	}
}
