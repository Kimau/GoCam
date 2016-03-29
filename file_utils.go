package main

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"time"
)

func extractNameDate(filename string) (name string, tStamp time.Time, err error) {
	re := regexp.MustCompile("_([A-Za-z]*)_([0-9]*).*")
	substr := re.FindAllStringSubmatch(filename, -1)

	if len(substr) > 0 {
		camDate, _ := strconv.ParseInt(substr[0][2], 10, 64)
		tStampDate := time.Unix(0, camDate)
		return substr[0][1], tStampDate, nil
	}

	return "", time.Unix(0, 0), errors.New("Unable to extract name")
}

func setupCaptureFolder() error {
	// Create Capture Folder
	FileErr := os.RemoveAll(CAPTURE_FOLDER)
	if FileErr != nil && !os.IsNotExist(FileErr) {
		return FileErr
	}

	os.Mkdir(CAPTURE_FOLDER, os.ModePerm)
	return nil
}
