#!/bin/sh
# gen.sh — regenerate params/cue_types_gen.go from the ONE source,
# ../schema/*.cue.
#
# The served schema MUST stay package-less (the host splices it onto its base as
# `base ++ plugin`; a package clause in the appended body would be a syntax
# error). `cue exp gengotypes` can only generate from a NAMED package, so this
# script stages a copy with the params package clause and generates from that copy
# in its own directory. A second run is a byte-identical no-op.
#
# This is the SAME mechanism the org's candy-validate projection check runs
# (opencharly/.github/.github/workflows/candy-validate.yml: stage `package params`,
# concatenate the schema, `cue exp gengotypes .`), so `go generate ./...` and that
# gate agree on what the committed projection must be.
#
# The generation happens in a FRESH EMPTY directory, and the only write to
# params/ is the final `mv`. Two properties follow, and they are the reason for
# the shape: the output's existence is a real proof that cue ran (nothing was
# there before), and a generator failure leaves the committed projection
# untouched instead of half-written.
set -eu

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

{
	echo "package params"
	echo
	cat "$here"/../schema/*.cue
} >"$tmp/schema.cue"

( cd "$tmp" && cue exp gengotypes schema.cue )

# The generator names its output from the INPUT it was handed — `cue_types_<pkg>_gen.go`
# for a package, `cue_types_gen.go` for a single file, as the command's own help states
# ("the package name is omitted from the filename if it is implied by the import
# path"). This script does not depend on which: it requires that the fresh dir hold
# EXACTLY ONE generated file and installs that one, so a change to the naming rule
# cannot turn the guard below into a vacuous pass.
set -- "$tmp"/cue_types*_gen.go
if [ ! -f "$1" ]; then
	echo "gen.sh: cue emitted no cue_types*_gen.go in its output dir" >&2
	exit 1
fi
if [ "$#" -ne 1 ]; then
	echo "gen.sh: cue emitted $# generated files, want exactly one: $*" >&2
	exit 1
fi
gofmt -w "$1"
mv "$1" "$here/cue_types_gen.go"
