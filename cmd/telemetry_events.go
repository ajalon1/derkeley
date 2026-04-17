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

package cmd

import (
	"github.com/amplitude/analytics-go/amplitude/types"
	"github.com/datarobot/cli/internal/telemetry"
	"github.com/spf13/cobra"
)

// telemetryEventMap is a declarative mapping from cobra command paths to telemetry event factories.
// Each entry defines how to construct a telemetry event for that command.
// To wire a new command to telemetry, add an entry here.
var telemetryEventMap = map[string]func(*cobra.Command, []string) types.Event{
	"dr start": func(*cobra.Command, []string) types.Event { return telemetry.NewDrStartEvent("") },
	"dr run":   func(_ *cobra.Command, args []string) types.Event { return telemetry.NewDrRunEvent("", firstArg(args)) },
	"dr task":  func(_ *cobra.Command, args []string) types.Event { return telemetry.NewDrTaskEvent("", firstArg(args)) },
	"dr auth set-url": func(_ *cobra.Command, args []string) types.Event {
		return telemetry.NewDrAuthSetURLEvent(firstArg(args))
	},
	"dr dotenv setup":    func(*cobra.Command, []string) types.Event { return telemetry.NewDrDotenvSetupEvent("") },
	"dr dotenv update":   func(*cobra.Command, []string) types.Event { return telemetry.NewDrDotenvUpdateEvent("") },
	"dr dotenv validate": func(*cobra.Command, []string) types.Event { return telemetry.NewDrDotenvValidateEvent("") },
	"dr component add": func(_ *cobra.Command, args []string) types.Event {
		return telemetry.NewDrComponentAddEvent(firstArg(args), "")
	},
	"dr component update": func(_ *cobra.Command, args []string) types.Event {
		return telemetry.NewDrComponentUpdateEvent(firstArg(args), "")
	},
	"dr template setup": func(*cobra.Command, []string) types.Event { return telemetry.NewDrTemplateSetupEvent("") },
	"dr plugin install": func(cmd *cobra.Command, args []string) types.Event {
		ver, _ := cmd.Flags().GetString("version")
		return telemetry.NewDrPluginInstallEvent(firstArg(args), ver)
	},
	"dr plugin uninstall": func(_ *cobra.Command, args []string) types.Event {
		return telemetry.NewDrPluginUninstallEvent(firstArg(args), "")
	},
	"dr plugin update": func(_ *cobra.Command, args []string) types.Event {
		return telemetry.NewDrPluginUpdateEvent(firstArg(args), "")
	},
}

// fireCommandEvent fires the telemetry event for the given command if it appears in telemetryEventMap
// or carries plugin-execute annotations (set by createPluginCommand in plugin/discovery.go).
// Common properties (e.g., template_name) are merged automatically by client.Track().
func fireCommandEvent(cmd *cobra.Command, args []string, client *telemetry.Client) {
	if factory, ok := telemetryEventMap[cmd.CommandPath()]; ok {
		client.Track(factory(cmd, args))
		return
	}

	// Dynamic plugin commands declare their metadata via annotations.
	if name, ok := cmd.Annotations["telemetry:plugin_name"]; ok {
		version := cmd.Annotations["telemetry:plugin_version"]
		client.Track(telemetry.NewDrPluginExecuteEvent(name, version))
	}
}

// firstArg returns the first argument from args, or an empty string if none exist.
func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}

	return ""
}
