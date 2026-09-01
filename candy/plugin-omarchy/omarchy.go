package omarchy

// omarchy.go — the omarchy CLI surface as a charly check verb. The verb
// dispatches the omarchy command center (version, debug, capture, lock,
// theme, audio, bar, battery, bluetooth, brightness, channel, clipboard,
// cmd, config, default, dev, font, menu, migrate, network, plugin, power,
// system, update, upload, voxtype, weather, webapp, windows, and more) to
// the guest over the executor reverse channel, so a check bed can assert
// and drive the installed Omarchy's own tooling.

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

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

// Reserved is the verb word: `omarchy: <args>`.
func (v *omarchyVerb) Reserved() string { return "omarchy" }

// RunVerb runs `omarchy <args>` in the venue and returns the result.
func (v *omarchyVerb) RunVerb(ctx context.Context, cc kit.CheckContext, op *spec.Op) spec.CheckVerbResult {
	var in struct {
		Args string `json:"args"`
	}
	raw, merr := json.Marshal(op.PluginInput)
	if merr != nil {
		return spec.CheckVerbResult{Status: spec.StatusFail, Message: "omarchy: marshal input: " + merr.Error()}
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return spec.CheckVerbResult{Status: spec.StatusFail, Message: "omarchy: decode input: " + err.Error()}
	}
	if strings.TrimSpace(in.Args) == "" {
		return spec.CheckVerbResult{Status: spec.StatusFail, Message: "omarchy: args is required (e.g. version, debug, capture screenshot fullscreen save)"}
	}
	stdout, stderr, exit, err := cc.Exec().RunCapture(ctx, "omarchy "+in.Args)
	if err != nil {
		return spec.CheckVerbResult{Status: spec.StatusFail, Message: "omarchy: exec: " + err.Error()}
	}
	if exit != 0 {
		return spec.CheckVerbResult{Status: spec.StatusFail, Message: fmt.Sprintf("omarchy %s: exit %d: %s", in.Args, exit, strings.TrimSpace(stderr))}
	}
	return spec.CheckVerbResult{Status: spec.StatusPass, Message: strings.TrimSpace(stdout)}
}

// SchemaFS is the embedded CUE schema for the omarchy verb.
//go:embed schema/*.cue
var SchemaFS embed.FS

// SchemaDir is the schema directory within SchemaFS.
const SchemaDir = "schema"

// InputDefs names the CUE input definition for the verb.
const InputDefs = "#OmarchyInput"

// NewMeta returns the plugin meta server (calver + capability + schema).
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta("2026.244.1400", []sdk.ProvidedCapability{
		{Class: "verb", Word: "omarchy", InputDef: InputDefs},
	}, SchemaFS)
}
