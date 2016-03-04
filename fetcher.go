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
	srcImg     image.Image
}

func captureFilterCameraPipe(addr string, name string) (shutdown chan int, lastFile chan string) {

	shutdown = make(chan int)
	lastFile = make(chan string)
	everyFrame := make(chan image.Image)
	everyBlock := make(chan computeBlock)
	filterBlock := make(chan computeBlock)

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
			d := DiffImg(prevBlock.computeImg, newBlk.computeImg)
			diff := lumTotal(d)

			if diff > 5 {
				prevBlock = newBlk
				outBlock <- prevBlock
			}
		}
	}
}

func saveLoopToFile(inBlock chan computeBlock, filename string, outfilename chan string) {
	historyBlocks := []computeBlock{}

	for {
		newBlk, ok := <-inBlock
		if !ok {
			close(outfilename)
			return
		}

		// Save To File
		newFilename := fmt.Sprintf("%s/_%s_%d.jpg", CAPTURE_FOLDER, filename, newBlk.stamp.UnixNano())
		saveJPEGToFolder(newFilename, newBlk.srcImg)

		// Non Blocking Channel
		select {
		case outfilename <- newFilename:
		default:
		}

		// Clear out mem
		newBlk.computeImg = nil
		newBlk.srcImg = nil
		historyBlocks = append(historyBlocks, newBlk)

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

		for off, y := i, 0; y <= int(hOff); off, y = i+y*width, y+1 {
			m.Pix[off] = hOff
		}

	}

	return m
}
