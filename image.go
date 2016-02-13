package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
)

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

/// Compute Stuff

func lumTotal(src *image.Gray) (lum uint8) {
	lum = 128

	var lumTotal int
	for _, v := range src.Pix {
		lumTotal += int(v)
	}

	lum = uint8(lumTotal / len(src.Pix))

	return lum
}
