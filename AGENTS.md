# AGENTS.md — plugin-omarchy

Standalone plugin repo for the `omarchy:` check verb (`verb:omarchy`) — the omarchy
command center dispatched to the guest over the executor reverse channel. The plugin
is a Go module at `candy/plugin-omarchy/` (module path
`github.com/opencharly/plugin-omarchy/candy/plugin-omarchy`). There is **no root
`charly.yml`**, so `charly box validate` at the repo root has no project manifest to
parse; the candy is validated from a staging project that carries a `discover:`
block.

This repo has **no `skill:` entity** in its candy manifest, so there is no dedicated
owning skill projected into the marketplace corpus. The gap is recorded against
`opencharly/opencharly#291` (the batch that authors missing `skill:` entities).

Canonical files:

- `candy/plugin-omarchy/charly.yml` — the `plugin-omarchy:` candy entity (`plugin:`
  block, the four `plan:` checks).
- `candy/plugin-omarchy/omarchy.go` — the verb, the method→argv dispatch table, and
  the shared verdict pipeline.
- `candy/plugin-omarchy/schema/omarchy.cue` — the self-contained `#OmarchyInput`.
- `candy/plugin-omarchy/params/gen.go` — the `go:generate` line for the params struct.
- `candy/plugin-omarchy/omarchy_test.go` — the dispatch and validation unit tests.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:` block,
  the unified Provider model, the external (out-of-process) shape, the per-plugin
  CUE-schema contract, placement. Load before touching the verb or schema.
- `/charly-check:check` — the check orchestrator, the `check:` plan step, the
  runtime context, and the R10 sequence.
- `/charly-distros:omarchy` — the Omarchy base image whose CLI this verb drives.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...`, `go vet ./...`, `go test ./...` in `candy/plugin-omarchy/` —
  compile, vet, and run the dispatch/validation unit tests.
- There is **no root `charly.yml`**, so `charly box validate` at the repo root
  parses nothing and the org-wide candy gate
  (`opencharly/.github/.github/workflows/candy-validate.yml`) **skips cleanly**.
  The manifest is gated only from a consumer project that has this candy in scan
  range (the `eval-omarchy` corpus / the Omarchy spike bed), which is what catches
  a malformed `plugin:` block.
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- R10 consumer: a disposable Omarchy bed (the `check-omarchy-verb-spike` pod, or the
  `eval-omarchy` corpus) that composes this candy and drives the verb live.

## Modify this repo

- Edit the `plugin-omarchy:` candy entity, `omarchy.go`, and `schema/omarchy.cue`
  **together** — the schema is the single source for the `params/` struct, so a
  field change not mirrored in the schema desyncs the generated types.
- A method added to the `#OmarchyInput` enum must also be added to `omarchyCommand`'s
  dispatch switch, and the required-field validation for each method.
- Reuse the **shared** verdict pipeline (`op.ExitStatus` + `sdk.MatchAll` +
  `expect_non_zero`) — do not hand-roll an exit-0-only check.
- The plugin is **out-of-tree external** (connected out-of-process by word); do not
  describe it as compiled-in.

## Landing

- PR-only. Every change lands through a pull request; the org-required
  `charly/pr-validator` validates the diff and body and arms native auto-merge on
  PASS. Direct pushes to `main` are blocked.
- History lives in `CHANGELOG/` (written by `tag-on-merge` at merge time); the PR
  body IS the changelog.
- The authoritative rulebook is the umbrella `AGENTS.md` in
  `opencharly/opencharly` and `charly/AGENTS.md` in the charly repo. Do not
  restate its rules here.
