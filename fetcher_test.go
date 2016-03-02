package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestHourReport(t *testing.T) {

	timeStart := time.Now().Add(-time.Hour * 2)
	lastFrame := 450 + time.Duration(rand.Intn(100))
	var timCount uint8 = 0

	blkListRand := []computeBlock{}
	blkListRamp := []computeBlock{}

	for t := timeStart; time.Now().After(t); t = t.Add(time.Duration(time.Millisecond * lastFrame)) {
		lastFrame = 450 + time.Duration(rand.Intn(100))

		blkListRand = append(blkListRand, computeBlock{
			stamp: t,
			lum:   uint8(rand.Intn(0xFF)),
		})

		blkListRamp = append(blkListRamp, computeBlock{
			stamp: t,
			lum:   timCount,
		})

		timCount += 1
	}

	saveGIFToFolder("_testReport0.gif", makeLumHourlyImg(blkListRand), 256)
	saveGIFToFolder("_testReport1.gif", makeLumHourlyImg(blkListRamp), 256)

	saveGIFToFolder("_testTimeline0.gif", makeLumTimeline(blkListRand), 256)
	saveGIFToFolder("_testTimeline1.gif", makeLumTimeline(blkListRamp), 256)
}
