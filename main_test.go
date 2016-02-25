package main

import (
	"image"
	"image/color"
	"testing"
	"time"
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

func TestHourReport(t *testing.T) {
	co := camObject{
		name:        "_test",
		folder:      CAPTURE_FOLDER,
		imgCur:      0,
		filesToLoop: MAX_IMAGE_PER_CAM,
		imgBuffer:   make([]image.Image, MAX_IMAGE_PER_CAM, MAX_IMAGE_PER_CAM),
		data:        []computeData{},
	}

	camImageChan := make(chan image.Image)
	timeStart := time.Now().Add(-time.Hour * 2)

	for t := timeStart; t < time.Now(); t = t.Add(time.Second * 0.5) {
		data = append(data, computeData{stamp: t, lum: 0, frameDuration: 500})
	}

	m := makeLumTimeline([]*camObject{&co})
	saveGIFToFolder("_testReport.gif", m, 256)
}
