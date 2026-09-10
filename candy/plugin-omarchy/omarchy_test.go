package omarchy

import (
	"context"
	"testing"
	"time"

	"github.com/opencharly/plugin-omarchy/candy/plugin-omarchy/params"
	"github.com/opencharly/spec/spec"
)

// fakeExec is a minimal CheckExecutor that returns canned output and records
// the last command it was handed (the dispatch assertions read it).
type fakeExec struct {
	stdout string
	stderr string
	exit   int
	cmd    string
}

func (f *fakeExec) RunCapture(_ context.Context, cmd string) (string, string, int, error) {
	f.cmd = cmd
	return f.stdout, f.stderr, f.exit, nil
}
func (f *fakeExec) Kind() string { return "vm" }

type fakeCC struct{ ex *fakeExec }

func (c *fakeCC) Exec() spec.CheckExecutor { return c.ex }
func (c *fakeCC) Mode() spec.CheckRunMode  { return spec.CheckModeLive }
func (c *fakeCC) HTTPDo(context.Context, spec.CheckHTTPRequest) (spec.CheckHTTPResponse, error) {
	return spec.CheckHTTPResponse{}, nil
}
func (c *fakeCC) ResolveEndpoint(context.Context, int) (string, error) { return "", nil }
func (c *fakeCC) ResolveGraphicsEndpoint(context.Context, string) (spec.CheckGraphicsEndpoint, error) {
	return spec.CheckGraphicsEndpoint{}, nil
}
func (c *fakeCC) ResolveImageLabel(context.Context, string) (string, error) { return "", nil }
func (c *fakeCC) DialTimeout() time.Duration                                { return 0 }
func (c *fakeCC) Box() string                                               { return "" }
func (c *fakeCC) Instance() string                                          { return "" }
func (c *fakeCC) Distros() []string                                         { return nil }
func (c *fakeCC) AddBackground(int)                                         {}

func input(args string) map[string]any {
	return map[string]any{"args": args}
}

func methodInput(method string, extra map[string]any) map[string]any {
	m := map[string]any{"method": method}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// opWith builds an op from the plugin input plus optional matcher/exit mods.
func opWith(pluginInput map[string]any, mods ...func(*spec.Op)) *spec.Op {
	op := &spec.Op{PluginInput: pluginInput}
	for _, m := range mods {
		m(op)
	}
	return op
}

func TestOmarchyVerb_Reserved(t *testing.T) {
	if got := NewCheckVerb().Reserved(); got != "omarchy" {
		t.Fatalf("Reserved() = %q, want omarchy", got)
	}
}

func TestOmarchyVerb_RunsCommand(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "4.0.1-1", exit: 0}}
	res := v.RunVerb(context.Background(), cc, opWith(input("version")))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(version) = %v, want pass", res.Status)
	}
}

func TestOmarchyVerb_NonZeroExitFails(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stderr: "command not found", exit: 127}}
	res := v.RunVerb(context.Background(), cc, opWith(input("nonexistent")))
	if res.Status != spec.StatusFail {
		t.Fatalf("RunVerb(nonexistent) = %v, want fail", res.Status)
	}
}

func TestOmarchyVerb_EmptyArgsFails(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "", exit: 0}}
	res := v.RunVerb(context.Background(), cc, opWith(input("")))
	if res.Status != spec.StatusFail {
		t.Fatalf("RunVerb(empty) = %v, want fail", res.Status)
	}
}

func TestOmarchyVerb_ExpectNonZeroPasses(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stderr: "no such theme", exit: 1}}
	res := v.RunVerb(context.Background(), cc, opWith(input("theme set nope"), func(op *spec.Op) {
		op.PluginInput["expect_non_zero"] = true
	}))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(expect_non_zero, exit 1) = %v, want pass", res.Status)
	}
}

func TestOmarchyVerb_ExpectNonZeroFailsOnZero(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "ok", exit: 0}}
	res := v.RunVerb(context.Background(), cc, opWith(input("theme set nope"), func(op *spec.Op) {
		op.PluginInput["expect_non_zero"] = true
	}))
	if res.Status != spec.StatusFail {
		t.Fatalf("RunVerb(expect_non_zero, exit 0) = %v, want fail", res.Status)
	}
}

