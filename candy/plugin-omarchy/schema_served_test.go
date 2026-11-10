package omarchy

// schema_served_test.go is the schema-LOAD gate, and the §10.1 S3 gate.
//
// `charly box validate` does NOT compile a served schema: a dangling reference
// inside a served def is green at plugin load and only surfaces when the host
// validates an authored `omarchy: {...}` against it (opencharly/spec#301). So this
// file composes the plugin's OWN Describe reply and compiles the body it serves,
// with the same CUE library the runtime links — the served body, not the file on
// disk, is the surface the host splices onto its base.
//
// It pins the S3 half of D-PAT-3 as well: every closed set of words is a
// vocabulary table, its word list is DERIVED from that table, and the enum an
// author meets is `or()` of the derived list — so a word cannot be added to the
// table and forgotten by the list, nor accepted by the enum while the table it
// claims to derive from lacks it.
//
// And it pins the ONE judgement call this sweep made: `method` carries a literal
// default (`*"cli"`), which S3 permits inline and which the derived enum must keep
// carrying. `TestTheSchemaDefaultIsTheVerbDefault` is its control.

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	pb "github.com/opencharly/spec/proto"

	"github.com/opencharly/plugin-omarchy/candy/plugin-omarchy/params"
)

// vocabHarness LIFTS the hidden vocabulary tables and their derived word lists
// into one public definition. An `_`-prefixed label is hidden by CUE and cannot
// be looked up from outside the compilation unit, so the served body alone
// exposes only the enum; this fragment adds no semantics, it re-exports the table
// the schema already carries so the test can read it instead of restating its
// words.
const vocabHarness = `
#TestVocabulary: {
	tables: {
		method: _method
	}
	words: {
		method: _methodWord
	}
}
`

// vocabProbe pairs a vocabulary table (named as in vocabHarness) with ONE authored
// input that uses a word from it, so the enum is exercised through the same def the
// host validates against.
type vocabProbe struct {
	name string
	use  func(word string) string
}

// vocabProbes covers every vocabulary table this schema carries. A new table
// belongs in vocabHarness and here: the probe is what proves it is wired into the
// enum an author actually meets.
var vocabProbes = []vocabProbe{
	{"method", func(w string) string {
		return fmt.Sprintf(`{#OmarchyInput: {method: %q}}`, w)
	}},
}

