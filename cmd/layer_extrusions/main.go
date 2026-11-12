package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/gcodefun"
)

func main() {
	var inputPath string
	flag.StringVar(&inputPath, "input-path", "", "path to input G-code file")
	flag.Parse()

	if inputPath == "" {
		essentials.Die("Missing -input-path argument. See -help.")
	}

	contents, err := os.ReadFile(inputPath)
	essentials.Must(err)

	parsed := gcodefun.ParseFile(string(contents))
	layers := gcodefun.PrintLayers(parsed)

	for i, layerCode := range layers {
		extrusion, err := layerCode.Extrusion()
		essentials.Must(err)
		fmt.Printf("%d: %f", i, extrusion)
		fmt.Println()
	}
}
