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
