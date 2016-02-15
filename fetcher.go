/*  JOB: Fetch Images and Save to File

 */
package main

import (
	"image"
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

	for {
		img := <-inImg
		co.lock.Lock()
		if dataResult != nil {
			co.data = append(co.data, *dataResult)
		}
		co.imgCur = i
		co.lastImg = img
		co.imgBuffer[i] = co.lastImg
		computeImg := ToComputeImage(img)
		co.lock.Unlock()

		// Do Compute Image
		dataResult = &computeData{
			stamp: time.Now(),
		}

		dataResult.lum = lumTotal(computeImg)
		if len(co.data) == 0 {
			dataResult.frameDuration = 0
		} else {
			prev := &co.data[len(co.data)-1]
			dataResult.frameDuration = dataResult.stamp.Sub(prev.stamp)
		}

		i = (i + 1) % co.filesToLoop
	}
}

//------------------------------------------------------------------------------
// Merge Camera Feeds
func mergeCamFeeds(camObjs []*camObject) image.Image {
	imgList := []image.Image{}
	for _, v := range camObjs {
		v.lock.Lock()
		imgList = append(imgList, v.lastImg)
		v.lock.Unlock()
	}

	finalImg := mergeImage(imgList)
	return finalImg
}

//------------------------------------------------------------------------------
// Lum
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
/*
func makePaletted(img image.Image) *image.Paletted {


		for _, v := range camObjs {
						v.lock.Lock()
						gifData := gif.GIF{
							LoopCount: -1,
						}

						numImg := len(v.imgBuffer)

						palImages := make([]*image.Paletted, numImg, numImg)
						gifData.Disposal = make([]byte, numImg, numImg)
						gifData.Delay = make([]int, numImg, numImg)

						for i, img := range v.imgBuffer {
							newPal := getColours(img)
							newPalImg := image.NewPaletted(img.Bounds(), newPal)
							draw.Draw(newPalImg, img.Bounds(), img, image.ZP, draw.Over)

							saveGIFToFolder(fmt.Sprintf("_%s_%d.gif", v.name, i), newPalImg)

							palImages = append(palImages, newPalImg)
							gifData.Disposal[i] = gif.DisposalBackground
							gifData.Delay[i] = 500
						}
						v.lock.Unlock()

						gifData.Image = palImages

						filename := fmt.Sprintf("_%s.gif", v.name)
						saveAllGIFToFolder(filename, &gifData)

}
*/
