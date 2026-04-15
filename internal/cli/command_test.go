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

package cli

import (
	"testing"

	"github.com/datarobot/cli/internal/features"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommandAdder(t *testing.T) {
	tests := []struct {
		name               string
		setupCommands      func() []*cobra.Command
		enabledFeatures    []string
		expectedCmdNames   []string
		expectedNotPresent []string
	}{
		{
			name: "adds ungated commands",
			setupCommands: func() []*cobra.Command {
				return []*cobra.Command{
					{Use: "ungated"},
				}
			},
			expectedCmdNames: []string{"ungated"},
		},
		{
			name: "filters out disabled gated commands",
			setupCommands: func() []*cobra.Command {
				gated := &cobra.Command{Use: "gated"}
				features.SetGate(gated, "my-feature")

				return []*cobra.Command{
					{Use: "ungated"},
					gated,
				}
			},
			enabledFeatures:    []string{},
			expectedCmdNames:   []string{"ungated"},
			expectedNotPresent: []string{"gated"},
		},
		{
			name: "adds enabled gated commands",
			setupCommands: func() []*cobra.Command {
				gated := &cobra.Command{Use: "gated"}
				features.SetGate(gated, "my-feature")

				return []*cobra.Command{
					{Use: "ungated"},
					gated,
				}
			},
			enabledFeatures:  []string{"MY_FEATURE"},
			expectedCmdNames: []string{"ungated", "gated"},
		},
		{
			name: "handles nested gated commands with CommandAdder wrapper",
			setupCommands: func() []*cobra.Command {
				parent := &cobra.Command{Use: "parent"}

				gatedChild := &cobra.Command{Use: "gated-child"}
				features.SetGate(gatedChild, "child-feature")

				parentAdder := &CommandAdder{Command: parent}
				parentAdder.AddCommand(
					&cobra.Command{Use: "ungated-child"},
					gatedChild,
				)

				return []*cobra.Command{parent}
			},
			enabledFeatures:    []string{},
			expectedCmdNames:   []string{"parent"},
			expectedNotPresent: []string{"gated-child"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars for enabled features
			for _, feature := range tt.enabledFeatures {
				t.Setenv("DATAROBOT_CLI_FEATURE_"+feature, "true")
			}

			root := &CommandAdder{
				Command: &cobra.Command{Use: "root"},
			}
			root.AddCommand(tt.setupCommands()...)

			// Check expected commands are present
			for _, expectedName := range tt.expectedCmdNames {
				found := false

				for _, cmd := range root.Commands() {
					if cmd.Name() == expectedName {
						found = true
						break
					}
				}

				assert.True(t, found, "expected command %s to be present", expectedName)
			}

			// Check expected not present are absent
			for _, notExpectedName := range tt.expectedNotPresent {
				found := false

				for _, cmd := range root.Commands() {
					if cmd.Name() == notExpectedName {
						found = true
						break
					}
				}

				assert.False(t, found, "expected command %s to be absent", notExpectedName)
			}
		})
	}
}