// servedSchema returns the plugin's served CUE body COMPILED, with the test's
// table re-export appended. NewMeta's Describe already compiles the concatenated
// body (sdk.BuildCapabilities), so a failure here means the schema the host
// receives does not stand on its own.
func servedSchema(t *testing.T) cue.Value {
	t.Helper()
	caps, err := NewMeta().Describe(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	body := caps.GetSchemaCue()
	if body == "" {
		t.Fatal("Describe served an EMPTY CUE body")
	}
	v := cuecontext.New().CompileString(body + vocabHarness)
	if err := v.Err(); err != nil {
		t.Fatalf("served schema does not compile: %v", err)
	}
	return v
}

// schemaWordList returns the sorted alternatives of a disjunction at path — for an
// S3 enum, exactly the words its derived list feeds.
//
// The alternatives are read from the EXPRESSION, not from the text and not through
// `String()`: a disjunction that carries a default (`*"cli" | or(_methodWord)`)
// hands `String()` the default, so probing with it would read the 17-word catalog
// as the single word "cli".
func schemaWordList(t *testing.T, v cue.Value, path string) []string {
	t.Helper()
	seen := map[string]bool{}
	collectWords(t, path, v.LookupPath(cue.ParsePath(path)), 0, seen)
	words := make([]string, 0, len(seen))
	for w := range seen {
		words = append(words, w)
	}
	sort.Strings(words)
	return words
}

func collectWords(t *testing.T, path string, val cue.Value, depth int, seen map[string]bool) {
	t.Helper()
	if err := val.Err(); err != nil {
		t.Fatalf("%s does not resolve in the served schema: %v", path, err)
	}
	if depth > 16 {
		t.Fatalf("%s: the disjunction nests deeper than 16 levels", path)
	}
	if val.IsConcrete() && val.IncompleteKind() == cue.StringKind {
		w, err := val.String()
		if err != nil {
			t.Fatalf("%s: a concrete string does not stringify: %v", path, err)
		}
		seen[w] = true
		return
	}
	_, alts := val.Expr()
	if len(alts) == 0 {
		t.Fatalf("%s has no alternatives — it is not a word list", path)
	}
	for _, alt := range alts {
		collectWords(t, path, alt, depth+1, seen)
	}
}

// TestServedSchemaResolvesItsDefs — the capability NAME is the def the host looks up
// after splicing; a served body that compiles but no longer carries it is a
// registration-time surprise, so it must resolve here.
func TestServedSchemaResolvesItsDefs(t *testing.T) {
	v := servedSchema(t)
	for _, path := range []string{"#OmarchyInput", "#OmarchyInput.method"} {
		if err := v.LookupPath(cue.ParsePath(path)).Err(); err != nil {
			t.Errorf("served schema: %s does not resolve: %v", path, err)
		}
	}
}

// TestServedCatalogCoversTheWholeEnum guards the READER and the served catalog in one
// step: the table's first and last rows must be present, so neither a truncated read
// nor a table whose tail never reaches the enum can exempt the catalog's tail.
func TestServedCatalogCoversTheWholeEnum(t *testing.T) {
	got := schemaWordList(t, servedSchema(t), "#OmarchyInput.method")
	has := func(name string) bool {
		for _, m := range got {
			if m == name {
				return true
			}
		}
		return false
	}
	if !has("cli") {
		t.Errorf("served enum is missing its HEAD (cli) — got %d methods: %v", len(got), got)
	}
	if !has("version") {
		t.Errorf("served enum is missing its TAIL (version) — the derivation is losing rows; got %d methods: %v", len(got), got)
	}
	if len(got) < 17 {
		t.Errorf("read only %d methods — the served #OmarchyInput.method derivation looks truncated", len(got))
	}
}

// TestTheSchemaDefaultIsTheVerbDefault pins the ONE judgement call of this sweep.
// `*"cli"` is S3-permitted inline, and deriving the enum must not drop it: the
// schema has to keep defaulting to "cli", and the verb has to keep treating an
// OMITTED method (the Go zero value) as that same word — otherwise the surface an
// author meets and the arm that dispatches disagree about what "no method" means.
func TestTheSchemaDefaultIsTheVerbDefault(t *testing.T) {
	v := servedSchema(t)
	d, ok := v.LookupPath(cue.ParsePath("#OmarchyInput.method")).Default()
	if !ok {
		t.Fatal("served #OmarchyInput.method has NO default — the `*\"cli\"` marker was dropped")
	}
	w, err := d.String()
	if err != nil {
		t.Fatalf("served default is not a string: %v", err)
	}
	if w != "cli" {
		t.Errorf("served default is %q, want \"cli\"", w)
	}
	omitted, err := omarchyCommand(params.OmarchyInput{Args: "version"})
	if err != nil {
		t.Fatalf("omarchyCommand(omitted method) = error %v, want the cli arm", err)
	}
	explicit, err := omarchyCommand(params.OmarchyInput{Method: w, Args: "version"})
	if err != nil {
		t.Fatalf("omarchyCommand(method %q) = error %v, want the cli arm", w, err)
	}
	if omitted != explicit {
		t.Errorf("an omitted method dispatches %q but method %q dispatches %q — the schema default and the verb's zero-value arm disagree", omitted, w, explicit)
	}
}

// TestS3VocabularyTablesDriveTheirEnums proves, for every vocabulary table, that
// (a) the derived word list IS the table's row set, and (b) each word is accepted
// where an author uses it while a non-word is refused. (b) is what a table that is
// merely DECLARED — never wired into the enum — fails.
func TestS3VocabularyTablesDriveTheirEnums(t *testing.T) {
	v := servedSchema(t)
	ctx := v.Context()
	var lifted struct {
		Tables map[string]map[string]any
		Words  map[string][]string
	}
	if err := v.LookupPath(cue.ParsePath("#TestVocabulary")).Decode(&lifted); err != nil {
		t.Fatalf("lift the vocabulary tables: %v", err)
	}
	for _, p := range vocabProbes {
		rows, ok := lifted.Tables[p.name]
		if !ok {
			t.Errorf("no vocabulary table %q in the served schema", p.name)
			continue
		}
		if len(rows) == 0 {
			t.Errorf("vocabulary table %q has no rows", p.name)
			continue
		}
		words, ok := lifted.Words[p.name]
		if !ok {
			t.Errorf("no derived word list for %q in the served schema", p.name)
			continue
		}
		keys := make([]string, 0, len(rows))
		for w := range rows {
			keys = append(keys, w)
		}
		sort.Strings(keys)
		sort.Strings(words)
		if fmt.Sprint(keys) != fmt.Sprint(words) {
			t.Errorf("%q's derived list %v is not its table's row set %v", p.name, words, keys)
		}
		for _, w := range keys {
			if err := v.Unify(ctx.CompileString(p.use(w))).Err(); err != nil {
				t.Errorf("%q word %q is refused where an author uses it: %v", p.name, w, err)
			}
		}
		if err := v.Unify(ctx.CompileString(p.use("zzz-not-a-word"))).Err(); err == nil {
			t.Errorf("%q accepts a non-word — the enum is not the table", p.name)
		}
	}
}
