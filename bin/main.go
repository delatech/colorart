package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/delatech/colorart"
	"github.com/disintegration/gift"
)

type Palette struct {
	BackgroundColor string `json:"background"`
	PrimaryColor    string `json:"primary"`
	SecondaryColor  string `json:"secondary"`
	DetailColor     string `json:"detail"`
}

var resizeThreshold int = 350
var resizeSize int = 320
var blurSigma float64 = 40.0
var useBlur bool = true

func init() {
	flag.IntVar(&resizeThreshold, "resize-threshold", resizeThreshold, "Resize threshold")
	flag.IntVar(&resizeSize, "resize-size", resizeSize, "Resize size")
	flag.Float64Var(&colorart.ContrastRatio, "contrast", colorart.ContrastRatio, "Mininum contrast to have between 2 colors (use 2 for accessibility compliance)")
	flag.Float64Var(&blurSigma, "blur-sigma", blurSigma, "Blur Sigma")
	flag.BoolVar(&useBlur, "blur", useBlur, "Blug picture (like iTunes12 do)")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("%s img\n", os.Args[0])
	}

	palette := analyzeFile(args[0])
	b, err := json.Marshal(palette)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintln(os.Stdout, string(b))

	if err != nil {
		log.Fatal(err)
	}
}

func analyzeFile(filename string) *Palette {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	b := img.Bounds()
	var g *gift.GIFT

	if b.Max.X-b.Min.X > resizeThreshold || b.Max.Y-b.Min.Y > resizeThreshold {
		g = gift.New(
			gift.Resize(resizeSize, 0, gift.LanczosResampling),
			gift.GaussianBlur(float32(blurSigma)))
	} else {
		g = gift.New(gift.GaussianBlur(float32(blurSigma)))
	}

	dst := image.NewRGBA(image.Rect(0, 0, resizeSize, resizeSize))
	g.Draw(dst, img)
	img = dst

	bg, c1, c2, c3 := colorart.Analyze(img)

	return &Palette{
		BackgroundColor: bg.String(),
		PrimaryColor:    c1.String(),
		SecondaryColor:  c2.String(),
		DetailColor:     c3.String(),
	}
}
