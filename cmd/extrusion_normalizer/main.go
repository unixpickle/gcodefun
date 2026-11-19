package main

import (
	"flag"
	"fmt"
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
	prevSideLength := 0.0
	for zIdx, layerVolume := range layerVolumes {
		z := float64(zIdx) * layerHeight
		s := sideForVolume(layerVolume)

		min := model3d.XYZ(-s/2, -s/2, z)
		max := model3d.XYZ(s/2, s/2, z+layerHeight)
		sides := model3d.NewMeshRect(min, max)
		sides.Iterate(func(t *model3d.Triangle) {
			if math.Abs(t.Normal().Z) < 0.9 {
				mesh.Add(t)
			}
		})

		// Triangulate bottom from previous rect.
		if prevSideLength == 0 {
			// Add the bottom face directly.
			mesh.AddQuad(
				model3d.XYZ(min.X, min.Y, min.Z),
				model3d.XYZ(min.X, max.Y, min.Z),
				model3d.XYZ(max.X, max.Y, min.Z),
				model3d.XYZ(max.X, min.Y, min.Z),
			)
		} else if prevSideLength != s {
			// Add the interface between the last block and this one.
			smallerS := math.Min(prevSideLength, s)
			largerS := math.Max(prevSideLength, s)

			subMesh := model3d.NewMesh()
			subMesh.AddQuad(
				model3d.XYZ(-largerS/2, largerS/2, z),
				model3d.XYZ(largerS/2, largerS/2, z),
				model3d.XYZ(smallerS/2, smallerS/2, z),
				model3d.XYZ(-smallerS/2, smallerS/2, z),
			)
			subMesh.AddQuad(
				model3d.XYZ(-smallerS/2, -smallerS/2, z),
				model3d.XYZ(smallerS/2, -smallerS/2, z),
				model3d.XYZ(largerS/2, -largerS/2, z),
				model3d.XYZ(-largerS/2, -largerS/2, z),
			)
			subMesh.AddQuad(
				model3d.XYZ(-largerS/2, -largerS/2, z),
				model3d.XYZ(-largerS/2, largerS/2, z),
				model3d.XYZ(-smallerS/2, smallerS/2, z),
				model3d.XYZ(-smallerS/2, -smallerS/2, z),
			)
			subMesh.AddQuad(
				model3d.XYZ(smallerS/2, -smallerS/2, z),
				model3d.XYZ(smallerS/2, smallerS/2, z),
				model3d.XYZ(largerS/2, largerS/2, z),
				model3d.XYZ(largerS/2, -largerS/2, z),
			)

			if prevSideLength > s {
				subMesh = subMesh.InvertNormals()
			}

			mesh.AddMesh(subMesh)
		}

		prevSideLength = s
	}

	// Add the top face -- we should check the orientation.
	z := float64(len(layerVolumes)) * layerHeight
	s := prevSideLength
	mesh.AddQuad(
		model3d.XYZ(s/2, -s/2, z),
		model3d.XYZ(s/2, s/2, z),
		model3d.XYZ(-s/2, s/2, z),
		model3d.XYZ(-s/2, -s/2, z),
	)

	var n int
	_, n = mesh.RepairNormals(1e-8)
	fmt.Println("repaired", n, "normals")

	log.Println("exporting STL...")
	mesh.SaveGroupedSTL(outputPath)
}
