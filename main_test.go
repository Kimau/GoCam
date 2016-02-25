package main

import (
	"image"
	"image/color"
	"math/rand"
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

func TestLumTimeline(t *testing.T) {
	co := []*camObject{{
		name:        "_testReportRand",
		folder:      CAPTURE_FOLDER,
		imgCur:      0,
		filesToLoop: MAX_IMAGE_PER_CAM,
		data:        []computeData{},
	},
		{
			name:        "_testReportScale",
			folder:      CAPTURE_FOLDER,
			imgCur:      0,
			filesToLoop: MAX_IMAGE_PER_CAM,
			data:        []computeData{},
		},
	}

	timeStart := time.Now().Add(-time.Hour * 2)
	timeDFactor := 0.45 + rand.Float32()*0.15
	var timCount uint8 = 0

	for t := timeStart; time.Now().After(t); t = t.Add(time.Duration(float32(time.Second) * timeDFactor)) {
		co[0].data = append(co[0].data, computeData{
			stamp:         t,
			lum:           uint8(rand.Intn(0xFF)),
			frameDuration: 450 + time.Duration(rand.Intn(100)),
		})

		co[1].data = append(co[1].data, computeData{
			stamp:         t,
			lum:           timCount,
			frameDuration: 450 + time.Duration(rand.Intn(100)),
		})

		timCount += 1
		timeDFactor = 0.45 + rand.Float32()*0.15
	}

	m := makeLumTimeline(co)
	saveGIFToFolder("_testTimeline.gif", m, 256)
}

func TestHourReport(t *testing.T) {
	co := []*camObject{{
		name:        "_testReportRand",
		folder:      CAPTURE_FOLDER,
		imgCur:      0,
		filesToLoop: MAX_IMAGE_PER_CAM,
		data:        []computeData{},
	},
		{
			name:        "_testReportScale",
			folder:      CAPTURE_FOLDER,
			imgCur:      0,
			filesToLoop: MAX_IMAGE_PER_CAM,
			data:        []computeData{},
		},
	}

	timeStart := time.Now().Add(-time.Hour * 2)
	timeDFactor := 0.45 + rand.Float32()*0.15
	var timCount uint8 = 0

	for t := timeStart; time.Now().After(t); t = t.Add(time.Duration(float32(time.Second) * timeDFactor)) {
		co[0].data = append(co[0].data, computeData{
			stamp:         t,
			lum:           uint8(rand.Intn(0xFF)),
			frameDuration: 450 + time.Duration(rand.Intn(100)),
		})

		co[1].data = append(co[1].data, computeData{
			stamp:         t,
			lum:           timCount,
			frameDuration: 450 + time.Duration(rand.Intn(100)),
		})

		timCount += 1
		timeDFactor = 0.45 + rand.Float32()*0.15
	}

	saveGIFToFolder("_testReport0.gif", makeLumHourlyImg(co[0]), 256)
	saveGIFToFolder("_testReport1.gif", makeLumHourlyImg(co[1]), 256)
}
