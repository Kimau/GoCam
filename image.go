package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"log"
	"os"
	"sort"

	"github.com/disintegration/gift"
)

var (
	tempColourRange = 13
	tempColour      = []color.Color{
		color.RGBA{0x00, 0x00, 0x00, 0x00},
		color.RGBA{0xFE, 0x26, 0x12, 0xAA},
		color.RGBA{0xFE, 0x55, 0x08, 0xAA},
		color.RGBA{0xFC, 0x9A, 0x03, 0xAA},
		color.RGBA{0xFA, 0xBD, 0x02, 0xAA},
		color.RGBA{0xFE, 0xFE, 0x33, 0xAA},
		color.RGBA{0xD0, 0xEB, 0x2C, 0xAA},
		color.RGBA{0x66, 0xB1, 0x32, 0xAA},
		color.RGBA{0x03, 0x93, 0xCF, 0xAA},
		color.RGBA{0x02, 0x48, 0xFF, 0xAA},
		color.RGBA{0x3E, 0x01, 0xA4, 0xAA},
		color.RGBA{0x00, 0x00, 0x00, 0xFF},
		color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
	}
)

//------------------------------------------------------------------------------
// Image Functions
func mergeImage(imgs []image.Image) image.Image {
	width := 0
	height := 0

	imgs[0] = RotateImageLeft(imgs[0])

	for _, v := range imgs {
		if v.Bounds().Size().X > width {
			width = v.Bounds().Size().X
		}
		height += v.Bounds().Size().Y
	}

	m := image.NewRGBA(image.Rect(0, 0, width, height))

	height = 0
	for _, v := range imgs {
		h := v.Bounds().Size().Y
		draw.Draw(m,
			image.Rect(0, height, v.Bounds().Size().X, h+height),
			v,
			v.Bounds().Min,
			draw.Over)

		height += h
	}

	return m
}

func RotateImageLeft(src image.Image) image.Image {
	width := src.Bounds().Size().X
	height := src.Bounds().Size().Y

	m := draw.Image(image.NewRGBA(image.Rect(0, 0, height, width)))

	for x := 0; x < width; x += 1 {
		for y := 0; y < width; y += 1 {
			m.Set(y, width-x, src.At(x, y))
		}
	}

	return m
}

func saveJPEGToFolder(name string, img image.Image) {
	f, e := os.Create(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
	} else {
		jpeg.Encode(f, img, &jpeg.Options{80})
		f.Close()
	}
}

func loadJPEGFromFolder(name string) image.Image {
	f, e := os.Open(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
		return nil
	} else {
		img, _ := jpeg.Decode(f)
		f.Close()
		return img
	}
}

func saveAllGIFToFolder(name string, imgGif *gif.GIF) {
	f, e := os.Create(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
	} else {
		gif.EncodeAll(f, imgGif)
		f.Close()
	}
}

func saveGIFToFolder(name string, palImg *image.Paletted) {
	f, e := os.Create(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
	} else {
		gif.Encode(f, palImg, &gif.Options{NumColors: len(palImg.Palette)})
		f.Close()
	}
}

func loadGIFromFolder(name string) image.Image {
	f, e := os.Open(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
		return nil
	} else {
		img, _ := gif.Decode(f)
		f.Close()
		return img
	}
}

// Calculate Pallette
type colCount struct {
	c color.RGBA
	n int
}
type colListArr []colCount

func (cl colListArr) Len() int           { return len(cl) }
func (cl colListArr) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl colListArr) Less(i, j int) bool { return cl[i].n < cl[j].n }

