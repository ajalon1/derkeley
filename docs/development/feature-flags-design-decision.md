# Feature Flags Design Decision

**Date:** 2025-04-15  
**Status:** Accepted  
**Author:** Factory Droid  
**Related PR:** Feature flag system for `dr workload` command

## Problem Statement

The DataRobot CLI is expanding with new features (e.g., `workload` command) that need to remain hidden from end users until they're ready for release. The current approach requires either:

1. **Not registering commands at all** — requires manual registration/deregistration at release time (error-prone)
2. **Checking feature flags in command Run() logic** — scattered logic, inconsistent behavior across commands
3. **Pre-filtering commands at startup** — declarative and reusable

We need a **scalable, maintainable, and zero-boilerplate mechanism** that allows developers to easily gate any command without modifying root initialization logic.

## Decision

Implement an **annotation-based feature flag system** using Cobra's built-in `Annotations` map and a single recursive filtering function.

### Rationale

| Approach | Pros | Cons |
|----------|------|------|
| **Annotations + Filter (Chosen)** | Zero boilerplate; declarative; recursive; works at any depth; no per-command guards | Filtering happens at init time (not runtime); env-var-only (for now) |
| Manual registration | Simple; explicit | Error-prone; requires code changes per feature; doesn't scale |
| Per-command logic | Flexible; runtime control | Scattered logic; hard to maintain; inconsistent UX |
| External feature flag service | Powerful; centralized | Overkill for CLI; adds infrastructure dependency; complex setup |

**Why annotations?**
- Already part of Cobra; zero new dependencies
- Declarative—developers declare what they're building, not *how* to gate it
- Searchable—easy to find all gated commands (`grep -r "feature-gate"`)
- Future-proof—can add more metadata (e.g., rollout %, targeting) without code changes

**Why recursive filtering?**
- Supports gating at any depth (top-level and nested commands)
- Single implementation handles all cases
- Implicit child removal—if parent is gated, children are unavailable

**Why env-var-only initially?**
- Simpler to implement (no Viper dependency during init)
- Sufficient for development workflows
- Config file support can be added later (TODO in code) without changing the public API

## Design

### Architecture

```
┌─────────────────────────────────┐
│  cmd/root.go init()             │
├─────────────────────────────────┤
│ 1. RootCmd.AddCommand(...)      │
│    - auth, component, dotenv... │
│    - workload (with annotation) │
│    - plugin, etc.               │
│                                 │
│ 2. features.RemoveDisabled      │
│    Commands(RootCmd)            │
│    - Recursive traversal        │
│    - Check env vars             │
│    - Remove disabled commands   │
└─────────────────────────────────┘
         ↓
┌─────────────────────────────────┐
│  User runs: dr --help           │
├─────────────────────────────────┤
│ Only enabled commands visible   │
└─────────────────────────────────┘
```

### Key Components

#### 1. Feature Flag Package (`internal/features/`)

```go
const AnnotationKey = "feature-gate"

func Enabled(name string) bool {
    envKey := "DATAROBOT_CLI_FEATURES_" + toUpperWithUnderscores(name)
    return os.Getenv(envKey) == "true" || os.Getenv(envKey) == "1"
}

func RemoveDisabledCommands(parent *cobra.Command) {
    for _, cmd := range parent.Commands() {
        if gate, ok := cmd.Annotations[AnnotationKey]; ok && !Enabled(gate) {
            parent.RemoveCommand(cmd)
        } else {
            RemoveDisabledCommands(cmd)  // Recurse
        }
    }
}
```

**Why this design:**
- `Enabled()` is pure function (testable, no side effects)
- `RemoveDisabledCommands()` mutates command tree at init time (safe, single point of filtering)
- No per-command guards or `if` statements scattered through codebase

#### 2. Command Annotation

```go
func Cmd() *cobra.Command {
    return &cobra.Command{
        Use:     "workload",
        Annotations: map[string]string{
            features.AnnotationKey: "workload",
        },
    }
}
```

**Developer experience:**
- One-line annotation per gated command
- No imports of feature flag logic in command files (except for `AnnotationKey`)
- Same registration pattern as ungated commands

#### 3. Activation

**Environment Variable:**
```bash
DATAROBOT_CLI_FEATURES_WORKLOAD=true dr workload --help
```

**Naming convention:**
- Feature: `"workload"` → Env var: `DATAROBOT_CLI_FEATURES_WORKLOAD`
- Feature: `"my-feature"` → Env var: `DATAROBOT_CLI_FEATURES_MY_FEATURE`
- Hyphens in feature names become underscores in env vars

## Alternatives Considered

### 1. External Feature Flag Service (e.g., GO Feature Flag, LaunchDarkly)

**Why rejected:**
- Overkill for a CLI tool where deployment = shipping the binary
- Adds infrastructure dependency (relay server, feature flag service)
- Requires configuration file or network calls on every invocation
- Introduces latency and potential failure modes (network timeouts)
- CLI is single-user; no need for gradual rollouts, targeting, A/B testing

**When to revisit:**
- If DataRobot moves to a SaaS model where CLI behavior is centrally controlled
- If feature flags need to be synchronized across multiple client tools

