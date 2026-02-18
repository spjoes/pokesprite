// This script reads a pokesprite.scss file and generates a TypeScript file
// containing a map of sprite positions for use in a web application.
//
// Usage:
//
//	task positions                                         # uses defaults
//	task positions -- <input.scss> <output.ts>             # custom paths
//
// Defaults:
//
//	input:  ./output/pokesprite.scss
//	output: ./output/sprite-positions.ts
package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SpritePosition struct {
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	BackgroundPosition string `json:"backgroundPosition"`
}

func main() {
	scssPath := "./output/pokesprite.scss"
	outPath := "./output/sprite-positions.ts"

	switch len(os.Args) {
	case 1:
		// use defaults
	case 3:
		scssPath = os.Args[1]
		outPath = os.Args[2]
	default:
		fmt.Fprintln(os.Stderr, "Usage: task positions -- [<input.scss> <output.ts>]")
		fmt.Fprintln(os.Stderr, "Example: task positions -- output/pokesprite.scss output/sprite-positions.ts")
		os.Exit(1)
	}

	fmt.Println("Reading SCSS file...")
	raw, err := os.ReadFile(scssPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", scssPath, err)
		os.Exit(1)
	}

	fmt.Println("Extracting sprite positions...")
	positions, order := extractSpritePositions(string(raw))
	fmt.Printf("Found %d sprite positions\n", len(positions))

	fmt.Println("Writing output file...")
	if err := writeTS(outPath, positions, order); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s\n", outPath)
}

// lineRE matches lines like:
// .pkicon.pkicon-025.form-cap.game-family-legends_arceus.color-shiny { width: 21px; height: 20px; background-position: -67px -56px; }
var lineRE = regexp.MustCompile(
	`\.pkicon\.pkicon-(\d+)` +
		`(?:\.form-([a-z0-9_-]+))?` +
		`(?:\.game-family-([a-z0-9_-]+))?` +
		`(?:\.color-([a-z_-]+))?` +
		`\s*\{\s*width:\s*(\d+)px;\s*height:\s*(\d+)px;\s*background-position:\s*(-?\d+)px\s+(-?\d+)px;\s*\}`,
)

func extractSpritePositions(scss string) (map[string]SpritePosition, []string) {
	positions := make(map[string]SpritePosition)
	var order []string

	for _, line := range strings.Split(scss, "\n") {
		m := lineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		pokemonID, form, gameFamily, color := m[1], m[2], m[3], m[4]
		width := atoi(m[5])
		height := atoi(m[6])
		xPos, yPos := m[7], m[8]

		key := "pokemon-" + strconv.Itoa(atoi(pokemonID))
		if form != "" {
			key += "-" + form
		}
		if gameFamily != "" {
			key += "-" + gameFamily
		}
		if color == "shiny" {
			key += "-shiny"
		}

		positions[key] = SpritePosition{
			Width:              width,
			Height:             height,
			BackgroundPosition: xPos + "px " + yPos + "px",
		}
		order = append(order, key)
	}

	return positions, order
}

func writeTS(path string, positions map[string]SpritePosition, order []string) error {
	var b strings.Builder

	b.WriteString("// Auto-generated from pokesprite.scss\n")
	b.WriteString("// Do not edit manually\n")
	b.WriteString("\n")
	b.WriteString("import { SpritePosition } from './sprite-utils';\n")
	b.WriteString("\n")
	b.WriteString("/**\n")
	b.WriteString(" * Map of sprite position data from the spritesheet\n")
	b.WriteString(" */\n")
	b.WriteString("export const spritePositions: Record<string, SpritePosition> = {\n")

	for _, key := range order {
		p := positions[key]
		b.WriteString(fmt.Sprintf(
			"  %q: { width: %d, height: %d, backgroundPosition: %q },\n",
			key, p.Width, p.Height, p.BackgroundPosition,
		))
	}

	b.WriteString("};\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
