/*  JOB: Fetch Images and Save to File

 */
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"net/http"
	"sync"
	"time"

	mjpeg "./mjpeg"
)

type computeData struct {
	stamp         time.Time
	lum           uint8
	frameDuration time.Duration
}

type camObject struct {
	name        string
	folder      string
	addr        string
	imgCur      int
	filesToLoop int
	lastImg     image.Image
	imgBuffer   []image.Image
	data        []computeData
	lock        sync.Mutex
}

func startCamCapture(filename string, address string) *camObject {
	co := camObject{
		name:        filename,
		folder:      CAPTURE_FOLDER,
		addr:        address,
		imgCur:      0,
		filesToLoop: MAX_IMAGE_PER_CAM,
		imgBuffer:   make([]image.Image, MAX_IMAGE_PER_CAM, MAX_IMAGE_PER_CAM),
		data:        []computeData{},
	}

	camImageChan := make(chan image.Image)
	go fetchMPEGCamLoop(&co, camImageChan)
	go saveLoopToFile(&co, camImageChan)
	return &co
}

func fetchMPEGCamLoop(co *camObject, outImg chan image.Image) {
	var decodeErr error
	var img image.Image

	for {
		resp, errA := http.Get(co.addr)

		if errA != nil {
			log.Println(co.addr, errA)
			return
		}

		log.Println("Fetching... ", co.addr, decodeErr)

		d, err := mjpeg.NewDecoderFromResponse(resp)
		if err != nil {
			log.Println("Failed to create Decoder:", co.addr, err)
			return
		}

		for decodeErr = d.Decode(&img); decodeErr == nil; decodeErr = d.Decode(&img) {
			outImg <- img
		}
	}
}

func saveLoopToFile(co *camObject, inImg <-chan image.Image) {
	i := 0
	var dataResult *computeData = nil
	var prevResult *computeData = nil
	addImg := true
	prevHour := time.Now().Hour()

	for {
		img := <-inImg
		addImg = true
		dataResult = &computeData{
			stamp: time.Now(),
		}

		// Do Compute Image
		computeImg := ToComputeImage(img)
		dataResult.lum = lumTotal(computeImg)

		if prevResult == nil {
			dataResult.frameDuration = time.Millisecond * 500
		} else {
			dataResult.frameDuration = dataResult.stamp.Sub(prevResult.stamp)

			var diff uint8
			if dataResult.lum > prevResult.lum {
				diff = dataResult.lum - prevResult.lum
			} else {
				diff = prevResult.lum - dataResult.lum
			}
			if diff < 3 {
				addImg = false
			}

		}

		if addImg {
			prevResult = dataResult

			co.lock.Lock()
			co.data = append(co.data, *dataResult)

			co.imgCur = i
			co.lastImg = img
			co.imgBuffer[i] = co.lastImg
			co.lock.Unlock()

			// Advance Image
			i = (i + 1) % co.filesToLoop
		}

		// Do Hourly Reports
		newHour := time.Now().Hour()
		if prevHour != newHour {
			prevHour = newHour
			co.lock.Lock()
			co.data = []computeData{*prevResult}
			lumImg := makeLumHourlyImg(co)
			saveGIFToFolder(fmt.Sprintf("_report_%s_%d.gif", co.name, newHour), lumImg, len(lumImg.Palette))
			co.lock.Unlock()
		}

	}
}

//------------------------------------------------------------------------------
// Merge Camera Feeds
func mergeCamFeeds(camObjs []*camObject) image.Image {
	imgList := []image.Image{}
	for _, v := range camObjs {
		v.lock.Lock()
		if v.lastImg == nil {
			log.Println("Empty Image")
			return nil
		}
		imgList = append(imgList, v.lastImg)
		v.lock.Unlock()
	}

	finalImg := mergeImage(imgList)
	return finalImg
}

//------------------------------------------------------------------------------
// Lum
func makeLumHourlyImg(co *camObject) *image.Paletted {

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

	// Setup Data
	endTime := co.data[len(co.data)-1].stamp
	t := endTime.Add(-time.Hour)

	offset := 0
	pOff := t.Second()
timeloop:
	for t.Before(endTime) {

		// Check offset isn't too far into the future
		tNext := t.Add(time.Second)

		// Run offset until current time
		for co.data[offset].stamp.Before(t) {
			if (offset + 1) >= len(co.data) {
				break timeloop
			}

			offset += 1
		}

		// Set Pixel
		m.Pix[pOff] = co.data[offset].lum

		//endTimeLoop:
		pOff += 1
		if pOff >= len(m.Pix) {
			pOff = 0
		}
		t = tNext
	}

	// End Point
	m.Pix[pOff] = 0
	if pOff > 0 {
		m.Pix[pOff-1] = uint8(numColours - 1)
	} else {
		m.Pix[len(m.Pix)-1] = uint8(numColours - 1)
	}

	return m
}

func makeLumTimeline(camObjs []*camObject) *image.Paletted {
	camHeight := 256
	width := 100
	stepHeight := camHeight / tempColourRange
	height := stepHeight * len(camObjs)

	m := image.NewPaletted(image.Rect(0, 0, width, height), tempColour)

	for camNum, cam := range camObjs {
		i := len(cam.data) - width
		if i < 0 {
			i = 0
		}
		x := 0

		for ; i < len(cam.data); i += 1 {
			y := int(cam.data[i].lum)%tempColourRange + stepHeight*camNum
			hOff := cam.data[i].lum / uint8(255/tempColourRange)
			if y == 0 {
				y = stepHeight*camNum - 1
				if hOff > 0 {
					hOff -= 1
				}
			}

			for ; y >= (stepHeight * camNum); y -= 1 {
				m.SetColorIndex(x, y, hOff+1)
			}
			x += 1
		}
	}

	return m
}

//
func makeCamGIF(co *camObject) *gif.GIF {

	co.lock.Lock()
	numImg := len(co.imgBuffer)

	g := gif.GIF{
		Image:     make([]*image.Paletted, numImg, numImg),
		Delay:     make([]int, numImg, numImg),
		LoopCount: -1,
		Disposal:  make([]byte, numImg, numImg),
	}

	hasNill := false

	for i, img := range co.imgBuffer {

		g.Disposal[i] = gif.DisposalBackground
		g.Delay[i] = 50

		if img == nil {
			hasNill = true
			g.Image[i] = nil
			continue
		}

		b := img.Bounds()

		pm, ok := img.(*image.Paletted)
		if !ok {
			pm = image.NewPaletted(b, palette.Plan9[:255])
			// pm.Palette = draw.FloydSteinberg.Quantize(make(color.Palette, 0, 256), img)
			draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
		}
		g.Image[i] = pm

	}
	co.lock.Unlock()

	if hasNill {
		i := 0
		for _, v := range g.Image {
			if v != nil {
				g.Image[i] = v
				i += 1
			}
		}
		g.Image = g.Image[:i]
		g.Delay = g.Delay[:i]
		g.Disposal = g.Disposal[:i]
	}

	return &g

}
