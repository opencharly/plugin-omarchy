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
	"fmt"
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

// omarchyCommand maps the method enum to the exact argv. The shell-ipc
// methods run the omarchy-shell IPC (the SAME surface the menu and hotkeys
// use); the cli-read methods run the omarchy-* bin commands; "cli" (default)
// runs the omarchy command center with args.
func omarchyCommand(in params.OmarchyInput) (string, error) {
	switch in.Method {
	case "", "cli":
		if strings.TrimSpace(in.Args) == "" {
			return "", fmt.Errorf("args is required for method cli (e.g. version, debug, capture screenshot fullscreen save)")
		}
		return "omarchy " + in.Args, nil
	case "shell-ping":
		return "omarchy-shell shell ping", nil
	case "shell-summon":
		if strings.TrimSpace(in.Plugin) == "" {
			return "", fmt.Errorf("plugin is required for method shell-summon")
		}
		if strings.TrimSpace(in.Payload) != "" {
			return "omarchy-shell shell summon " + in.Plugin + " '" + in.Payload + "'", nil
		}
		return "omarchy-shell shell summon " + in.Plugin, nil
	case "shell-hide":
		if strings.TrimSpace(in.Plugin) == "" {
			return "", fmt.Errorf("plugin is required for method shell-hide")
		}
		return "omarchy-shell shell hide " + in.Plugin, nil
	case "shell-list-plugins":
		return "omarchy-shell shell listPlugins", nil
	case "shell-reload-config":
		return "omarchy-shell shell reloadConfig", nil
	case "shell-notifications-dismiss":
		return "omarchy-shell notifications dismissAll", nil
	case "shell-notifications-send":
		if strings.TrimSpace(in.Title) == "" {
			return "", fmt.Errorf("title is required for method shell-notifications-send")
		}
		return "omarchy-notification-send " + in.Title + " " + in.Text, nil
	case "channel-current":
		return "omarchy-channel-current", nil
	case "default-browser":
		return "omarchy-default-browser", nil
	case "default-terminal":
		return "omarchy-default-terminal", nil
	case "default-editor":
		return "omarchy-default-editor", nil
	case "theme-current":
		return "omarchy-theme-current", nil
	case "theme-bg-current":
		return "omarchy-theme-bg-current", nil
	case "font-current":
		return "omarchy-font-current", nil
	case "weather-location":
		if strings.TrimSpace(in.Args) == "" {
			return "", fmt.Errorf("args is required for method weather-location (e.g. --set San Francisco 37.7749,-122.4194)")
		}
		return "omarchy-weather-location " + in.Args, nil
	case "version":
		return "omarchy version", nil
	default:
		return "", fmt.Errorf("unknown method %q", in.Method)
	}
}

// RunVerb runs the omarchy surface in the venue and returns the verdict.
// The method enum dispatches: "cli" (default) runs `omarchy <args>`; the
// shell-ipc methods run the omarchy-shell IPC; the cli-read methods run the
// omarchy-* bin commands. All run with OMARCHY_PATH set (the installed tree).
func (v *omarchyVerb) RunVerb(ctx context.Context, cc kit.CheckContext, op *spec.Op) kit.Result {
	var in params.OmarchyInput
	raw, merr := json.Marshal(op.PluginInput)
	if merr != nil {
		return kit.Failf("omarchy: marshal input: %s", merr.Error())
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return kit.Failf("omarchy: decode input: %s", err.Error())
	}
	cmd, err := omarchyCommand(in)
	if err != nil {
		return kit.Failf("omarchy: %s", err.Error())
	}
	stdout, stderr, exitCode, err := cc.Exec().RunCapture(ctx, "export OMARCHY_PATH=/usr/share/omarchy; "+cmd)
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
