package main

import (
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"os"

	//"github.com/llgcode/draw2d"
	//"github.com/llgcode/draw2d/draw2dimg"
)

func mergeImage(imgs []image.Image) image.Image {
	width := 0
	height := 0

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
