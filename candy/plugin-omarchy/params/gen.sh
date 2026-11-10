#!/bin/sh
# gen.sh — regenerate params/cue_types_gen.go from the ONE source,
# ../schema/*.cue.
#
# The served schema MUST stay package-less (the host splices it onto its base as
# `base ++ plugin`; a package clause in the appended body would be a syntax
# error). `cue exp gengotypes` can only generate from a NAMED package, so this
# script stages a copy with the params package clause, generates from it, and the
# generator writes cue_types_gen.go into this directory. A second run is a
# byte-identical no-op.
#
# This is the SAME mechanism the org's candy-validate projection check runs
# (opencharly/.github/.github/workflows/candy-validate.yml: stage `package params`,
# concatenate the schema, `cue exp gengotypes .`), so `go generate ./...` and that
# gate agree on what the committed projection must be.
set -eu

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

{
	echo "package params"
	echo
	cat "$here"/../schema/*.cue
} >"$tmp/schema.cue"

cd "$here"
cue exp gengotypes "$tmp/schema.cue"

[ -f cue_types_gen.go ] || {
	echo "gen.sh: cue emitted no cue_types_gen.go" >&2
	exit 1
}
gofmt -w cue_types_gen.go
