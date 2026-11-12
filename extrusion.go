package gcodefun

import (
	"sync"
)

var extrusionCmds = sync.OnceValue(func() map[string]struct{} {
	return map[string]struct{}{
		"G0": {},
		"G1": {},
		"G2": {},
		"G3": {},
	}
})

func (g *GCode) Extrusion() (float64, error) {
	cmds := extrusionCmds()
	var result float64
	for _, cmd := range g.Lines {
		if _, ok := cmds[cmd.CommandID]; !ok {
			continue
		}
		flags, err := cmd.ArgMap()
		if err != nil {
			return 0, err
		}
		if amount, ok := flags["E"]; ok {
			result += amount
		}
	}
	return result, nil
}
