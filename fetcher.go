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

	mjpeg "./mjpeg"
)

func fetchMPEGCamLoop(name string, addr string) {
	var decodeErr error
	var img image.Image
	var prevImg image.Image

	i := 0

	for {
		resp, errA := http.Get(addr)

		if errA != nil {
			log.Println(name, addr, errA)
			return
		}

		log.Println("Fetching... ", name, addr, decodeErr)

		d, err := mjpeg.NewDecoderFromResponse(resp)
		if err != nil {
			log.Println("Failed to create Decoder:", name, addr, err)
			return
		}

		for decodeErr = d.Decode(&img); decodeErr == nil; decodeErr = d.Decode(&img) {
			f, e := os.Create(fmt.Sprintf("%s/%s %d.jpeg", CAPTURE_FOLDER, name, i))
			i = (i + 1) % MAX_IMAGE_PER_CAM
			if e != nil {
				log.Println("Failed to Write", name, addr, e)
			} else {
				jpeg.Encode(f, img, &jpeg.Options{80})
				f.Close()
			}

			if prevImg != nil {

			}
			prevImg = img
		}
	}

}
