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

package get

import (
	"fmt"
	"os"

	"github.com/datarobot/cli/internal/auth"
	"github.com/datarobot/cli/internal/workload"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get <artifact-id>",
		Short: "Display details of a workload artifact.",
		Long: `Display details of a single workload artifact.

This command fetches an artifact by ID and shows:
  • Name and current status
  • Code reference (catalog ID and version)
  • Creation and last update timestamps

By default, output is human-readable. Use --output json for machine-parseable output.

Example:
  dr workload artifact get art-abc-123
  dr workload artifact get art-abc-123 --output json`,
		Args:    cobra.ExactArgs(1),
		PreRunE: auth.EnsureAuthenticatedE,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputFormat != "" && outputFormat != "json" {
				return fmt.Errorf("invalid output format: %s (supported: json)", outputFormat)
			}

			artifact, err := workload.GetArtifact(args[0])
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				return workload.ArtifactJSONRenderer{}.Render(os.Stdout, *artifact)
			}

			return workload.ArtifactRenderer{}.Render(os.Stdout, *artifact)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "", "Output format (json)")

	return cmd
}
