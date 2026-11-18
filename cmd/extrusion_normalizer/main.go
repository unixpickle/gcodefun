package main

import (
	"flag"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/gcodefun"
	"github.com/unixpickle/model3d/model3d"
)

func main() {
	var inputPath string
	var outputPath string
	var extraVolume float64
	flag.StringVar(&inputPath, "input-path", "", "path to input G-code file")
	flag.StringVar(&outputPath, "output-path", "", "path to output mesh file")
	flag.Float64Var(&extraVolume, "extra-volume", 100, "extra volume to smooth out shape")
	flag.Parse()

	if inputPath == "" || outputPath == "" {
		essentials.Die("Missing -input-path or -output-path argument. See -help.")
	}

	contents, err := os.ReadFile(inputPath)
	essentials.Must(err)

	parsed := gcodefun.ParseFile(string(contents))
	layers := gcodefun.PrintLayers(parsed)

	diamExpr := regexp.MustCompile("; filament_diameter: ([0-9\\.]*),([0-9\\.]*)")
	layerHeightExpr := regexp.MustCompile("; layer_height = ([0-9\\.]*)")
	filDiameter := 0.0
	layerHeight := 0.0
	for _, line := range parsed.Lines {
		if match := diamExpr.FindStringSubmatch(line.Raw); match != nil {
			d, err := strconv.ParseFloat(match[1], 64)
			essentials.Must(err)
			filDiameter = d
		} else if match := layerHeightExpr.FindStringSubmatch(line.Raw); match != nil {
			h, err := strconv.ParseFloat(match[1], 64)
			essentials.Must(err)
			layerHeight = h
		}
	}
	if filDiameter == 0 {
		essentials.Die("could not find filament diameter annotation in G-code")
	}
	if layerHeight == 0 {
		essentials.Die("could not find layer height")
	}

	layerVolumes := []float64{}
	for _, layerCode := range layers {
		extrusion, err := layerCode.Extrusion()
		essentials.Must(err)
		layerVolumes = append(layerVolumes, filDiameter*math.Pi*extrusion)
	}

	maxVolume := slices.Max(layerVolumes)
	sideForVolume := func(v float64) float64 {
		fillerAmount := maxVolume - v
		return math.Sqrt(fillerAmount) + extraVolume
	}
	mesh := model3d.NewMesh()
	for zIdx, layerVolume := range layerVolumes {
		z := float64(zIdx) * layerHeight
		s := sideForVolume(layerVolume)
		mesh.AddMesh(model3d.NewMeshRect(
			model3d.XYZ(-s/2, -s/2, z),
			model3d.XYZ(s/2, s/2, z+layerHeight),
		))
	}
	log.Println("exporting STL...")
	mesh.SaveGroupedSTL(outputPath)
}
