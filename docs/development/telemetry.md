# Telemetry Event Wiring

This document explains how telemetry events are wired to CLI commands and how to add telemetry for new commands.

## Overview

Telemetry events are fired declaratively in `cmd/telemetry_events.go`. The wiring mechanism is centralized in `cmd/root.go`'s `PersistentPreRunE` hook, which fires the appropriate event before the command's business logic runs.

This approach ensures:
- **Declarative**: All wirings are visible in one place (`telemetryEventMap`)
- **Safe**: Events fire in `PersistentPreRunE` before commands that call `os.Exit` directly
- **Minimal code**: No changes needed to individual command files
- **Extensible**: Adding a new event requires only one map entry

## Architecture

```
User invokes command
    ↓
Cobra parses flags
    ↓
PersistentPreRunE (root.go)
    ├─ Initialize telemetry client
    ├─ Lookup command in telemetryEventMap
    └─ Fire event via client.Track()
    ↓
RunE / Run executes (may call os.Exit)
    ↓
PersistentPostRunE (root.go)
    └─ Flush telemetry (3-second timeout)
```

The telemetry client is stored in the command's context for access by subcommands and packages. Common properties (session ID, CLI version, user ID, environment, template name, etc.) are automatically merged into every event by `client.Track()`.

## How to Add Telemetry to a New Command

### 1. Identify the Event and Command Path

First, determine:
- The event type (e.g., `"dr dotenv setup"`)
- The parameters the event should include (e.g., `template_name`)
- How to extract those parameters from the command or arguments

### 2. Create an Event Constructor (if not already present)

Add a typed constructor to `internal/telemetry/events.go`:

```go
func NewDrFooBarEvent(param1, param2 string) types.Event {
    return types.Event{
        EventType: "dr foo bar",
        EventProperties: map[string]any{
            "param_1": param1,
            "param_2": param2,
        },
    }
}
```

Also add a test in `internal/telemetry/events_test.go`.

### 3. Wire the Command in `cmd/telemetry_events.go`

Add an entry to `telemetryEventMap`:

```go
"dr foo bar": func(cmd *cobra.Command, args []string) types.Event {
    // Extract parameters from cmd flags or args
    return telemetry.NewDrFooBarEvent(firstArg(args), "")
},
```

**Common patterns:**

- **Simple command with no parameters:**
  ```go
  "dr foo": func(*cobra.Command, []string) types.Event { 
      return telemetry.NewDrFooEvent("") 
  },
  ```

- **Command with argument:**
  ```go
  "dr foo bar": func(_ *cobra.Command, args []string) types.Event { 
      return telemetry.NewDrFooBarEvent(firstArg(args), "") 
  },
  ```

- **Command with flag:**
  ```go
  "dr foo bar": func(cmd *cobra.Command, _ []string) types.Event { 
      val, _ := cmd.Flags().GetString("flag-name")
      return telemetry.NewDrFooBarEvent(val, "") 
  },
  ```

### 4. Test the Wiring

Add a test in `cmd/telemetry_events_test.go` (create if it doesn't exist):

```go
func TestTelemetryEventMap_DrFooBar(t *testing.T) {
    cmd := &cobra.Command{
        Use: "bar",
        // ... other fields ...
    }
    cmd.Flags().String("my-flag", "default", "Help text")
    _ = cmd.Flags().Set("my-flag", "test-value")

    factory, ok := telemetryEventMap["dr foo bar"]
    assert.True(t, ok)

    event := factory(cmd, []string{"arg1"})
    assert.Equal(t, "dr foo bar", event.EventType)
    assert.Equal(t, "arg1", event.EventProperties["param_1"])
}
```

## Special Case: Dynamic Commands (Plugins)

For commands that are discovered/registered at runtime (e.g., plugin commands), use **cobra annotations** instead of the static map:

```go
// In createPluginCommand or wherever the command is created:
cmd.Annotations = map[string]string{
    "telemetry:plugin-name":    pluginName,
    "telemetry:plugin-version": pluginVersion,
}
```

The `fireCommandEvent` function in `cmd/telemetry_events.go` checks for these annotations and fires `NewDrPluginExecuteEvent` automatically.

## Template Name Handling

You may notice that event constructors are called with `template_name=""`. This is intentional:

- The telemetry client automatically detects the template name by scanning `.datarobot/answers/` in the repository
- This value is merged into `CommonProperties` and included in every event
- There's no need to extract it per-command; it's detected once per CLI invocation and reused

If you need a different template name for a specific event, pass it explicitly to the event constructor.

## Testing

Run the telemetry test suite:

```bash
task test -- internal/telemetry/...
```

To test wiring in an integration context:

```bash
# Dry-run mode shows events without sending them
DATAROBOT_CLI_DISABLE_TELEMETRY=false dr <command> --help
```

Events will be logged to the debug logger (`.dr-tui-debug.log` if `--debug` is set).

## Cleanup

- **Changing a command name?** Update the key in `telemetryEventMap`.
- **Removing a command?** Remove the entry from `telemetryEventMap`.
- **Changing event properties?** Update the event constructor and its tests.

All changes are localized to `cmd/telemetry_events.go` and `internal/telemetry/events.go`.
