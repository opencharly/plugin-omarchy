package main

import (
	"github.com/opencharly/plugin-omarchy/candy/plugin-omarchy"
	"github.com/opencharly/sdk"
)

func main() {
	sdk.ServeCheckVerb(omarchy.NewCheckVerb(), omarchy.NewMeta())
}
