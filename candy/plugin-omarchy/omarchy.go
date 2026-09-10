package omarchy

// omarchy.go — the omarchy CLI surface as a charly check verb. The verb
// dispatches the omarchy command center (version, debug, capture, lock,
// theme, audio, bar, battery, bluetooth, brightness, channel, clipboard,
// cmd, config, default, dev, font, menu, migrate, network, plugin, power,
// system, update, upload, voxtype, weather, webapp, windows, and more) to
// the guest over the executor reverse channel, so a check bed can assert
// and drive the installed Omarchy's own tooling.
//
// The verdict pipeline mirrors plugin-command (R3): expect_non_zero asserts
// the command FAILED (any non-zero exit); otherwise the exact exit code
// (op.ExitStatus, default 0); then the shared stdout/stderr matchers
// (sdk.MatchAll). The step-level exit_status/stdout/stderr modifiers ride
// the base #Op — the verb never re-implements them.

import (
	"context"
	"embed"
	"encoding/json"
	"strings"

	"github.com/opencharly/plugin-omarchy/candy/plugin-omarchy/params"
	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// NewCheckVerb returns the omarchy check verb provider.
func NewCheckVerb() kit.CheckVerbProvider {
	return &omarchyVerb{}
}

type omarchyVerb struct{}

// Reserved is the verb word: 'omarchy: <args>'.
func (v *omarchyVerb) Reserved() string { return "omarchy" }

// RunVerb runs 'omarchy <args>' in the venue and returns the verdict.
func (v *omarchyVerb) RunVerb(ctx context.Context, cc kit.CheckContext, op *spec.Op) kit.Result {
	var in params.OmarchyInput
	raw, merr := json.Marshal(op.PluginInput)
	if merr != nil {
		return kit.Failf("omarchy: marshal input: %s", merr.Error())
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return kit.Failf("omarchy: decode input: %s", err.Error())
	}
	if strings.TrimSpace(in.Args) == "" {
		return kit.Failf("omarchy: args is required (e.g. version, debug, capture screenshot fullscreen save)")
	}
	stdout, stderr, exitCode, err := cc.Exec().RunCapture(ctx, "export OMARCHY_PATH=/usr/share/omarchy; omarchy "+in.Args)
	if err != nil {
		return kit.Failf("omarchy: exec: %s", err.Error())
	}
	// expect_non_zero asserts the command FAILED (any non-zero code) and IGNORES
	// exit_status — the two are mutually-exclusive intents (any-non-zero vs
	// exact-code). Otherwise assert the exact code (exit_status, default 0).
	// Mirrors plugin-command.
	if in.ExpectNonZero {
		if exitCode == 0 {
			return kit.Failf("omarchy %s: expected non-zero exit, got 0 (stdout: %s)", in.Args, sdk.Preview(stdout))
		}
	} else {
		wantExit := 0
		if op.ExitStatus != nil {
			wantExit = *op.ExitStatus
		}
		if exitCode != wantExit {
			return kit.Failf("omarchy %s: exit=%d, want %d (stderr: %s)", in.Args, exitCode, wantExit, sdk.Preview(stderr))
		}
	}
	if err := sdk.MatchAll(stdout, op.Stdout); err != nil {
		return kit.Failf("omarchy %s: stdout: %v (got: %s)", in.Args, err, sdk.Preview(stdout))
	}
	if err := sdk.MatchAll(stderr, op.Stderr); err != nil {
		return kit.Failf("omarchy %s: stderr: %v (got: %s)", in.Args, err, sdk.Preview(stderr))
	}
	return kit.Passf("omarchy %s: exit=%d", in.Args, exitCode)
}

// SchemaFS is the embedded CUE schema for the omarchy verb.
//
//go:embed schema/*.cue
var SchemaFS embed.FS

// SchemaDir is the schema directory within SchemaFS.
const SchemaDir = "schema"

// InputDefs names the CUE input definition for the verb.
const InputDefs = "#OmarchyInput"

// NewMeta returns the plugin meta server (calver + capability + schema).
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta("2026.253.1500", []sdk.ProvidedCapability{
		{Class: "verb", Word: "omarchy", InputDef: InputDefs, Primary: "args"},
	}, SchemaFS)
}
