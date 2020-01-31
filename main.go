//go:generate go-assets-builder -s="/data" -o image.go data
//go:generate goimports -w image.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
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
	flag.IntVar(&nlong, "n", 1, "how long cat")
	flag.StringVar(&text, "t", "知識", "what is this chishiki")
	flag.StringVar(&fontPath, "f", "", "font path")
	flag.Parse()

	if fontPath == "" && len([]byte(text)) != len([]rune(text)) {
		fmt.Println("\x1b[33m[WARNING] If you use multibyte string, you should font path\x1b[0m")
		fmt.Printf(`Example
	chishiki -n %d -t '%s' -f [YOUR_LANGUAGE_PATH]

`, nlong, text)
	}

	top, middle, bottom, err := getImages()
	if err != nil {
		fmt.Println(err)
		return
	}
	width := top.Bounds().Dx()

	textImg, err := getTextImage(text, width, fontPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	height := top.Bounds().Dy() + bottom.Bounds().Dy()*nlong + middle.Bounds().Dy() + textImg.Bounds().Dy()
	dst := imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
	h := 0
	dst = imaging.Paste(dst, top, image.Pt(0, h))
	h += top.Bounds().Dy()
	for i := 0; i < nlong; i++ {
		dst = imaging.Paste(dst, middle, image.Pt(0, h))
		h += middle.Bounds().Dy()
	}
	dst = imaging.Paste(dst, bottom, image.Pt(0, h))
	h += bottom.Bounds().Dy()
	dst = imaging.Paste(dst, textImg, image.Pt(0, h))

	enc := sixel.NewEncoder(os.Stdout)
	enc.Encode(dst)
}

func getImages() (top, middle, bottom image.Image, err error) {
	var f http.File
	f, err = Assets.Open("/data.png")
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("cannot open top file: %w", err)
	}
	defer f.Close()
	top, _, _ = image.Decode(f)

	f, err = Assets.Open("/data2.png")
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("cannot open bottom file: %w", err)
	}
	defer f.Close()
	bottom, _, _ = image.Decode(f)

	f, err = Assets.Open("/data3.png")
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("cannot open middle file: %w", err)
	}
	defer f.Close()
	middle, _, _ = image.Decode(f)
	return
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

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		return nil, xerrors.Errorf("cannot encode img: %w", err)
	}
	return img, nil
}
