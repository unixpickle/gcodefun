package gcodefun

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var argumentExpr = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile("^([a-zA-Z]+)([-0-9\\.]*)$")
})

// GCode represents a parsed G-code file.
type GCode struct {
	Lines []*GCodeLine
}

// ParseFile parses the commands in the file line-by-line.
func ParseFile(contents string) *GCode {
	result := &GCode{}
	for i, line := range strings.Split(contents, "\n") {
		cmd := ParseGCodeLine(i, line)
		result.Lines = append(result.Lines, cmd)
	}
	return result
}

// GCodeLine represents a single line in a G-code file.
type GCodeLine struct {
	LineNum   int
	Raw       string
	Tokens    []string
	CommandID string
}

// ParseGCodeLine returns a line which may represent a command, a blank
// line or a comment, but will never be nil.
func ParseGCodeLine(idx int, raw string) *GCodeLine {
	line := strings.TrimSpace(raw)
	if strings.Contains(line, ";") {
		line = line[:strings.Index(line, ";")]
	}
	if line == "" {
		return &GCodeLine{Raw: raw}
	}
	tokens := strings.Fields(line)
	return &GCodeLine{
		LineNum:   idx,
		Raw:       line,
		Tokens:    tokens,
		CommandID: tokens[0],
	}
}

// ArgMap parses numeric arguments, e.g. "G1 X3 Y4" into {"X": 3, "Y": 4}.
func (c *GCodeLine) ArgMap() (map[string]float64, error) {
	expr := argumentExpr()
	result := map[string]float64{}
	for i := 1; i < len(c.Tokens); i++ {
		match := expr.FindStringSubmatch(c.Tokens[i])
		if match == nil {
			return nil, fmt.Errorf("token %d is invalid: %#v", i, c.Tokens[i])
		}
		n, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return nil, fmt.Errorf("token %d is invalid: %#v: %w", i, c.Tokens[i], err)
		}
		result[match[1]] = n
	}
	return result, nil
}
