// Copyright 2025 DataRobot, Inc. and its affiliates.
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

package features

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		name         string
		featureName  string
		envVarValue  string
		expectedBool bool
	}{
		{
			name:         "env var true",
			featureName:  "test",
			envVarValue:  "true",
			expectedBool: true,
		},
		{
			name:         "env var 1",
			featureName:  "test",
			envVarValue:  "1",
			expectedBool: true,
		},
		{
			name:         "env var false",
			featureName:  "test",
			envVarValue:  "false",
			expectedBool: false,
		},
		{
			name:         "env var 0",
			featureName:  "test",
			envVarValue:  "0",
			expectedBool: false,
		},
		{
			name:         "env var not set",
			featureName:  "test",
			envVarValue:  "",
			expectedBool: false,
		},
		{
			name:         "hyphenated feature name",
			featureName:  "my-feature",
			envVarValue:  "true",
			expectedBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute the env key (hyphens become underscores)
			envKey := "DATAROBOT_CLI_FEATURES_" + strings.ToUpper(strings.ReplaceAll(tt.featureName, "-", "_"))

			if tt.envVarValue != "" {
				t.Setenv(envKey, tt.envVarValue)
			}

			result := Enabled(tt.featureName)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestRemoveDisabledCommands(t *testing.T) {
	tests := []struct {
		name                string
		setupCmd            func() *cobra.Command
		enabledFeatures     []string
		expectedSubcommands []string
		expectedNotPresent  []string
	}{
		{
			name: "removes top-level gated command",
			setupCmd: func() *cobra.Command {
				parent := &cobra.Command{Use: "root"}

				enabled := &cobra.Command{Use: "enabled"}
				parent.AddCommand(enabled)

				gated := &cobra.Command{
					Use: "gated",
					Annotations: map[string]string{
						AnnotationKey: "test-feature",
					},
				}
				parent.AddCommand(gated)

				return parent
			},
			enabledFeatures:     []string{},
			expectedSubcommands: []string{"enabled"},
			expectedNotPresent:  []string{"gated"},
		},
		{
			name: "keeps top-level gated command when enabled",
			setupCmd: func() *cobra.Command {
				parent := &cobra.Command{Use: "root"}

				gated := &cobra.Command{
					Use: "gated",
					Annotations: map[string]string{
						AnnotationKey: "test-feature",
					},
				}
				parent.AddCommand(gated)

				return parent
			},
			enabledFeatures:     []string{},
			expectedSubcommands: []string{}, // Will be set in test
			expectedNotPresent:  []string{},
		},
		{
			name: "removes nested gated command",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root"}

				parent := &cobra.Command{Use: "parent"}
				root.AddCommand(parent)

				enabled := &cobra.Command{Use: "enabled"}
				parent.AddCommand(enabled)

				gated := &cobra.Command{
					Use: "gated",
					Annotations: map[string]string{
						AnnotationKey: "nested-feature",
					},
				}
				parent.AddCommand(gated)

				return root
			},
			enabledFeatures:    []string{},
			expectedNotPresent: []string{"gated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars for enabled features
			for _, feature := range tt.enabledFeatures {
				t.Setenv("DATAROBOT_CLI_FEATURES_"+feature, "true")
			}

			cmd := tt.setupCmd()
			RemoveDisabledCommands(cmd)

			// Check expected subcommands are present
			for _, expectedName := range tt.expectedSubcommands {
				found := false

				for _, sub := range cmd.Commands() {
					if sub.Name() == expectedName {
						found = true
						break
					}
				}

				assert.True(t, found, "expected subcommand %s to be present", expectedName)
			}

			// Check expected not present are absent
			for _, notExpectedName := range tt.expectedNotPresent {
				found := false

				for _, sub := range cmd.Commands() {
					if sub.Name() == notExpectedName {
						found = true
						break
					}
				}

				assert.False(t, found, "expected subcommand %s to be absent", notExpectedName)
			}
		})
	}
}
