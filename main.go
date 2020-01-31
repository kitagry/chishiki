//go:generate go-assets-builder -s="/data" -o image.go data
//go:generate goimports -w image.go
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/mattn/go-sixel"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/math/fixed"
	"golang.org/x/xerrors"
)

func main() {
	var nlong int
	var text string
	var fontPath string
	var outputFile string
	flag.IntVar(&nlong, "n", 1, "how long cat")
	flag.StringVar(&text, "t", "", "what is this chishiki")
	flag.StringVar(&fontPath, "f", "", "font path")
	flag.StringVar(&outputFile, "o", "", "output file name (only png)")
	flag.Parse()

	if fontPath == "" && len([]byte(text)) != len([]rune(text)) {
		fmt.Println("\x1b[33m[WARNING] If you use multibyte string, you should font path\x1b[0m")
		fmt.Printf(`Example
	chishiki -n %d -t '%s' -f [YOUR_FONT_PATH]

`, nlong, text)
	}

	if outputFile != "" && !strings.HasSuffix(outputFile, ".png") {
		fmt.Printf("Cannot output to '%s'\nYou should set '*.png' file\n", outputFile)
		return
	}

	top, err := getImage("/data.png")
	if err != nil {
		fmt.Printf("cannot load /data.png: %v", err)
		return
	}
	bottom, err := getImage("/data2.png")
	if err != nil {
		fmt.Printf("cannot load /data.png: %v", err)
		return
	}
	middle, err := getImage("/data3.png")
	if err != nil {
		fmt.Printf("cannot load /data.png: %v", err)
		return
	}
	width := top.Bounds().Dx()

	textImg, err := getTextImage(text, width, fontPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	height := top.Bounds().Dy() + bottom.Bounds().Dy()*nlong + middle.Bounds().Dy() + textImg.Bounds().Dy()
	rect := image.Rect(0, 0, width, height)
	canvas := image.NewRGBA(rect)
	h := 0
	draw.Draw(canvas, image.Rect(0, h, width, h+top.Bounds().Dy()), top, image.ZP, draw.Over)
	h += top.Bounds().Dy()
	for i := 0; i < nlong; i++ {
		draw.Draw(canvas, image.Rect(0, h, width, h+middle.Bounds().Dy()), middle, image.ZP, draw.Over)
		h += middle.Bounds().Dy()
	}
	draw.Draw(canvas, image.Rect(0, h, width, h+bottom.Bounds().Dy()), bottom, image.ZP, draw.Over)
	h += bottom.Bounds().Dy()
	draw.Draw(canvas, image.Rect(0, h, width, h+textImg.Bounds().Dy()), textImg, image.ZP, draw.Over)

	if outputFile == "" {
		enc := sixel.NewEncoder(os.Stdout)
		enc.Encode(canvas)
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("Cannot create %s", outputFile)
			return
		}
		defer f.Close()

		png.Encode(f, canvas)
	}
}

func getImage(path string) (image.Image, error) {
	f, err := Assets.Open(path)
	if err != nil {
		return nil, xerrors.Errorf("cannot open top file: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, xerrors.Errorf("cannot decode image: %w", err)
	}
	return img, nil
}

func getTextImage(text string, imageWidth int, fontPath string) (*image.RGBA, error) {
	var ft *truetype.Font
	var err error
	if fontPath == "" {
		ft, err = truetype.Parse(gobold.TTF)
	} else {
		ftBinary, err := ioutil.ReadFile(fontPath)
		if err != nil {
			return nil, xerrors.Errorf("cannot read file(%s): %w", fontPath, err)
		}
		ft, err = truetype.Parse(ftBinary)
	}

	if err != nil {
		return nil, xerrors.Errorf("cannnot parse to truetype: %w", err)
	}

	opt := truetype.Options{
		Size:              90,
		DPI:               0,
		Hinting:           0,
		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}

	imageHeight := 200
	textTopMargin := 90
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	face := truetype.NewFace(ft, &opt)

	dr := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{},
	}

	dr.Dot.X = (fixed.I(imageWidth) - dr.MeasureString(text)) / 2
	dr.Dot.Y = fixed.I(textTopMargin)

	dr.DrawString(text)
	return img, nil
}
