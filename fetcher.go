/*  JOB: Fetch Images and Save to File

 */
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"time"

	mjpeg "./mjpeg"
)

func fetchMPEGCamLoop(addr string, outImg chan image.Image) {
	var decodeErr error
	var img image.Image

	for {
		resp, errA := http.Get(addr)

		if errA != nil {
			log.Println(addr, errA)
			return
		}

		log.Println("Fetching... ", addr, decodeErr)

		d, err := mjpeg.NewDecoderFromResponse(resp)
		if err != nil {
			log.Println("Failed to create Decoder:", addr, err)
			return
		}

		for decodeErr = d.Decode(&img); decodeErr == nil; decodeErr = d.Decode(&img) {
			outImg <- img
		}
	}
}

type computeBlock struct {
	stamp      time.Time
	lum        uint8
	computeImg *image.Gray
	srcImg     image.Image
}

func startComputePipe(srcImg chan image.Image, outBlock chan computeBlock) {
	computeMaker := MakeComputeMaker(<-srcImg)
	for {
		cb := computeBlock{
			stamp:  time.Now(),
			srcImg: <-srcImg,
		}

		// Make Compute
		cb.computeImg = computeMaker.Convert(cb.srcImg)
		cb.lum = lumTotal(cb.computeImg)

		outBlock <- cb
	}
}

func checkNewImage(inBlock chan computeBlock, outBlock chan computeBlock) {
	var prevBlock computeBlock
	prevBlock.computeImg = nil
	for {
		newBlk := <-inBlock

		// First Image, Diff Lum or
		if (prevBlock.computeImg == nil) || (prevBlock.lum != newBlk.lum) {
			prevBlock = newBlk
			outBlock <- prevBlock
			continue
		}
	}
}

func saveLoopToFile(inBlock chan computeBlock, filename string) {
	historyBlocks := []computeBlock{}
	for {
		newBlk := <-inBlock

		newBlk.computeImg = nil
		newBlk.srcImg = nil

		historyBlocks = append(historyBlocks, newBlk)

		// Save To File
		saveJPEGToFolder(fmt.Sprintf("%s_%s.gif", filename, newBlk.stamp.String()), newBlk.srcImg)

		// Do Hourly Reports
		if newBlk.stamp.Hour() != historyBlocks[0].stamp.Hour() {
			lumImg := makeLumHourlyImg(historyBlocks)
			saveGIFToFolder(fmt.Sprintf("%s_%d.gif", filename, historyBlocks[0].stamp.Hour()), lumImg, len(lumImg.Palette))
			historyBlocks = []computeBlock{}
		}
	}
}

//------------------------------------------------------------------------------
// Lum
func makeLumHourlyImg(blkList []computeBlock) *image.Paletted {
	// Make Colour Pal
	numColours := 256
	w := 60
	h := 60

	cg := ColourGrad([]GradStop{
		{0.0, color.RGBA{0, 0, 255, 255}},
		{0.5, color.RGBA{0, 255, 0, 255}},
		{1.0, color.RGBA{255, 0, 0, 255}},
	})
	tempPal := cg.makePal(0.0, 1.0, numColours)
	m := image.NewPaletted(image.Rect(0, 0, w, h), tempPal)

	// Start
	for _, blk := range blkList {
		pOff := blk.stamp.Second() + blk.stamp.Minute()*60
		m.Pix[pOff] = blk.lum
	}

	return m
}

func makeLumTimeline(blkList []computeBlock) *image.Paletted {
	// Make Colour Pal
	numColours := 256
	width := len(blkList)
	height := numColours

	cg := ColourGrad([]GradStop{
		{0.0, color.RGBA{0, 0, 255, 255}},
		{0.5, color.RGBA{0, 255, 0, 255}},
		{1.0, color.RGBA{255, 0, 0, 255}},
	})
	tempPal := cg.makePal(0.0, 1.0, numColours)
	m := image.NewPaletted(image.Rect(0, 0, width, height), tempPal)

	for i, blk := range blkList {
		hOff := blk.lum

		for off, y := i, 0; y <= int(hOff); off, y = i+y*numColours, y+1 {
			m.Pix[i] = hOff
		}

	}

	return m
}