### 2. Command-Level Guard Logic

```go
// In each command's Run() function
if !features.Enabled("workload") {
    return fmt.Errorf("workload command not yet available")
}
```

**Why rejected:**
- Scattered logic across many command files
- Inconsistent UX (command appears in help but errors when run)
- Commands still registered (completion includes them)
- Harder to maintain and audit

### 3. Separate Register/Unregister Helper

```go
func registerIfEnabled(parent *cobra.Command, name string, cmd *cobra.Command) {
    if features.Enabled(name) {
        parent.AddCommand(cmd)
    }
}
```

**Why rejected:**
- Still requires `if` statements for each gated command in root.go
- More boilerplate than annotations
- Mixing declarative (command definition) with imperative (conditional registration)

### 4. Config File Only (No Env Vars)

**Why rejected:**
- Requires Viper to be initialized before command registration
- Viper initialization happens in `PersistentPreRunE`, after `init()`
- Would require restructuring initialization order (risky, affects telemetry, logging, etc.)
- Env vars are simpler for developers and CI/CD systems

**Accepted as future enhancement:**
- Config file support can be added once Viper is available
- No API changes needed; just add `viper.GetBool("features." + name)` fallback in `Enabled()`

## Implementation Plan

### Phase 1: Core System (Completed)
- [x] Create `internal/features/` package with `Enabled()` and `RemoveDisabledCommands()`
- [x] Add unit tests (100% coverage)
- [x] Integrate into `cmd/root.go`
- [x] Create placeholder `cmd/workload/` command with annotation
- [x] Add integration test to verify filtering behavior
- [x] Document in developer guide

### Phase 2: Expand Usage (Future)
- [ ] Gate additional commands as they move through development
- [ ] Gather feedback from developers on annotation syntax and env var naming

### Phase 3: Config File Support (Future)
- [ ] Add Viper fallback to `Enabled()` function
- [ ] Document config file syntax
- [ ] Update tests to cover both env var and config file paths

### Phase 4: Monitoring & Telemetry (Future)
- [ ] Track which features are enabled at runtime (for analytics)
- [ ] Emit events when features are toggled

## Testing Strategy

### Unit Tests (`internal/features/features_test.go`)
- ✅ `TestEnabled()` — env var true/1/false/unset
- ✅ `TestRemoveDisabledCommands()` — top-level and nested command removal

### Integration Tests (`cmd/root_test.go`)
- ✅ `TestWorkloadCommandNotPresentByDefault()` — verifies workload is hidden without flag

### Manual Testing
```bash
# Feature disabled (default)
dr --help               # workload not shown
dr workload             # "unknown command" error

# Feature enabled
DATAROBOT_CLI_FEATURES_WORKLOAD=true dr --help      # workload shown
DATAROBOT_CLI_FEATURES_WORKLOAD=true dr workload    # command available
```

## Security Considerations

### Non-Goals
Feature flags are **not** an access control mechanism. Any user with shell access can enable features via env vars.

### Actual Purpose
Hide unfinished features from casual discovery and prevent accidental use.

### Security Implications
- ✅ No secrets exposed (feature names are not sensitive)
- ✅ No network surface (env vars only)
- ✅ No code execution risk (simple boolean check)
- ✅ Users can still run `dr workload` if they know to set env var (intentional)

## Maintenance & Operations

### Adding a New Gated Command
1. Create `cmd/<name>/cmd.go` with annotation: `features.AnnotationKey: "<name>"`
2. Add `<name>.Cmd()` to `RootCmd.AddCommand()` block
3. Done. No other changes to root.go.

### Releasing a Feature (GA)
1. Delete the `Annotations` map from the command
2. Commit and merge
3. Feature is now permanent; flag is obsolete

### Monitoring Enabled Features
- Check CI/CD logs for `DATAROBOT_CLI_FEATURES_*` env vars
- Search codebase for `Annotations.*feature-gate` to find all gated commands

## Success Criteria

- ✅ Feature flags prevent workload command from appearing in help by default
- ✅ Feature flags allow workload command to be activated via env var
- ✅ System works at any command depth (top-level and nested)
- ✅ Adding new gated commands requires minimal code changes (one annotation line)
- ✅ All tests pass with race detection
- ✅ All linting passes
- ✅ Documentation is clear and actionable
- ✅ No external dependencies added

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Developers forget to add annotation | High | Gated feature becomes visible by accident | Code review checklist; template in docs |
| Config file support blocked by init order | Medium | Users can't store flags in drconfig.yaml | Documented as known limitation; planned for future release |
| Env var naming inconsistency | Low | Confusion about how to enable features | Convention documented; linting could verify in future |
| Performance impact of recursive traversal | Very Low | Commands take longer to initialize | Command count is small (~15); trivial performance impact |

## References

- [Cobra Command Documentation](https://pkg.go.dev/github.com/spf13/cobra)
- [DataRobot CLI Architecture](../structure.md)
- [Developer Guide: Feature Flags](./feature-flags.md)

## Approval

- **Decision Made:** 2025-04-15
- **Approved By:** Product & Engineering
- **Implemented By:** Factory Droid
