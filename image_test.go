package main

import (
	"fmt"
	"image"
	"image/color"
	"testing"
)

func TestGrad(t *testing.T) {
	numColours := 256
	w := 440
	h := 220

	cg := ColourGrad([]GradStop{
		{0.0, color.RGBA{0, 0, 255, 255}},
		{0.5, color.RGBA{0, 255, 0, 255}},
		{1.0, color.RGBA{255, 0, 0, 255}},
	})
	tempPal := cg.makePal(0.0, 1.0, numColours)
	m := image.NewPaletted(image.Rect(0, 0, w, h), tempPal)

	// Setup Data
	for y := 0; y < h; y += 1 {
		var colIDX uint8 = uint8(numColours * y / h)
		for x := 0; x < w; x += 1 {
			m.SetColorIndex(x, y, colIDX)
		}
	}

	saveGIFToFolder("_testPal.gif", m, numColours)
}

func TestComputeImage(t *testing.T) {

	camImageChan := make(chan image.Image)
	shutdown := make(chan int)

	go fetchMPEGCamLoop("http://admin:admin@192.168.1.99/goform/video", camImageChan, shutdown)

	img := <-camImageChan
	saveJPEGToFolder("_testCamRaw.jpg", img)

	cm := MakeComputeMaker(img)
	if cm == nil {
		t.Log("Unable to make Compute Maker")
		t.FailNow()
	}

	fmt.Println("Converted Made")

	comp := cm.Convert(img)
	comp2 := ToComputeImageManual(img)
	comp3 := ToComputeImageNearest(img)
	comp4 := DiffImg(comp2, comp3)
	saveGIFToFolder("_testCamCompute.gif", comp, 256)
	saveGIFToFolder("_testCamCompute2.gif", comp2, 256)
	saveGIFToFolder("_testCamCompute3.gif", comp3, 256)
	saveGIFToFolder("_testCamCompute4.gif", comp4, 256)

	close(shutdown)

	fmt.Println("Done")

}
