// This script is used to take an existing spritesheet and chop it up into
// individual sprites. It supports two modes:
//
//  1. JSON mode: Pass a .json file that describes the spritesheet (filename,
//     columns, rows, outline, padding, and a list of Pokemon). Used for
//     sheets like Legends: Arceus from spriters-resource.
//
//  2. SCSS mode: Pass a .scss file (e.g. toExtractFrom/pokesprite.scss). The
//     script finds pokesprite.png in the same directory and parses the SCSS
//     for width, height, and background-position of each .pkicon rule, then
//     extracts each sprite into ./images/ with the correct filenames.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Data struct {
	Filename string    `json:"filename"`
	Columns  int       `json:"columns"`
	Rows     int       `json:"rows"`
	Outline  int       `json:"outline_px_size"`
	Padding  int       `json:"padding_px_size"`
	Suffix   *string   `json:"suffix"`
	Pokemon  []Pokemon `json:"pokemon"`
}

type Pokemon struct {
	ID        int     `json:"id"`
	Form      *string `json:"form"`
	Skip      bool    `json:"skip"`
	SkipCount int     `json:"skip_count"` // default 1
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: task chop -- <filename.json|filename.scss>")
		fmt.Println("  JSON: task chop -- ./data/spritesheet.json")
		fmt.Println("  SCSS: task chop -- toExtractFrom/pokesprite.scss")
		os.Exit(1)
	}

	filename := os.Args[1]

	if strings.HasSuffix(filename, ".scss") {
		chopFromSCSS(filename)
		return
	}

	chopFromJSON(filename)
}

// chopFromSCSS reads a pokesprite.scss and pokesprite.png from the same
// directory and extracts each sprite into ./images/.
func chopFromSCSS(scssPath string) {
	dir := filepath.Dir(scssPath)
	pngPath := filepath.Join(dir, "pokesprite.png")

	raw, err := ioutil.ReadFile(scssPath)
	if err != nil {
		panic(err)
	}
	imgFile, err := os.Open(pngPath)
	if err != nil {
		panic(fmt.Errorf("open spritesheet %s: %w", pngPath, err))
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	// Match lines like: .pkicon.pkicon-001.color-shiny { width: 20px; height: 19px; background-position: 0px -56px; }
	// or .pkicon.pkicon-ball-love { width: 18px; height: 18px; background-position: 0px 0px; }
	lineRE := regexp.MustCompile(`\.pkicon\.pkicon-([^\s{]+)\s*\{\s*width:\s*(\d+)px;\s*height:\s*(\d+)px;\s*background-position:\s*(-?\d+)px\s*(-?\d+)px;`)
	lines := strings.Split(string(raw), "\n")

	for _, line := range lines {
		matches := lineRE.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		selectorParts := matches[1]
		w := atoi(matches[2])
		h := atoi(matches[3])
		bgX := atoi(matches[4])
		bgY := atoi(matches[5])

		outName := scssSelectorToFilename(selectorParts)
		if outName == "" {
			continue
		}

		// CSS background-position is the offset of the sprite (negative in the sheet).
		srcX := -bgX
		srcY := -bgY
		if srcX < 0 || srcY < 0 || srcX+w > img.Bounds().Dx() || srcY+h > img.Bounds().Dy() {
			fmt.Fprintf(os.Stderr, "chop: skip %s (bounds %d,%d %dx%d outside image)\n", outName, srcX, srcY, w, h)
			continue
		}

		r := image.Rect(0, 0, w, h)
		rgba := image.NewRGBA(r)
		draw.Draw(rgba, r, img, image.Point{srcX, srcY}, draw.Src)

		outPath := filepath.Join("images", outName)
		out, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(out, rgba); err != nil {
			out.Close()
			panic(err)
		}
		out.Close()
	}
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// scssSelectorToFilename converts the middle part of a .pkicon.pkicon-XXX... selector
// to a filename matching the repo convention: 001.png, 001-shiny.png, 025-legends_arceus.png, love-ball.png, etc.
func scssSelectorToFilename(parts string) string {
	segments := strings.Split(parts, ".")
	var id string
	var form, gameFamily string
	var shiny bool
	for _, s := range segments {
		switch {
		case s == "color-shiny":
			shiny = true
		case strings.HasPrefix(s, "form-"):
			form = strings.TrimPrefix(s, "form-")
		case strings.HasPrefix(s, "game-family-"):
			gameFamily = strings.TrimPrefix(s, "game-family-")
		case strings.HasPrefix(s, "pkicon-"):
			// shouldn't appear in this segment string
		default:
			// id: "001" or "ball-love" or "025"
			id = s
		}
	}
	if id == "" {
		return ""
	}
	if id == "ball-love" {
		return "love-ball.png"
	}
	name := id
	if shiny {
		name += "-shiny"
	}
	if gameFamily != "" {
		name += "-" + gameFamily
	}
	if form != "" {
		name += "-" + form
	}
	return name + ".png"
}

func chopFromJSON(filename string) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	data := Data{}

	err = json.Unmarshal(raw, &data)
	if err != nil {
		panic(err)
	}

	spritesheet, err := os.Open(data.Filename)
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(spritesheet)
	spritesheet.Close()
	if err != nil {
		panic(err)
	}

	// height = (total height of spritesheet - ((rows + 1) * outline size)) / rows
	height := (img.Bounds().Size().Y - ((data.Rows + 1) * data.Outline)) / data.Rows
	width := (img.Bounds().Size().X - ((data.Columns + 1) * data.Outline)) / data.Columns

	spriteIndex := 0
	for _, pokemon := range data.Pokemon {
		if pokemon.Skip {
			if pokemon.SkipCount == 0 {
				pokemon.SkipCount = 1
			}
			spriteIndex += pokemon.SkipCount
			continue
		}

		// Calculate which row and column we're on based on the index.
		row := spriteIndex / data.Columns
		column := spriteIndex % data.Columns

		// Create new image data.
		r := image.Rectangle{image.Point{0, 0}, image.Point{width - 2*data.Padding, height - 2*data.Padding}}
		rgba := image.NewRGBA(r)
		draw.Draw(rgba, r.Bounds(), img, image.Point{column*height + (column+1)*data.Outline + data.Padding, row*width + (row+1)*data.Outline + data.Padding}, draw.Src)

		// Generate the new filename.
		outName := fmt.Sprintf("./images/%03d", pokemon.ID)
		if data.Suffix != nil {
			outName += "-" + *data.Suffix
		}
		if pokemon.Form != nil {
			outName += "-" + *pokemon.Form
		}
		outName += ".png"

		// Write the new chopped up png out.
		out, err := os.Create(outName)
		if err != nil {
			panic(err)
		}
		encoder := png.Encoder{}
		err = encoder.Encode(out, rgba)
		out.Close()
		if err != nil {
			panic(err)
		}

		spriteIndex++
	}
}
