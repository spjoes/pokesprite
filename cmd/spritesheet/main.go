// This script looks at the `images` directory and stitches all the images into
// a single spritesheet. If there are new images copied from
// https://github.com/msikma/pokesprite that are not in the `images` directory,
// you should run the `rename` script first. It will put the love ball sprite on
// the first row, and then put all Pokemon on subsequent rows. The `height` and
// `width` variables should be the max height and width possible. If a new
// generate produces larger sprites, you should update those values. While this
// script produces a PNG with the best compression by default. You can choose
// WebP output with `-format webp`.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/pokedextracker/pokesprite/pkg/size"
	"github.com/pokedextracker/pokesprite/pkg/sorter"
)

const (
	columns = 32
)

var (
	height          = 0
	width           = 0
	progressLineLen = 0
)

func main() {
	format := flag.String("format", "png", "output format: png or webp")
	flag.Parse()

	files, err := ioutil.ReadDir("./images")
	if err != nil {
		panic(err)
	}

	height, width, err = size.Max(files)
	if err != nil {
		panic(err)
	}

	// Minus 1 to exclude the love ball, and plus 1 to include it again since it
	// will be on its own line.
	rows := int(math.Ceil(float64(len(files)-1)/columns)) + 1
	r := image.Rectangle{image.Point{0, 0}, image.Point{columns * width, rows * height}}
	rgba := image.NewRGBA(r)

	// Draw the love ball on its own row.
	totalToDraw := len(files)
	drawn := 0
	printProgress(drawn, totalToDraw, "starting")

	err = drawImage(rgba, "love-ball.png", 0, 0)
	if err != nil {
		panic(err)
	}
	drawn++
	printProgress(drawn, totalToDraw, "love-ball.png")

	// Sort files alphabetically.
	sort.Sort(sorter.New(files))

	for i, file := range files {
		name := file.Name()
		// Skip drawing the love ball since we already drew it on its own row.
		if name == "love-ball.png" {
			continue
		}

		column := int(math.Mod(float64(i), float64(columns)))
		row := i/columns + 1

		err := drawImage(rgba, name, column, row)
		if err != nil {
			panic(err)
		}
		drawn++
		printProgress(drawn, totalToDraw, name)
	}
	fmt.Println()

	outputPath := getOutputPath(*format)
	fmt.Printf("Preparing to encode %s\n", outputPath)

	switch strings.ToLower(*format) {
	case "png":
		err = withSpinner("Encoding PNG", func() error {
			out, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer out.Close()

			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			return encoder.Encode(out, rgba)
		})
	case "webp":
		err = encodeWebP(rgba, outputPath)
	default:
		err = fmt.Errorf("unsupported format %q, expected png or webp", *format)
	}

	if err != nil {
		panic(err)
	}
}

func getOutputPath(format string) string {
	switch strings.ToLower(format) {
	case "webp":
		return "./output/pokesprite.webp"
	default:
		return "./output/pokesprite.png"
	}
}

func encodeWebP(rgba image.Image, outputPath string) error {
	if _, err := exec.LookPath("cwebp"); err != nil {
		return fmt.Errorf("cwebp not found in PATH; install libwebp/cwebp to enable -format webp")
	}

	// cwebp only accepts image files as input, so encode a temporary PNG first.
	tempPNG, err := os.CreateTemp("./output", "pokesprite-*.png")
	if err != nil {
		return err
	}

	tempPath := tempPNG.Name()
	defer os.Remove(tempPath)

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(tempPNG, rgba); err != nil {
		tempPNG.Close()
		return err
	}
	if err := tempPNG.Close(); err != nil {
		return err
	}

	cmd := exec.Command("cwebp", "-lossless", "-z", "9", "-progress", tempPath, "-o", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := withSpinner("Encoding WebP", cmd.Run); err != nil {
		return fmt.Errorf("cwebp failed: %w", err)
	}

	return nil
}

func printProgress(current, total int, name string) {
	if total <= 0 {
		return
	}
	percent := int(math.Round((float64(current) / float64(total)) * 100))
	if percent > 100 {
		percent = 100
	}
	line := fmt.Sprintf("Drawing sprites: %d/%d (%d%%) - %s", current, total, percent, name)
	if len(line) < progressLineLen {
		line += strings.Repeat(" ", progressLineLen-len(line))
	}
	progressLineLen = len(line)
	fmt.Printf("\r%s", line)
}

func withSpinner(label string, fn func() error) error {
	done := make(chan error, 1)
	start := time.Now()
	frames := []string{"|", "/", "-", "\\"}

	go func() {
		done <- fn()
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	i := 0

	for {
		select {
		case err := <-done:
			elapsed := time.Since(start).Round(time.Second)
			if err != nil {
				fmt.Printf("\r%s failed (%s)                        \n", label, elapsed)
				return err
			}
			fmt.Printf("\r%s complete (%s)                      \n", label, elapsed)
			return nil
		case <-ticker.C:
			elapsed := time.Since(start).Round(time.Second)
			fmt.Printf("\r%s %s (%s)", label, frames[i%len(frames)], elapsed)
			i++
		}
	}
}

func drawImage(rgba draw.Image, name string, column, row int) error {
	file, err := os.Open("./images/" + name)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// wdiff := width - img.Bounds().Dx()
	// if wdiff < 0 {
	// 	return fmt.Errorf("width (%dpx) is too small for %s (%dpx)", width, name, img.Bounds().Dx())
	// }
	// hdiff := height - img.Bounds().Dy()
	// if hdiff < 0 {
	// 	return fmt.Errorf("height (%dpx) is too small for %s (%dpx)", height, name, img.Bounds().Dy())
	// }

	// woffset := wdiff / 2
	// hoffset := hdiff / 2
	x := column * width
	y := row * height
	// top, right, bottom, left := trim(img)

	// rect := image.Rectangle{image.Point{x, y}, image.Point{x + (right - left), y + (bottom - top)}}
	rect := image.Rectangle{image.Point{x, y}, image.Point{x + img.Bounds().Dx(), y + img.Bounds().Dy()}}
	draw.Draw(rgba, rect, img, image.Point{0, 0}, draw.Src)

	return nil
}

// func trim(img image.Image) (int, int, int, int) {
// 	bounds := img.Bounds()
// 	top := 0
// 	right := 0
// 	bottom := 0
// 	left := 0

// top:
// 	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
// 		for x := bounds.Min.X; x < bounds.Max.X; x++ {
// 			_, _, _, a := img.At(x, y).RGBA()
// 			transparent := a == 0
// 			if !transparent {
// 				top = y
// 				break top
// 			}
// 		}
// 	}
// right:
// 	for x := bounds.Max.X; x >= bounds.Min.X; x-- {
// 		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
// 			_, _, _, a := img.At(x, y).RGBA()
// 			transparent := a == 0
// 			if !transparent {
// 				right = x + 1
// 				break right
// 			}
// 		}
// 	}
// bottom:
// 	for y := bounds.Max.Y; y >= bounds.Min.Y; y-- {
// 		for x := bounds.Min.X; x < bounds.Max.X; x++ {
// 			_, _, _, a := img.At(x, y).RGBA()
// 			transparent := a == 0
// 			if !transparent {
// 				bottom = y + 1
// 				break bottom
// 			}
// 		}
// 	}
// left:
// 	for x := bounds.Min.X; x < bounds.Max.X; x++ {
// 		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
// 			_, _, _, a := img.At(x, y).RGBA()
// 			transparent := a == 0
// 			if !transparent {
// 				left = x
// 				break left
// 			}
// 		}
// 	}

// 	return top, right, bottom, left
// }
