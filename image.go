package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"

	"errors"
	"log"
	"os"

	"github.com/bamiaux/rez"
	//"github.com/disintegration/gift"
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
			draw.Src)

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
		log.Println("saveJPEGToFolder Failed to Create", name, e)
	} else {
		jpeg.Encode(f, img, &jpeg.Options{80})
		f.Close()
	}
}

func loadJPEGFromFolder(name string) image.Image {
	f, e := os.Open(name)
	if e != nil {
		log.Println("loadJPEGFromFolder Failed to Open", name, e)
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
		log.Println("saveAllGIFToFolder Failed to Create", name, e)
	} else {
		gif.EncodeAll(f, imgGif)
		f.Close()
	}
}

func saveGIFToFolder(name string, img image.Image, numCol int) {
	f, e := os.Create(name)
	palImg, found := img.(*image.Paletted)

	if e != nil {
		log.Println("saveGIFToFolder Failed to Create", name, e)
		return
	}

	if found {
		gif.Encode(f, palImg, &gif.Options{NumColors: len(palImg.Palette)})
	} else {
		gif.Encode(f, img, &gif.Options{NumColors: numCol})
	}
	f.Close()
}

func loadGIFromFolder(name string) image.Image {
	f, e := os.Open(name)
	if e != nil {
		log.Println("loadGIFromFolder Failed to Open", name, e)
		return nil
	} else {
		img, _ := gif.Decode(f)
		f.Close()
		return img
	}
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
func lumTotal(src *image.Gray) (lum int) {
	lum = 0
	for _, v := range src.Pix {
		lum += int(v)
	}

	return lum
}

func lumAvg(src *image.Gray) (lum uint8) {
	lum = uint8(lumTotal(src) / len(src.Pix))
	return lum
}

func totalValFilter(src *image.Gray, lowFilter uint8) (lum int) {
	lum = 0
	for _, v := range src.Pix {
		if v > lowFilter {
			vi := int(v)
			lum += vi * vi
		}
	}

	return lum
}

func LowFilter(src *image.Gray, lowFilter uint8) {

	for o, omax := 0, len(src.Pix); o < omax; o += 1 {
		if src.Pix[o] <= lowFilter {
			src.Pix[o] = 0
		}
	}

}

func DiffImg(a, b *image.Gray) *image.Gray {
	if a.Bounds() != b.Bounds() {
		return nil
	}

	m := image.NewGray(a.Bounds())
	for o, omax := 0, len(m.Pix); o < omax; o += 1 {
		aPix := a.Pix[o]
		bPix := b.Pix[o]
		if aPix > bPix {
			m.Pix[o] = aPix - bPix
		} else {
			m.Pix[o] = bPix - aPix
		}
	}

	return m
}

func ToComputeImageManual(src image.Image) *image.Gray {
	fullBound := src.Bounds()
	m := image.NewGray(fullBound)
	draw.Draw(m, fullBound, src, image.ZP, draw.Src)

	// Smaller Image
	smallBound := image.Rect(0, 0, fullBound.Dx()/4, fullBound.Dy()/4)
	smallImg := image.NewGray(smallBound)

	swid := smallBound.Dx()
	wid := fullBound.Dx()
	for o, o2, omax := 0, 0, len(smallImg.Pix); o < omax; o, o2 = o+1, o2+4 {
		if (o % swid) == 0 {
			o2 = (o / swid) * wid * 4
		}

		smallImg.Pix[o] = uint8((int(m.Pix[o2+0]) +
			int(m.Pix[o2+3]) +
			int(m.Pix[o2+2+wid]) +
			int(m.Pix[o2+1+wid*2]) +
			int(m.Pix[o2+0+wid*3]) +
			int(m.Pix[o2+3+wid*3])) / 6)
	}

	return smallImg
}

func ToComputeImageNearest(src image.Image) *image.Gray {
	fullBound := src.Bounds()
	m := image.NewGray(fullBound)
	draw.Draw(m, fullBound, src, image.ZP, draw.Src)

	// Smaller Image
	smallBound := image.Rect(0, 0, fullBound.Dx()/4, fullBound.Dy()/4)
	smallImg := image.NewGray(smallBound)

	swid := smallBound.Dx()
	wid := fullBound.Dx()
	for o, o2, omax := 0, 0, len(smallImg.Pix); o < omax; o, o2 = o+1, o2+4 {
		if (o % swid) == 0 {
			o2 = (o / swid) * wid * 4
		}

		smallImg.Pix[o] = m.Pix[o2+0]
	}

	return smallImg
}

func ToRGBAImage(src image.Image) *image.RGBA {
	conv, ok := src.(*image.RGBA)
	if ok {
		return conv
	}

	fullBound := src.Bounds()
	m := image.NewRGBA(fullBound)
	draw.Draw(m, fullBound, src, image.ZP, draw.Src)
	return m
}

func ToComputeImage(src image.Image) *image.Gray {
	// Small M
	bSize := src.Bounds().Size()
	smallBounds := image.Rect(0, 0, bSize.X/16, bSize.Y/16)
	smallM := image.NewGray(smallBounds)

	rez.Convert(smallM, src, rez.NewBilinearFilter())
	return smallM
}

func ToComputeImageCol(src image.Image) *image.RGBA {
	// Small M
	//bSize := src.Bounds().Size()
	smallBounds := src.Bounds() // image.Rect(0, 0, bSize.X/16, bSize.Y/16)
	smallM := image.NewRGBA(smallBounds)

	rez.Convert(smallM, src, rez.NewBilinearFilter())

	return smallM
}

type ComMaker struct {
	bound         image.Rectangle
	width, height int
	rConv         rez.Converter
}

func MakeComputeMaker(src image.Image) *ComMaker {
	b := src.Bounds()
	sz := b.Size()
	width := sz.X / 8
	height := sz.Y / 8
	smallB := image.Rect(0, 0, width, height)

	cm := ComMaker{
		width:  sz.X / 8,
		height: sz.Y / 8,
		bound:  smallB,
	}

	// Make Grey Img
	gSrc := image.NewGray(b)
	draw.Draw(gSrc, b, src, image.ZP, draw.Src)

	// Make Smaller
	gDst := image.NewGray(smallB)

	// Prep Conversion
	cfg, err := rez.PrepareConversion(gDst, gSrc)
	if err != nil {
		log.Println(err)
		return nil
	}

	cm.rConv, err = rez.NewConverter(cfg, rez.NewBilinearFilter())
	if err != nil {
		log.Println(err)
		return nil
	}

	return &cm
}

func (cm *ComMaker) Convert(src image.Image) *image.Gray {
	// Make Grey Img
	b := src.Bounds()
	gSrc := image.NewGray(b)
	draw.Draw(gSrc, b, src, image.ZP, draw.Src)

	// Make Smaller
	gDst := image.NewGray(cm.bound)
	cm.rConv.Convert(gDst, gSrc)

	return gDst
}

//--------------------------------------------------------------

func MakeComposite(srcImages []*image.Gray) (*image.Gray, error) {
	// Possible low pass filter

	return nil, errors.New("Not Fixed")

	isSet := false
	var maxB image.Rectangle
	/* Assume first image is bound set
	for i := len(srcImages) - 1; i > 0; i -= 1 {
		maxB := maxB.Union(srcImages[i].Bounds())
	} */

	var res []int

	for _, img := range srcImages {
		if img == nil {
			continue
		}
		if !isSet {
			res = make([]int, len(img.Pix))
			maxB = img.Bounds()
		}
		if img.Bounds() != maxB {
			return nil, errors.New("Invalid Bounds Match")
		}

		for i := len(res) - 1; i >= 0; i -= 1 {
			res[i] += int(img.Pix[i])
		}
	}

	// Find Max
	maxVal := 0
	for _, v := range res {
		if v > maxVal {
			maxVal = v
		}
	}

	// Level Image
	resImg := image.NewGray(maxB)
	for i, v := range res {
		resImg.Pix[i] = uint8((v * 255) / maxVal)
	}

	return resImg, nil
}