func TestOmarchyCommand_Dispatch(t *testing.T) {
	cases := []struct {
		name string
		in   params.OmarchyInput
		want string
	}{
		{"cli default", params.OmarchyInput{Method: "", Args: "version"}, "omarchy version"},
		{"cli explicit", params.OmarchyInput{Method: "cli", Args: "debug --no-sudo --print"}, "omarchy debug --no-sudo --print"},
		{"shell-ping", params.OmarchyInput{Method: "shell-ping"}, "omarchy-shell shell ping"},
		{"shell-summon", params.OmarchyInput{Method: "shell-summon", Plugin: "omarchy.menu"}, "omarchy-shell shell summon omarchy.menu"},
		{"shell-summon payload", params.OmarchyInput{Method: "shell-summon", Plugin: "omarchy.menu", Payload: "{\"menu\":\"system\"}"}, "omarchy-shell shell summon omarchy.menu '{\"menu\":\"system\"}'"},
		{"shell-hide", params.OmarchyInput{Method: "shell-hide", Plugin: "omarchy.weather"}, "omarchy-shell shell hide omarchy.weather"},
		{"shell-list-plugins", params.OmarchyInput{Method: "shell-list-plugins"}, "omarchy-shell shell listPlugins"},
		{"shell-reload-config", params.OmarchyInput{Method: "shell-reload-config"}, "omarchy-shell shell reloadConfig"},
		{"shell-notifications-dismiss", params.OmarchyInput{Method: "shell-notifications-dismiss"}, "omarchy-shell notifications dismissAll"},
		{"shell-notifications-send", params.OmarchyInput{Method: "shell-notifications-send", Title: "T", Text: "B"}, "omarchy-notification-send T B"},
		{"channel-current", params.OmarchyInput{Method: "channel-current"}, "omarchy-channel-current"},
		{"default-browser", params.OmarchyInput{Method: "default-browser"}, "omarchy-default-browser"},
		{"default-terminal", params.OmarchyInput{Method: "default-terminal"}, "omarchy-default-terminal"},
		{"default-editor", params.OmarchyInput{Method: "default-editor"}, "omarchy-default-editor"},
		{"theme-current", params.OmarchyInput{Method: "theme-current"}, "omarchy-theme-current"},
		{"theme-bg-current", params.OmarchyInput{Method: "theme-bg-current"}, "omarchy-theme-bg-current"},
		{"font-current", params.OmarchyInput{Method: "font-current"}, "omarchy-font-current"},
		{"weather-location", params.OmarchyInput{Method: "weather-location", Args: "--set San Francisco 37.7749,-122.4194"}, "omarchy-weather-location --set San Francisco 37.7749,-122.4194"},
		{"version", params.OmarchyInput{Method: "version"}, "omarchy version"},
	}
	for _, c := range cases {
		got, err := omarchyCommand(c.in)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.name, err)
		}
		if got != c.want {
			t.Fatalf("%s: omarchyCommand = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestOmarchyCommand_Validation(t *testing.T) {
	cases := []struct {
		name string
		in   params.OmarchyInput
	}{
		{"cli no args", params.OmarchyInput{Method: "cli"}},
		{"summon no plugin", params.OmarchyInput{Method: "shell-summon"}},
		{"hide no plugin", params.OmarchyInput{Method: "shell-hide"}},
		{"send no title", params.OmarchyInput{Method: "shell-notifications-send", Text: "B"}},
		{"weather no args", params.OmarchyInput{Method: "weather-location"}},
		{"unknown method", params.OmarchyInput{Method: "nope"}},
	}
	for _, c := range cases {
		if _, err := omarchyCommand(c.in); err == nil {
			t.Fatalf("%s: expected an error, got none", c.name)
		}
	}
}

func TestOmarchyVerb_ShellPingDispatch(t *testing.T) {
	v := &omarchyVerb{}
	ex := &fakeExec{stdout: "ok", exit: 0}
	cc := &fakeCC{ex: ex}
	res := v.RunVerb(context.Background(), cc, opWith(methodInput("shell-ping", nil)))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(shell-ping) = %v, want pass", res.Status)
	}
	if ex.cmd != "export OMARCHY_PATH=/usr/share/omarchy; omarchy-shell shell ping" {
		t.Fatalf("RunVerb(shell-ping) cmd = %q, want the shell ping argv", ex.cmd)
	}
}

func TestOmarchyVerb_ChannelCurrentDispatch(t *testing.T) {
	v := &omarchyVerb{}
	ex := &fakeExec{stdout: "edge", exit: 0}
	cc := &fakeCC{ex: ex}
	res := v.RunVerb(context.Background(), cc, opWith(methodInput("channel-current", nil), func(op *spec.Op) {
		op.Stdout = []spec.Matcher{{Op: "contains", Value: "edge"}}
	}))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(channel-current) = %v, want pass", res.Status)
	}
	if ex.cmd != "export OMARCHY_PATH=/usr/share/omarchy; omarchy-channel-current" {
		t.Fatalf("RunVerb(channel-current) cmd = %q, want the channel argv", ex.cmd)
	}
}
