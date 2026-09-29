# plugin-omarchy

The out-of-tree charly plugin serving the **`omarchy:` check verb** — the omarchy
command center as a declarative probe, so a check bed can assert and drive an
installed Omarchy's own tooling.

## What it drives

The verb dispatches to the guest over the executor reverse channel, running the
**same commands the Omarchy menu and hotkeys use**:

- **`cli`** (default) — the omarchy command center with `args` (`version`, `debug`,
  `capture`, `theme`, `audio`, `bar`, `bluetooth`, `network`, `power`, …).
- **shell-IPC methods** — the `omarchy-shell` IPC: `shell-ping`, `shell-summon`,
  `shell-hide`, `shell-list-plugins`, `shell-reload-config`,
  `shell-notifications-dismiss`, `shell-notifications-send`.
- **cli-read methods** — the `omarchy-*` bin commands: `channel-current`,
  `default-browser`, `default-terminal`, `default-editor`, `theme-current`,
  `theme-bg-current`, `font-current`, `weather-location`, `version`.

All run with `OMARCHY_PATH=/usr/share/omarchy` (the installed tree).

## The verdict pipeline

The verb uses the **shared** verdict pipeline (mirroring `plugin-command`), not a
hand-rolled check:

- `expect_non_zero: true` asserts the command **failed** (any non-zero exit) — the
  CLI-rejects-X class.
- otherwise the exact exit code (`exit_status`, default `0`).
- then the step-level `stdout` / `stderr` matchers via `sdk.MatchAll`.

## How to use it

It is an **external (out-of-tree) plugin**: projects compose it via the
`@github.com/opencharly/plugin-omarchy/candy/plugin-omarchy:<ref>` candy ref and
charly connects it out-of-process by word at runtime.

```yaml
- '@github.com/opencharly/plugin-omarchy/candy/plugin-omarchy:<tag>'
```

Then author the verb in a plan:

```yaml
- check: the omarchy verb reports the installed version
  id: omarchy-version
  context: [runtime]
  omarchy:
      args: version
  stdout:
      - matches: '^[0-9]+.[0-9]+.[0-9]+'

- check: the omarchy verb asserts the CLI's failure behavior
  id: omarchy-rejects-unknown
  context: [runtime]
  omarchy:
      args: definitely-not-a-command
      expect_non_zero: true
  stderr:
      - contains: Unknown Omarchy command

- check: the omarchy verb pings the shell IPC
  id: omarchy-shell-ping
  context: [runtime]
  omarchy:
      method: shell-ping
  stdout:
      - contains: ok
```

The scalar shorthand (`omarchy: version`) is declared via the capability's
`Primary: "args"`; the map form is always available.

## Layout

- `candy/plugin-omarchy/charly.yml` — the `plugin-omarchy:` candy entity (`plugin:`
  block, the four `plan:` checks).
- `candy/plugin-omarchy/omarchy.go` — the verb, the method→argv dispatch table, and
  the verdict pipeline.
- `candy/plugin-omarchy/schema/omarchy.cue` — the self-contained `#OmarchyInput`
  (the method enum + fields).
- `candy/plugin-omarchy/params/cue_types_gen.go` — the generated params struct
  (`params/gen.go` carries the `go:generate` line).
- `candy/plugin-omarchy/omarchy_test.go` — the dispatch and validation unit tests.
- `candy/plugin-omarchy/cmd/serve/main.go` — the out-of-process entrypoint.
- `CHANGELOG/` — per-CalVer release history.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: none yet — `/charly-internals:plugin` is the authoring reference for
  the plugin surface, and `/charly-check:check` the check orchestrator. The missing
  `skill:` entity is recorded against
  [opencharly/opencharly#291](https://github.com/opencharly/opencharly/issues/291).
- `/charly-distros:omarchy` — the Omarchy base image whose CLI this verb drives.
- [`eval-omarchy`](https://github.com/opencharly/eval-omarchy) — the acceptance
  corpus that composes this plugin.
- [`opencharly/opencharly`](https://github.com/opencharly/opencharly) — the umbrella.
