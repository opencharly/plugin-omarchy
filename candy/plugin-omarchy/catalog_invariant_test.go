package omarchy

// catalog_invariant_test.go pins the invariant the jetkvm, cua, punktfunk and gh
// siblings keep and this plugin needs for the same reason: every method the SCHEMA
// advertises must be reachable — routed by omarchyCommand — and every method
// omarchyCommand routes must be in the schema. Without it a method can sit in the
// catalog as a phantom (advertised but unimplemented): the schema validates the
// authored `omarchy: {...}`, and the step then fails only at dispatch with
// `unknown method`, so the catalog misleads the agent that dispatches on it.
//
// The tests read the REAL sources — the SERVED enum (through the CUE API,
// schema_served_test.go) and omarchy.go's switch — rather than restating either, so
// they cannot drift alongside the code they guard.

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"cuelang.org/go/cue"

	"github.com/opencharly/plugin-omarchy/candy/plugin-omarchy/params"
)

var (
	// caseArm matches one `case "a", "b":` arm. omarchyCommand's cli arm is written
	// `case "", "cli":` — the empty string is the Go ZERO VALUE of an omitted
	// method (charly decodes an absent `method` to ""), not a schema word, so the
	// scanner keeps it apart from the routed word set instead of miscounting it.
	caseArm = regexp.MustCompile(`(?m)^[ \t]*case[ \t]+((?:"(?:[^"\\]|\\.)*"[ \t]*,?[ \t]*)+):`)
	quoted  = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

// schemaMethods returns the sorted method names the SERVED #OmarchyInput.method enum
// accepts. It reads the schema through the CUE API rather than its text: the enum is
// now §10.1 S3 (a vocabulary table plus a derived word list), and a text parse of the
// union would find ZERO methods and leave every invariant below passing vacuously.
func schemaMethods(t *testing.T) []string {
	t.Helper()
	return schemaWordList(t, servedSchema(t), "#OmarchyInput.method")
}

// dispatchArms returns every `case` arm of omarchyCommand: the words it routes, and
// the arms that ALSO accept the empty string (the decoded value of an omitted
// method). The function is scoped out of omarchy.go first, so no other switch in the
// file is mistaken for dispatch.
func dispatchArms(t *testing.T) (routed []string, zeroValueArms [][]string) {
	t.Helper()
	b, err := os.ReadFile("omarchy.go")
	if err != nil {
		t.Fatalf("read the verb: %v", err)
	}
	fn := regexp.MustCompile(`(?s)func omarchyCommand\(.*?\n\}`).FindString(string(b))
	if fn == "" {
		t.Fatal("could not locate omarchyCommand in omarchy.go")
	}
	seen := map[string]bool{}
	for _, arm := range caseArm.FindAllStringSubmatch(fn, -1) {
		var words []string
		for _, q := range quoted.FindAllStringSubmatch(arm[1], -1) {
			if q[1] == "" {
				continue
			}
			words = append(words, q[1])
			seen[q[1]] = true
		}
		if strings.Contains(arm[1], `""`) {
			zeroValueArms = append(zeroValueArms, words)
		}
	}
	routed = make([]string, 0, len(seen))
	for w := range seen {
		routed = append(routed, w)
	}
	sort.Strings(routed)
	return routed, zeroValueArms
}

// routedMethods is the word half of dispatchArms.
func routedMethods(t *testing.T) []string {
	t.Helper()
	routed, _ := dispatchArms(t)
	return routed
}

// TestCatalogHasNoPhantomMethods asserts every schema method is dispatched — no
// method is advertised that the verb cannot act on.
func TestCatalogHasNoPhantomMethods(t *testing.T) {
	reachable := map[string]bool{}
	for _, m := range routedMethods(t) {
		reachable[m] = true
	}
	if len(reachable) == 0 {
		t.Fatal("the dispatch surface is empty — the invariant would pass over nothing")
	}
	for _, m := range schemaMethods(t) {
		if !reachable[m] {
			t.Errorf("schema method %q is a PHANTOM: omarchyCommand does not route it — add an implementation or remove it from the schema", m)
		}
	}
}

// TestDispatchHasNoPhantomArms is the other direction: an arm for a word the schema
// no longer accepts is dead code, and it hides a rename that only half landed.
func TestDispatchHasNoPhantomArms(t *testing.T) {
	advertised := map[string]bool{}
	for _, m := range schemaMethods(t) {
		advertised[m] = true
	}
	for _, m := range routedMethods(t) {
		if !advertised[m] {
			t.Errorf("omarchyCommand routes %q, which the served #OmarchyInput.method enum does NOT accept — the arm is unreachable", m)
		}
	}
}

// TestEverySchemaMethodIsRouted is the two-way set equality in one assertion, so a
// reader sees the whole invariant at once.
func TestEverySchemaMethodIsRouted(t *testing.T) {
	schema := schemaMethods(t)
	routed := routedMethods(t)
	if len(schema) != len(routed) {
		t.Fatalf("served enum has %d methods but omarchyCommand routes %d — the catalog and the dispatch have drifted", len(schema), len(routed))
	}
	for i := range schema {
		if schema[i] != routed[i] {
			t.Errorf("served method %q has no matching routed arm (routed=%v)", schema[i], routed)
			break
		}
	}
}

// TestOmittedMethodRoutesToTheSchemaDefault pins the `case "", …` arm against the
// schema default. The empty string is not a word of the enum — it is what an omitted
// `method` decodes to — so the arm that carries it must carry exactly the word the
// schema defaults to. Otherwise `{}` and `{method: "cli"}` would be the same document
// to CUE and two different commands to the verb.
func TestOmittedMethodRoutesToTheSchemaDefault(t *testing.T) {
	_, zeroValueArms := dispatchArms(t)
	if len(zeroValueArms) == 0 {
		t.Fatal("no case arm accepts the empty method — an omitted `method` would fail dispatch instead of taking the default")
	}
	v := servedSchema(t)
	d, ok := v.LookupPath(cue.ParsePath("#OmarchyInput.method")).Default()
	if !ok {
		t.Fatal("served #OmarchyInput.method has NO default")
	}
	want, err := d.String()
	if err != nil {
		t.Fatalf("served default is not a string: %v", err)
	}
	for _, arm := range zeroValueArms {
		found := false
		for _, w := range arm {
			if w == want {
				found = true
			}
		}
		if !found {
			t.Errorf("the arm that accepts an omitted method routes %v, which does not include the schema default %q", arm, want)
		}
	}
}

// TestEverySchemaMethodBuilds is the strongest half: every advertised method must
// resolve through omarchyCommand. A word routed to a builder that rejects it is
// advertised and broken in one step — the exact split this test exists to make
// impossible. The input fills every field the per-method preconditions require, so a
// method's authoring requirements cannot mask a missing arm.
func TestEverySchemaMethodBuilds(t *testing.T) {
	for _, m := range schemaMethods(t) {
		in := params.OmarchyInput{
			Method:  m,
			Args:    "--set San Francisco 37.7749,-122.4194",
			Plugin:  "omarchy.menu",
			Payload: `{"menu":"system"}`,
			Title:   "T",
			Text:    "B",
		}
		if _, err := omarchyCommand(in); err != nil {
			t.Errorf("schema method %q is routed to the verb but omarchyCommand rejects it: %v", m, err)
		}
	}
}
