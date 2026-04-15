# Feature Flags Design Decision

**Date:** 2025-04-15  
**Status:** Accepted  
**Author:** Factory Droid  
**Related PR:** Feature flag system for `dr workload` command

## Problem Statement

The DataRobot CLI is expanding with new features (e.g., `workload` command) that need to remain hidden from end users until they're ready for release. The options were:

1. **Not registering commands at all** — requires manual code changes at release time (error-prone)
2. **Checking feature gates in command `Run()` logic** — scattered guards, inconsistent UX (command appears in help but errors when run)
3. **Filtering commands at registration time** — declarative and reusable

We need a mechanism that lets developers gate any command with minimal boilerplate, without touching root initialization logic for each new feature.

## Decision

Implement an **annotation-based feature gate system** using Cobra's built-in `Annotations` map and a `CommandAdder` wrapper that filters commands at registration time.

### Rationale

| Approach | Pros | Cons |
|----------|------|------|
| **Annotations + CommandAdder (Chosen)** | Zero boilerplate; declarative; single enforcement point; no per-command guards | Env-var-only (for now); filtering happens at init, not runtime |
| Manual registration | Simple; explicit | Error-prone; requires code changes per feature; doesn't scale |
| Per-command logic | Flexible; runtime control | Scattered logic; hard to maintain; inconsistent UX |
| External feature gate service | Powerful; centralized | Overkill for CLI; adds infrastructure dependency; complex setup |

**Why annotations?**
- Already part of Cobra; zero new dependencies
- Declarative — developers declare what they're building, not *how* to gate it
- Searchable — easy to find all gated commands (`grep -r "feature-gate"`)

**Why `CommandAdder` at registration time?**
- Commands with disabled gates never enter the tree — no post-hoc cleanup needed
- Single enforcement point (`AddCommand`) rather than a separate traversal pass
- Implicit child removal — if the parent is gated and not added, its children never exist
- The same pattern composes for nested gated subcommands: wrap the parent with `CommandAdder` too

**Why env-var-only initially?**
- Feature gating runs in `init()`, before Viper configuration is loaded
- Env vars require no initialization order changes
- Config file support can be added later without changing the public API

## Design

### Architecture

```
┌─────────────────────────────────┐
│  cmd/root.go init()             │
├─────────────────────────────────┤
│ RootCmd is a *cli.CommandAdder  │
│                                 │
│ RootCmd.AddCommand(...)         │
│   - auth, component, dotenv...  │
│   - workload (gated annotation) │← filtered here if feature is off
│   - plugin, etc.                │
└─────────────────────────────────┘
         ↓
┌─────────────────────────────────┐
│  User runs: dr --help           │
├─────────────────────────────────┤
│ Only enabled commands visible   │
└─────────────────────────────────┘
```

### Key Components

#### `internal/features/`

```go
const AnnotationKey = "feature-gate"

// Enabled checks DATAROBOT_CLI_FEATURE_<NAME>=true|1
func Enabled(name string) bool {
    envKey := "DATAROBOT_CLI_FEATURE_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
    v := os.Getenv(envKey)
    return strings.EqualFold(v, "true") || v == "1"
}

// SetGate attaches a feature gate annotation to a command.
func SetGate(cmd *cobra.Command, name string) { ... }
```

#### `internal/cli/`

```go
// CommandAdder wraps cobra.Command and filters gated children at AddCommand time.
type CommandAdder struct{ *cobra.Command }

func (gc *CommandAdder) AddCommand(cmds ...*cobra.Command) {
    for _, cmd := range cmds {
        if gate, ok := cmd.Annotations[features.AnnotationKey]; ok && !features.Enabled(gate) {
            continue // never added to the tree
        }
        gc.Command.AddCommand(cmd)
    }
}
```

**Why this design:**
- `Enabled()` is a pure function (testable, no side effects)
- `CommandAdder` filters at registration time — disabled commands never enter the tree
- No per-command guards or `if` statements scattered through the codebase

#### Command annotation

```go
func Cmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "workload",
        Short: "Workload management commands",
    }
    features.SetGate(cmd, "workload")
    return cmd
}
```

#### Activation

```bash
# Feature name → env var
# "workload"    → DATAROBOT_CLI_FEATURE_WORKLOAD
# "my-feature"  → DATAROBOT_CLI_FEATURE_MY_FEATURE

DATAROBOT_CLI_FEATURE_WORKLOAD=true dr workload --help
```

## Alternatives Considered

### 1. External Feature Flag Service (e.g., GO Feature Flag, LaunchDarkly)

**Why rejected:**
- Overkill for a CLI; deployment = shipping the binary
- Adds infrastructure dependency and potential network failure modes
- No need for gradual rollouts, targeting, or A/B testing in a single-user CLI

### 2. Command-Level Guard Logic

```go
if !features.Enabled("workload") {
    return fmt.Errorf("workload command not yet available")
}
```

**Why rejected:**
- Scattered logic across many command files
- Command still appears in help text even when disabled
- Tab completion includes disabled commands

### 3. Conditional `AddCommand` Helper

```go
func addIfEnabled(parent *cobra.Command, name string, cmd *cobra.Command) {
    if features.Enabled(name) {
        parent.AddCommand(cmd)
    }
}
```

**Why rejected:**
- Still requires a call-site change in `root.go` per gated command
- The feature name is specified twice (helper call + command definition)
- `CommandAdder` achieves the same with zero extra call-site code

### 4. Config File Only (No Env Vars)

**Why rejected:**
- Viper initializes in `PersistentPreRunE`, after `init()` where commands are registered
- Would require restructuring initialization order

**Accepted as future enhancement:** env var support now, config file support later via a `Provider` interface already in place.

## Testing

- ✅ `TestEnabled()` — env var true/1/false/unset (`internal/features/features_test.go`)
- ✅ `TestCommandAdder()` — top-level and nested filtering (`internal/cli/command_test.go`)
- ✅ `TestWorkloadCommandNotPresentByDefault()` — integration check (`cmd/root_test.go`)

## Security

Feature gates are **not** an access control mechanism. Any user with shell access can enable a disabled feature via env var. They exist solely to prevent casual discovery of unfinished features.

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|-----------|
| Developer forgets annotation | High | Code review; pattern documented in AGENTS.md |
| Config file support blocked by init order | Medium | Documented limitation; `Provider` interface enables future extension |
| Env var naming inconsistency | Low | Convention documented; `SetGate` enforces the annotation key |

## References

- [Cobra Command Documentation](https://pkg.go.dev/github.com/spf13/cobra)
- [Developer Guide: Feature Flags](./feature-flags.md)
- [DataRobot CLI Architecture](./structure.md)
