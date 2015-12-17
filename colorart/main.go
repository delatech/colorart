package main

//

import (
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/disintegration/gift"
	"github.com/sspencer/colorart"
)

const (
	ResizeThreshold = 250
	ResizeSide      = 20
)

type Cover struct {
	Filename, BackgroundColor, PrimaryColor, SecondaryColor, DetailColor string
}

func analyzeFile(filename string) (bg, c1, c2, c3 colorart.Color) {
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
	if b.Max.X-b.Min.X > ResizeThreshold && b.Max.Y-b.Min.Y > ResizeThreshold {
		g := gift.New(gift.Resize(ResizeSide, 0, gift.LanczosResampling))
		dst := image.NewRGBA(image.Rect(0, 0, ResizeSide, ResizeSide))
		g.Draw(dst, img)
		img = dst
	}

	return colorart.Analyze(img)
}

func main() {

	if len(os.Args) < 3 {
		log.Fatalf("%s <template> <img 1> <img 2> ... <img n>\n", os.Args[0])
	}

	tpl, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("webpage").Parse(string(tpl))

	covers := make([]Cover, 0, len(os.Args)-2)

	for i := 2; i < len(os.Args); i++ {
		bg, c1, c2, c3 := analyzeFile(os.Args[i])
		covers = append(covers, Cover{os.Args[i], bg.String(), c1.String(), c2.String(), c3.String()})
	}

	err = t.Execute(os.Stdout, covers)
	if err != nil {
		log.Fatal(err)
	}
}
