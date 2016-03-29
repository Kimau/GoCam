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

type computeBlock struct {
	stamp      time.Time
	lum        uint8
	computeImg *image.Gray
	diffImg    *image.Gray
	srcImg     image.Image
	diffVal    int
}

const (
	DIFF_VAL_LOWPASS = 20
	DIFF_VAL_CUTOFF  = 30500
)

func captureFilterCameraPipe(addr string, name string) (shutdown chan int, lastFile chan string) {

	shutdown = make(chan int)
	lastFile = make(chan string, 5)
	everyFrame := make(chan image.Image, 5)
	everyBlock := make(chan computeBlock, 5)
	filterBlock := make(chan computeBlock, 5)

	go fetchMPEGCamLoop(addr, everyFrame, shutdown)
	go makeComputeBlock(everyFrame, everyBlock)
	go checkNewImage(everyBlock, filterBlock)
	go saveLoopToFile(filterBlock, name, lastFile)

	return shutdown, lastFile
}

func fetchMPEGCamLoop(addr string, outImg chan image.Image, shutdown chan int) {
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
			select {
			case <-shutdown:
				close(outImg)
				return
			default:
				outImg <- img
			}
		}

	}
}

func makeComputeBlock(srcImg chan image.Image, outBlock chan computeBlock) {
	for {
		img, ok := <-srcImg
		if !ok {
			close(outBlock)
			return
		}

		cb := computeBlock{
			stamp:      time.Now(),
			srcImg:     img,
			computeImg: ToComputeImageManual(img),
		}

		cb.lum = lumAvg(cb.computeImg)

		outBlock <- cb
	}
}

func checkNewImage(inBlock chan computeBlock, outBlock chan computeBlock) {
	var prevBlock computeBlock
	prevBlock.computeImg = nil

	for {
		newBlk, ok := <-inBlock
		if !ok {
			close(outBlock)
			return
		}

		if prevBlock.computeImg == nil {
			// First Image
			prevBlock = newBlk
			outBlock <- prevBlock
		} else {
			// Compare Difference
			newBlk.diffImg = DiffImg(prevBlock.computeImg, newBlk.computeImg)
			newBlk.diffVal = totalValFilter(newBlk.diffImg, DIFF_VAL_LOWPASS)

			if newBlk.diffVal > DIFF_VAL_CUTOFF {
				prevBlock = newBlk
				outBlock <- prevBlock
			}
		}
	}
}

func saveLoopToFile(inBlock chan computeBlock, cameraName string, outfilename chan string) {
	historyBlocks := []computeBlock{}

	for {
		newBlk, ok := <-inBlock
		if !ok {
			saveMovie(cameraName)
			close(outfilename)
			return
		}

		// Save To File
		newFilename := fmt.Sprintf("%s/_%s_%d.jpg", CAPTURE_FOLDER, cameraName, newBlk.stamp.UnixNano())

		if newBlk.diffImg != nil {
			rgbImg := ToRGBAImage(newBlk.srcImg)
			DrawClock(rgbImg, &newBlk.stamp)
			saveJPEGToFolder(newFilename, rgbImg)
		}

		// Non Blocking Channel
		select {
		case outfilename <- newFilename:
		default:
		}

		// Clear out mem
		newBlk.srcImg = nil
		historyBlocks = append(historyBlocks, newBlk)

		// Do Hourly Reports
		if newBlk.stamp.Hour() != historyBlocks[0].stamp.Hour() {

			// Start Movie Saving
			go saveMovie(cameraName)

			// Clear Out
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

		for off, y := i, 0; y <= int(hOff); off, y = i+y*width, y+1 {
			m.Pix[off] = hOff
		}

	}

	return m
}

func makeMotionReport(dVal []int, maxVal int) *image.Paletted {
	// Make Colour Pal
	width := len(dVal)
	height := maxVal / 1000

	tempPal := color.Palette([]color.Color{
		color.RGBA{0, 0, 0, 0},
		color.RGBA{255, 255, 255, 255},
	})

	m := image.NewPaletted(image.Rect(0, 0, width, height), tempPal)

	for i, v := range dVal {
		hOff := int(v / 1000)
		if hOff >= height {
			hOff = height - 1
			fmt.Println("Height too large, corrected")
		}

		for off, y := i, 0; y <= hOff; off, y = off+width, y+1 {
			m.Pix[off] = 1
		}

	}

	return m
}
