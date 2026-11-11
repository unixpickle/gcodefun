package gcodefun

// PrintLayers splits the file into 3D print layers, removing code
// before the first layer and after the last.
//
// Layers are identified by "M73 L<N>" where <N> is a layer index.
func PrintLayers(file *GCode) []*GCode {
	seenFirst := false
	curLayer := &GCode{}
	result := []*GCode{}
	for _, inst := range file.Lines {
		if inst.CommandID == "M73" {
			args, _ := inst.ArgMap()
			if args != nil {
				if _, ok := args["L"]; ok {
					if seenFirst {
						result = append(result, curLayer)
						curLayer = &GCode{}
					}
					seenFirst = true
				}
			}
		}
		if seenFirst {
			curLayer.Lines = append(curLayer.Lines, inst)
		}
	}
	if len(curLayer.Lines) > 0 {
		result = append(result, curLayer)
	}
	return result
}
