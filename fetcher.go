/*  JOB: Fetch Images and Save to File

 */
package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"sync"

	mjpeg "./mjpeg"
)

type camObject struct {
	name        string
	folder      string
	addr        string
	filesToLoop int
	lock        sync.Mutex
}

func startCamCapture(filename string, address string) *camObject {
	co := camObject{
		name:        filename,
		folder:      CAPTURE_FOLDER,
		addr:        address,
		filesToLoop: MAX_IMAGE_PER_CAM,
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

	for {
		img := <-inImg
		co.lock.Lock()
		saveJPEGToFolder(fmt.Sprintf("%s/%s.jpeg", co.folder, co.name), img)
		co.lock.Unlock()

		saveJPEGToFolder(fmt.Sprintf("%s/%s_%d.jpeg", co.folder, co.name, i), img)

		i = (i + 1) % co.filesToLoop
	}
}

func saveJPEGToFolder(name string, img image.Image) {
	f, e := os.Create(name)
	if e != nil {
		log.Println("Failed to Write", name, e)
	} else {
		jpeg.Encode(f, img, &jpeg.Options{80})
		f.Close()
	}
}
