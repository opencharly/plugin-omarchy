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

func TestOmarchyVerb_ExitStatusMatcher(t *testing.T) {
	v := &omarchyVerb{}
	want := 1
	// exit 1 with exit_status: 1 → pass
	cc := &fakeCC{ex: &fakeExec{stderr: "boom", exit: 1}}
	res := v.RunVerb(context.Background(), cc, opWith(input("plugin validate"), func(op *spec.Op) {
		op.ExitStatus = &want
	}))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(exit_status 1, exit 1) = %v, want pass", res.Status)
	}
	// exit 0 with exit_status: 1 → fail
	cc2 := &fakeCC{ex: &fakeExec{stdout: "ok", exit: 0}}
	res2 := v.RunVerb(context.Background(), cc2, opWith(input("plugin validate"), func(op *spec.Op) {
		op.ExitStatus = &want
	}))
	if res2.Status != spec.StatusFail {
		t.Fatalf("RunVerb(exit_status 1, exit 0) = %v, want fail", res2.Status)
	}
}

func TestOmarchyVerb_StdoutMatcher(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stdout: "4.0.1-1", exit: 0}}
	res := v.RunVerb(context.Background(), cc, opWith(input("version"), func(op *spec.Op) {
		op.Stdout = spec.MatcherList{{Op: "contains", Value: "4.0"}}
	}))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(stdout contains 4.0) = %v, want pass", res.Status)
	}
	cc2 := &fakeCC{ex: &fakeExec{stdout: "5.0.0", exit: 0}}
	res2 := v.RunVerb(context.Background(), cc2, opWith(input("version"), func(op *spec.Op) {
		op.Stdout = spec.MatcherList{{Op: "contains", Value: "4.0"}}
	}))
	if res2.Status != spec.StatusFail {
		t.Fatalf("RunVerb(stdout contains 4.0, got 5.0) = %v, want fail", res2.Status)
	}
}

func TestOmarchyVerb_StderrMatcher(t *testing.T) {
	v := &omarchyVerb{}
	cc := &fakeCC{ex: &fakeExec{stderr: "no such theme: nope", exit: 1}}
	res := v.RunVerb(context.Background(), cc, opWith(input("theme set nope"), func(op *spec.Op) {
		op.PluginInput["expect_non_zero"] = true
		op.Stderr = spec.MatcherList{{Op: "contains", Value: "no such theme"}}
	}))
	if res.Status != spec.StatusPass {
		t.Fatalf("RunVerb(stderr contains 'no such theme') = %v, want pass", res.Status)
	}
}
