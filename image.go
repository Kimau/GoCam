package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"log"
	"os"
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

func saveGIFToFolder(name string, img image.Image, numCol int) {
	f, e := os.Create(name)
	palImg, found := img.(*image.Paletted)

	if e != nil {
		log.Println("Failed to Write", name, e)
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
		log.Println("Failed to Write", name, e)
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
	m := image.NewGray(src.Bounds())
	draw.Draw(m, m.Bounds(), src, image.ZP, draw.Src)

	return m
}
