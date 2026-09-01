package omarchy

import (
	"context"
	"testing"
	"time"

	"github.com/opencharly/spec/spec"
)

// fakeExec is a minimal CheckExecutor that returns canned output.
type fakeExec struct {
	stdout string
	stderr string
	exit   int
}

func (f *fakeExec) RunCapture(_ context.Context, _ string) (string, string, int, error) {
	return f.stdout, f.stderr, f.exit, nil
}
func (f *fakeExec) Kind() string { return "vm" }

type fakeCC struct{ ex *fakeExec }

func (c *fakeCC) Exec() spec.CheckExecutor { return c.ex }
func (c *fakeCC) Mode() spec.CheckRunMode   { return spec.CheckModeLive }
func (c *fakeCC) HTTPDo(context.Context, spec.CheckHTTPRequest) (spec.CheckHTTPResponse, error) {
	return spec.CheckHTTPResponse{}, nil
}
func (c *fakeCC) ResolveEndpoint(context.Context, int) (string, error) { return "", nil }
func (c *fakeCC) ResolveGraphicsEndpoint(context.Context, string) (spec.CheckGraphicsEndpoint, error) {
	return spec.CheckGraphicsEndpoint{}, nil
}
func (c *fakeCC) ResolveImageLabel(context.Context, string) (string, error) { return "", nil }
func (c *fakeCC) DialTimeout() time.Duration { return 0 }
func (c *fakeCC) Box() string                { return "" }
func (c *fakeCC) Instance() string           { return "" }
func (c *fakeCC) Distros() []string          { return nil }
func (c *fakeCC) AddBackground(int)          {}

func input(args string) map[string]any {
	return map[string]any{"args": args}
}

func TestOmarchyVerb_Reserved(t *testing.T) {
	if got := NewCheckVerb().Reserved(); got != "omarchy" {
		t.Fatalf("Reserved() = %q, want omarchy", got)
	}
}

func TestOmarchyVerb_RunsCommand(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "4.0.1-1\n", exit: 0}}
	res := v.RunVerb(context.Background(), cc, &spec.Op{PluginInput: input("version")})
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(version) = %v, want pass", res.Status)
	}
	if res.Message != "4.0.1-1" {
		t.Fatalf("RunVerb(version) message = %q, want 4.0.1-1", res.Message)
	}
}

func TestOmarchyVerb_NonZeroExitFails(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stderr: "command not found", exit: 127}}
	res := v.RunVerb(context.Background(), cc, &spec.Op{PluginInput: input("nonexistent")})
	if res.Status != spec.StatusFail {
		t.Fatalf("RunVerb(nonexistent) = %v, want fail", res.Status)
	}
}

func TestOmarchyVerb_EmptyArgsFails(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "", exit: 0}}
	res := v.RunVerb(context.Background(), cc, &spec.Op{PluginInput: input("")})
	if res.Status != spec.StatusFail {
		t.Fatalf("RunVerb(empty) = %v, want fail", res.Status)
	}
}
