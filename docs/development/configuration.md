# Configuration & viper integration

This guide covers how the CLI loads, reads, and writes its persistent
configuration (`drconfig.yaml`), and the rules contributors must follow when
adding new flags or persisted config keys.

## Table of contents

- [Where config lives](#where-config-lives)
- [How values reach viper](#how-values-reach-viper)
- [Writing back to drconfig.yaml](#writing-back-to-drconfigyaml)
- [Rules for new flags](#rules-for-new-flags)
- [Rules for new env vars](#rules-for-new-env-vars)
- [Rules for new persisted keys](#rules-for-new-persisted-keys)
- [Common pitfalls](#common-pitfalls)

## Where config lives

The CLI stores user configuration in a single YAML file:

- Default location: `$XDG_CONFIG_HOME/datarobot/drconfig.yaml`
  (falls back to `~/.config/datarobot/drconfig.yaml`)
- Override with `--config <path>` or `DATAROBOT_CLI_CONFIG=<path>`

Today this file holds connection credentials and a small set of sticky CLI
preferences. It is **not** a dumping ground for transient flag state.

## How values reach viper

Viper resolves a key from these sources in priority order:

1. Explicit `viper.Set(key, value)` call (e.g. after a successful login)
2. Flag bound via `viper.BindPFlag(key, flag)` (only persistent root flags
   today &mdash; see below)
3. Environment variable bound via `viper.BindEnv(key, "DATAROBOT_…")`
   or auto-mapped via `viper.SetEnvPrefix("DATAROBOT_CLI")`
4. Value loaded from `drconfig.yaml`
5. Default registered via `viper.SetDefault`

### Persistent root flags bound to viper

Only the persistent root flags listed in `cmd/root.go::init()` are bound
explicitly to viper. We **do not** call `viper.BindPFlags(cmd.Flags())`
because that would slurp every subcommand flag (such as `--yes`,
`--if-needed`) into `viper.AllSettings()` and risk leaking transient flag
state into `drconfig.yaml`.

### Subcommand flags

Read subcommand flag values directly from cobra:

```go
yesFlag, _ := cmd.Flags().GetBool("yes")
```

If a subcommand flag also needs an environment variable override, register
the env var only with `viper.BindEnv(...)` and merge the two sources
explicitly in your handler:

```go
// cmd/dotenv/cmd.go
_ = viper.BindEnv("yes", "DATAROBOT_CLI_NON_INTERACTIVE")

// In RunE:
yesFlag, _ := cmd.Flags().GetBool("yes")
yes := yesFlag || viper.GetBool("yes")
```

This keeps the explicit `--yes` flag value out of `viper.AllSettings()`
while preserving env-var support.

## Writing back to drconfig.yaml

Never call `viper.WriteConfig()` directly. Use the allowlisted writer in
`internal/config`:

```go
// Write all allowlisted keys that are currently set in viper:
config.UpdateConfigFile()

// Or write only specific keys (recommended when the call site knows
// exactly what changed):
config.UpdateConfigFile(config.DataRobotURL)
config.UpdateConfigFile(config.DataRobotAPIKey, config.DataRobotURL)
```

`UpdateConfigFile` reads the existing YAML, overlays only the allowlisted
keys (`config.PersistableKeys`), and writes the result back. Any other
viper state &mdash; including transient flags such as `--yes`, `--verbose`,
`--debug` &mdash; is intentionally dropped.

The wrappers in the auth package (`auth.WriteConfigFileSilent`,
`auth.WriteConfigFile`) call this writer under the hood.

## Rules for new flags

When adding a new flag, decide which category it falls into:

| Category | Bind to viper? | Persist to drconfig.yaml? |
| --- | --- | --- |
| Transient (per-invocation, e.g. `--yes`, `--all`) | No | No |
| Sticky preference (e.g. `--external-editor`) | Yes (root only) | Yes &mdash; add to `PersistableKeys` |
| Connection credential (e.g. `--token`) | Yes | Yes |

For transient flags:

- Define with `cmd.Flags().Bool(...)`
- Read with `cmd.Flags().GetBool(...)`
- Do **not** call `viper.BindPFlag(...)`

## Rules for new env vars

`viper.AutomaticEnv()` with prefix `DATAROBOT_CLI` is enabled in
`initializeConfig`, so any key you `viper.Get` will already check
`DATAROBOT_CLI_<KEY>` (with `-` replaced by `_`).

For env vars that should map to a different name (e.g.
`DATAROBOT_CLI_NON_INTERACTIVE` → key `yes`), use `viper.BindEnv` and read
the merged value as shown in [Subcommand flags](#subcommand-flags).

## Rules for new persisted keys

To make a key writable to `drconfig.yaml`:

1. Add the key to `PersistableKeys` in `internal/config/write.go`
2. Update its production write call sites to pass the key explicitly:
   `config.UpdateConfigFile("my-new-key")`
3. Add a regression test under `internal/auth/writeConfig_test.go` (or a
   dedicated test file) verifying the key round-trips correctly and that
   transient flags still do not leak.

Use viper dotted-path notation if the value is nested, e.g.
`"pulumi.config.passphrase"`. The writer handles nested map creation.

## Common pitfalls

- **Don't call `viper.WriteConfig()`** &mdash; it serializes
  `viper.AllSettings()`, which includes any flag that has ever been bound.
  Use `config.UpdateConfigFile(...)` instead.
- **Don't add `viper.BindPFlags(cmd.Flags())` anywhere.** It bulk-binds
  every flag including transient ones. Bind only the specific persistent
  flags you mean to expose via viper.
- **Don't read transient flags through viper.** `viper.GetBool("yes")`
  hides whether the value came from a flag, an env var, or a stale
  drconfig entry. Read flags directly from cobra and merge in env vars
  explicitly.
- **Don't write to `drconfig.yaml` from tests** without `viper.Reset()`
  and a temp `XDG_CONFIG_HOME`. See `internal/auth/writeConfig_test.go`
  for the recommended test pattern.

## See also

- [Flag development guide](flags.md)
- [Authentication flow](authentication.md)
- [User configuration guide](../user-guide/configuration.md)
