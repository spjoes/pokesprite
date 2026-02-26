// This script looks at the `images` directory and generates a plain .css file
// with sprite position rules. Each rule includes CSS custom properties --w and
// --h so the consuming application can compute scaling in pure CSS without a
// JavaScript lookup table.
//
// Usage:
//
//	task css
package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pokedextracker/pokesprite/pkg/size"
	"github.com/pokedextracker/pokesprite/pkg/sorter"
)

const (
	columns = 32

	preamble = `/* Auto-generated — do not edit manually */

`
)

var (
	height = 0
	width  = 0
)

var nameRE = regexp.MustCompile(`(\d+)(-shiny)?(-legends_arceus|-home)?(-.*)?\.png`)

func main() {
	var buf bytes.Buffer

	files, err := ioutil.ReadDir("./images")
	if err != nil {
		panic(err)
	}

	height, width, err = size.Max(files)
	if err != nil {
		panic(err)
	}

	buf.WriteString(preamble)

	// Love ball
	loveBallStyles, err := generateStyles("love-ball.png", 0, 0)
	if err != nil {
		panic(err)
	}
	buf.WriteString(fmt.Sprintf(".pkicon.pkicon-ball-love { %s }\n", loveBallStyles))

	// Sort files alphabetically.
	sort.Sort(sorter.New(files))

	for i, file := range files {
		name := file.Name()
		if name == "love-ball.png" {
			continue
		}

		matches := nameRE.FindAllStringSubmatch(name, -1)
		id := matches[0][1]
		shiny := matches[0][2] == "-shiny"
		gameFamily := strings.Trim(matches[0][3], "-")
		form := strings.Trim(matches[0][4], "-")

		class := ".pkicon.pkicon-" + id

		if form != "" {
			class += ".form-" + form
		}

		if gameFamily != "" {
			class += ".game-family-" + gameFamily
		}

		if shiny {
			class += ".color-shiny"
		}

		column := int(math.Mod(float64(i), float64(columns)))
		row := i/columns + 1

		styles, err := generateStyles(name, column, row)
		if err != nil {
			panic(err)
		}

		buf.WriteString(fmt.Sprintf("%s { %s }\n", class, styles))
	}

	err = ioutil.WriteFile("./output/pokesprite.css", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully generated output/pokesprite.css")
}

func generateStyles(name string, column, row int) (string, error) {
	file, err := os.Open("./images/" + name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	x := column * width
	y := row * height

	styles := fmt.Sprintf(
		"--w: %d; --h: %d; width: %dpx; height: %dpx; background-position: %dpx %dpx;",
		w, h, w, h, x*-1, y*-1,
	)

	return styles, nil
}
