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
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/datarobot/cli/internal/auth"
	"github.com/datarobot/cli/internal/workload"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var outputFormat string

	var limit int

	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workload artifacts.",
		Long: `List workload artifacts in your DataRobot deployment infrastructure.

This command fetches artifacts and shows:
  * Name and current status (draft or locked)
  * Code reference catalog ID and version ID
  * Last update timestamp

By default, output is a human-readable table. Use --output json for machine-parseable output.

Example:
  dr workload artifact list
  dr workload artifact list --limit 10
  dr workload artifact list --status draft
  dr workload artifact list --output json`,
		Args:    cobra.NoArgs,
		PreRunE: auth.EnsureAuthenticatedE,
		RunE: func(_ *cobra.Command, _ []string) error {
			if outputFormat != "" && outputFormat != "json" {
				return fmt.Errorf("invalid output format: %s (supported: json)", outputFormat)
			}

			if limit <= 0 {
				return fmt.Errorf("invalid --limit %d: must be positive", limit)
			}

			status, err := workload.ParseArtifactStatus(status)
			if err != nil {
				return err
			}

			artifacts, err := workload.ListArtifacts(limit, status)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return printJSON(artifacts)
			}

			printTable(artifacts)

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "", "Output format (json)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of artifacts to return")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (draft, locked)")

	return cmd
}

func printJSON(artifacts []workload.Artifact) error {
	outputs := make([]workload.ArtifactOutput, 0, len(artifacts))

	for _, a := range artifacts {
		outputs = append(outputs, workload.NewArtifactOutput(a))
	}

	data, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func printTable(artifacts []workload.Artifact) {
	if len(artifacts) == 0 {
		fmt.Println("No artifacts found.")

		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "ARTIFACT ID\tNAME\tSTATUS\tCATALOG ID\tVERSION ID\tUPDATED\n")

	for _, a := range artifacts {
		var catalogID, versionID string

		if codeRef := workload.ExtractCodeRef(a); codeRef != nil {
			catalogID = codeRef.CatalogID
			versionID = codeRef.CatalogVersionID
		}

		updated := a.UpdatedAt.UTC().Format("2006-01-02 15:04 UTC")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", a.ID, a.Name, a.Status, orDash(catalogID), orDash(versionID), updated)
	}

	w.Flush()
}

func orDash(v string) string {
	if v == "" {
		return "\u2014"
	}

	return v
}
