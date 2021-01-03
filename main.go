package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s JSON [JSON...]", os.Args[0])
	}
	for _, json := range os.Args[1:] {
		r := newRegion(json)
		nimg := r.countImageFiles()
		if nimg > 1 {
			return
		}
		lines := r.gatherGTLines()
		if nimg != 0 && (len(lines) == nimg || len(lines) <= 1) {
			return
		}
		segimg(r.Dir, r.Image, lines)
	}
}

func segimg(dir, name string, lines []string) {
	log.Printf("segmenting %s into %d lines", name, len(lines))
	chk(os.MkdirAll(dir, 0777))
	n := len(lines)
	img := yClip(openImage(name))
	bnds := img.Bounds()
	var cs []int
	for y := 0; y < bnds.Max.Y; y++ {
		cs = append(cs, xPixCount(img, y))
	}
	from := bnds.Min
	for i := 0; i < n; i++ {
		for ; from.Y < bnds.Max.Y; from.Y++ {
			if cs[from.Y] != 0 {
				break
			}
		}
		to := image.Point{X: bnds.Max.X, Y: bnds.Max.Y}
		if i != n-1 {
			// Find minimum black pixel line.
			h := bnds.Max.Y - from.Y
			lh := h / (n - i)
			s, e := from.Y+(2*lh/3), from.Y+(2*lh)
			css := cs[s:e]
			min := argmin(css)
			to = image.Point{X: bnds.Max.X, Y: s + min}
		}
		rect := image.Rectangle{Min: from, Max: to}
		if rect.Empty() {
			log.Printf("skipping snippet line %d: image is empty", i+1)
			continue
		}
		// Write image snippet and gt line file.
		snippet := img.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(rect)
		outName := filepath.Join(dir, fmt.Sprintf("%06d.bin.png", i+1))
		writePNG(outName, xClip(snippet))
		gtName := filepath.Join(dir, fmt.Sprintf("%06d.gt.txt", i+1))
		chk(ioutil.WriteFile(gtName, []byte(lines[i]+"\n"), 0666))
		from.Y = to.Y
	}
}

func openImage(name string) image.Image {
	in, err := os.Open(name)
	chk(err)
	defer in.Close()
	img, _, err := image.Decode(in)
	chk(err)
	return img
}

func writePNG(name string, img image.Image) {
	out, err := os.Create(name)
	chk(err)
	defer func() { chk(out.Close()) }()
	chk(png.Encode(out, img))
}

func yClip(img image.Image) image.Image {
	bnds := img.Bounds()
	var b int
	for b = 0; b < bnds.Max.Y; b++ {
		if xPixCount(img, b) != 0 {
			break
		}
	}
	var e int
	for e = bnds.Max.Y; e > b; e-- {
		if xPixCount(img, e-1) != 0 {
			break
		}
	}
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(0, b, bnds.Max.X, e))
}

func xClip(img image.Image) image.Image {
	bnds := img.Bounds()
	var b int
	for b = bnds.Min.X; b < bnds.Max.X; b++ {
		if yPixCount(img, b) != 0 {
			break
		}
	}
	var e int
	for e = bnds.Max.X; e > b; e-- {
		if yPixCount(img, e-1) != 0 {
			break
		}
	}
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(b, bnds.Min.Y, e, bnds.Max.Y))
}

func xPixCount(img image.Image, y int) int {
	black := img.ColorModel().Convert(color.Black)
	bnds := img.Bounds()
	var c int
	for x := bnds.Min.X; x < bnds.Max.X; x++ {
		if img.At(x, y) == black {
			c++
		}
	}
	return c
}

func yPixCount(img image.Image, x int) int {
	black := img.ColorModel().Convert(color.Black)
	bnds := img.Bounds()
	var c int
	for y := bnds.Min.Y; y < bnds.Max.Y; y++ {
		if img.At(x, y) == black {
			c++
		}
	}
	return c
}

// len(args)>0
func argmin(cs []int) int {
	min := cs[0]
	var argmin int
	for i := 1; i < len(cs); i++ {
		if cs[i] < min {
			min = cs[i]
			argmin = i
		}
		if cs[i] > min*10 {
			break
		}
	}
	return argmin
}

// Just extract the "interesting" parts from the region file.
type region struct {
	Text  string
	Dir   string
	Image string
}

func newRegion(name string) region {
	in, err := os.Open(name)
	chk(err)
	defer in.Close()
	var region region
	chk(json.NewDecoder(in).Decode(&region))
	return region
}

func (r region) countImageFiles() int {
	var n int
	err := filepath.Walk(r.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".png") {
			n++
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return n
}

func (r region) gatherGTLines() []string {
	var ret []string
	rd := strings.NewReader(r.Text)
	s := bufio.NewScanner(rd)
	for s.Scan() {
		ret = append(ret, s.Text())
	}
	chk(s.Err())
	return ret
}

func chk(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
