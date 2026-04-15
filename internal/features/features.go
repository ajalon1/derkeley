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

package features

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const AnnotationKey = "feature-gate"

// CommandAdder wraps cobra.Command and overrides AddCommand to filter gated commands.
// It allows commands with disabled feature gates to never be added to the command tree.
// This wrapper clarifies that the root command itself is not gated—it just intelligently
// adds its children, filtering out those with disabled feature gates.
type CommandAdder struct {
	*cobra.Command
}

// AddCommand adds commands to this command, skipping any that have a disabled feature gate.
// Gated commands (those with a feature-gate annotation) are filtered at registration time,
// never making it into the command tree if their feature is not enabled.
func (ca *CommandAdder) AddCommand(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		if gate, ok := cmd.Annotations[AnnotationKey]; ok && !Enabled(gate) {
			continue
		}

		ca.Command.AddCommand(cmd)
	}
}

// SetGate adds a feature gate annotation to a command, preserving any existing annotations.
func SetGate(cmd *cobra.Command, name string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations[AnnotationKey] = name
}

// Enabled checks env var DATAROBOT_CLI_FEATURES_<NAME>.
// Currently only env vars are supported; config file support requires
// Viper initialization which happens after command registration.
// TODO: Support config file (drconfig.yaml) feature flags once
// we move filtering to PersistentPreRunE or read config independently.
func Enabled(name string) bool {
	envKey := "DATAROBOT_CLI_FEATURES_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	if v := os.Getenv(envKey); v != "" {
		return strings.EqualFold(v, "true") || v == "1"
	}

	return false
}

// RemoveDisabledCommands recursively removes any subcommands
// that have a "feature-gate" annotation whose feature is not enabled.
func RemoveDisabledCommands(parent *cobra.Command) {
	for _, cmd := range parent.Commands() {
		if gate, ok := cmd.Annotations[AnnotationKey]; ok && !Enabled(gate) {
			parent.RemoveCommand(cmd)
		} else {
			RemoveDisabledCommands(cmd)
		}
	}
}
