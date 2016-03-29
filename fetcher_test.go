package main

import (
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"testing"
	"time"
)

const (
	TESTDATA_FOLDER = "./_TestFolder"
	TEST_FILE_LIMIT = 100
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

func TestFolder(t *testing.T) {
	// Read Files
	setupCaptureFolder()
	rawfiles, _ := ioutil.ReadDir(TESTDATA_FOLDER)
	t.Logf("TEST---- %d", len(rawfiles))

	type testData struct {
		t time.Time
		f string
	}

	camData := make(map[string][]testData)

	for i, f := range rawfiles {
		if testing.Short() && i > TEST_FILE_LIMIT {
			t.Logf("Short test stopping at %d file", i)
			break
		}

		fn, _ := filepath.Abs(TESTDATA_FOLDER + "\\" + f.Name())

		camName, camDate, camErr := extractNameDate(f.Name())
		if camErr != nil {
			t.Log(camErr)
		} else {
			v, ok := camData[camName]
			if ok {
				camData[camName] = append(v, testData{t: camDate, f: fn})
			} else {
				camData[camName] = []testData{{t: camDate, f: fn}}
			}
		}
	}

	tFunc := func(name string, tList []testData) {

		lastFile := make(chan string, 5)
		diffValChan := make(chan int, 5)
		everyBlock := make(chan computeBlock, 5)
		filterBlock := make(chan computeBlock, 5)

		go checkNewImage(everyBlock, filterBlock, diffValChan)
		go saveLoopToFile(filterBlock, name, lastFile)
		go saveMotionReport(name, diffValChan)

		for _, dat := range tList {

			file := loadJPEGFromFolder(dat.f)

			cb := computeBlock{
				stamp:      dat.t,
				srcImg:     file,
				computeImg: ToComputeImageManual(file),
			}

			cb.lum = lumAvg(cb.computeImg)

			everyBlock <- cb
		}

		close(everyBlock)
	}

	for k, v := range camData {
		tFunc(k, v)
	}

}