func getDist(a uint8, b uint8) uint8 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func reduceColours(colList []colCount, numCol int) color.Palette {
	distThreshold := 4
	countThreshold := 1

	for len(colList) > numCol {
		// Merge Any Colours within minDist
		for i := 0; i < len(colList); i += 1 {
			colA := colList[i].c

			minDist := []int{0, 0}

			for j := i + 1; j < len(colList); j += 1 {
				colB := colList[j].c
				dist := []int{
					j,
					int(getDist(colB.R, colA.R)) + int(getDist(colB.G, colA.G)) + int(getDist(colB.B, colA.B)) + int(getDist(colB.A, colA.A)),
				}

				if (minDist[0] == 0) || (minDist[1] > dist[1]) {
					minDist = dist
				}
			}

			// Close Enough
			if minDist[1] < distThreshold {
				a := &colList[i]
				b := &colList[minDist[0]]

				inv := 1.0 / float64(a.n+b.n)
				a.c.R = uint8(float64((int(a.c.R)*a.n + int(b.c.R)*b.n)) * inv)
				a.c.G = uint8(float64((int(a.c.G)*a.n + int(b.c.G)*b.n)) * inv)
				a.c.B = uint8(float64((int(a.c.B)*a.n + int(b.c.B)*b.n)) * inv)
				a.c.A = uint8(float64((int(a.c.A)*a.n + int(b.c.A)*b.n)) * inv)

				a.n += b.n
				b.n = 0
			}

		}

		// Remove Anything with too few instances
		newColList := []colCount{}
		for _, v := range colList {
			if v.n > countThreshold {
				newColList = append(newColList, v)
			}
		}

		colList = newColList

		// Sort by Occurance
		sort.Sort(colListArr(colList))

		if len(colList) > (numCol * 4 / 5) {

			finalPal := []color.Color{}
			for i := 0; i < numCol; i += 1 {
				finalPal = append(finalPal, colList[i].c)
			}

			return color.Palette(finalPal)
		}

		m := numCol - 1
		if len(colList) < m {
			m = len(colList) - 1
		}

		//
		distThreshold *= 2
		countThreshold = (colList[m].n+countThreshold)/2 + 1
	}

	// Export
	sort.Sort(colListArr(colList))

	finalPal := []color.Color{}
	for i := 0; i < len(colList); i += 1 {
		finalPal = append(finalPal, colList[i].c)
	}

	return finalPal
}

func getColours(img image.Image) []colCount {
	// Gather Colours
	colMap := make(map[color.RGBA]int)
	minPt := img.Bounds().Min
	maxPt := img.Bounds().Max
	for x := minPt.X; x < maxPt.X; x += 1 {
		for y := minPt.Y; y < maxPt.Y; y += 1 {
			cr, cg, cb, ca := img.At(x, y).RGBA()
			cRGBA := color.RGBA{R: uint8(cr), G: uint8(cg), B: uint8(cb), A: uint8(ca)}
			colMap[cRGBA] += 1
		}
	}

	// Convert to Pair Colours
	colList := []colCount{}
	for col, num := range colMap {
		colList = append(colList, colCount{c: col, n: num})
	}

	return colList
}

//------------------------------------------------------------------------------
// Outline Image
func outlineImg(img draw.Image, col color.Color) image.Image {
	minPt := img.Bounds().Min
	maxPt := img.Bounds().Max

	for x := minPt.X; x < maxPt.X; x += 1 {
		img.Set(x, minPt.Y, col)
		img.Set(x, maxPt.Y-1, col)
	}

	for y := minPt.X; y < maxPt.Y; y += 1 {
		img.Set(minPt.X, y, col)
		img.Set(maxPt.X-1, y, col)
	}

	return img
}

//------------------------------------------------------------------------------
// Compute Stuff
func lumTotal(src *image.Gray) (lum uint8) {
	lum = 128

	var lumTotal int
	for _, v := range src.Pix {
		lumTotal += int(v)
	}

	lum = uint8(lumTotal / len(src.Pix))

	return lum
}

func ToComputeImage(src image.Image) *image.Gray {
	width := src.Bounds().Size().X
	height := src.Bounds().Size().Y

	m := image.NewGray(image.Rect(0, 0, width, height))

	for x := 0; x < width; x += 1 {
		for y := 0; y < width; y += 1 {
			m.Set(x, y, src.At(x, y))
		}
	}

	return m
}

//------------------------------------------------------------------------------
// Paletted
func makePaletted(img image.Image, numCol int) *image.Paletted {
	colList := getColours(img)
	newPal := reduceColours(colList, numCol)

	newPalImg := image.NewPaletted(img.Bounds(), newPal)
	draw.Draw(newPalImg, img.Bounds(), img, image.ZP, draw.Over)

	for i := 0; i < len(newPal); i += 1 {
		for y := i; y < len(newPalImg.Pix); y += newPalImg.Stride {
			newPalImg.Pix[y] = uint8(i)
		}
	}

	return newPalImg
}
