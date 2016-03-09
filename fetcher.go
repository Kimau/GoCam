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
	lastFile = make(chan string, 5)
	diffValChan := make(chan int, 5)
	everyFrame := make(chan image.Image, 5)
	everyBlock := make(chan computeBlock, 5)
	filterBlock := make(chan computeBlock, 5)

	go fetchMPEGCamLoop(addr, everyFrame, shutdown)
	go makeComputeBlock(everyFrame, everyBlock)
	go checkNewImage(everyBlock, filterBlock, diffValChan)
	go saveLoopToFile(filterBlock, name, lastFile)
	go saveMotionReport(name, diffValChan)

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

func checkNewImage(inBlock chan computeBlock, outBlock chan computeBlock, dValChan chan int) {
	var prevBlock computeBlock
	prevBlock.computeImg = nil

	for {
		newBlk, ok := <-inBlock
		if !ok {
			close(outBlock)
			close(dValChan)
			return
		}

		if prevBlock.computeImg == nil {
			// First Image
			prevBlock = newBlk
			outBlock <- prevBlock
		} else {
			// Compare Difference
			d := DiffImg(prevBlock.computeImg, newBlk.computeImg)

			diffVal := totalValFilter(d, 15)
			dValChan <- diffVal

			if diffVal > 1000 {
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
			saveMovie(filename)
			close(outfilename)
			return
		}

		// Save To File
		newFilename := fmt.Sprintf("%s/_%s_%d.jpg", CAPTURE_FOLDER, filename, newBlk.stamp.UnixNano())

		rgbImg := ToRGBAImage(newBlk.srcImg)
		DrawClock(rgbImg, &newBlk.stamp)
		saveJPEGToFolder(newFilename, rgbImg)

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
			// Composite Image
			mo := make([]*image.Gray, len(historyBlocks))
			for i, v := range historyBlocks {
				mo[i] = v.computeImg
			}

			cimg, cerr := MakeComposite(mo)
			if cerr == nil && cimg != nil {
				saveGIFToFolder(fmt.Sprintf("%s_comp_%d.gif", filename, historyBlocks[0].stamp.Hour()), cimg, 256)
			}

			// Start Movie Saving
			go saveMovie(filename)

			// Clear Out
			historyBlocks = []computeBlock{}
		}
	}
}

func saveMotionReport(filename string, dValChan chan int) {
	historyVal := []int{}
	dMax := 0
	lastReportHour := time.Now().Hour()

	for {
		newVal, ok := <-dValChan
		if !ok {
			lumImg := makeMotionReport(historyVal, dMax)
			saveGIFToFolder(fmt.Sprintf("_motion%s_%d.gif", filename, lastReportHour), lumImg, 2)
			return
		}

		historyVal = append(historyVal, newVal)
		if newVal > dMax {
			dMax = newVal
		}

		// Do Hourly Reports
		if lastReportHour != time.Now().Hour() {
			lumImg := makeMotionReport(historyVal, dMax)
			saveGIFToFolder(fmt.Sprintf("_motion%s_%d.gif", filename, lastReportHour), lumImg, 2)
			lastReportHour = time.Now().Hour()
			historyVal = []int{}
			dMax = 0
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
