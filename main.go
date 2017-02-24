package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	mjpeg "GoCamCapture/mjpeg"
)

type comandFunc func(string) error

var (
	comandFuncs = make(map[string]comandFunc)

	camMap = map[string]string{
		"lounge":  "http://admin:admin@192.168.1.99/goform/video",
		"kitchen": "http://admin:admin@192.168.1.100/goform/video",
	}
)

func init() {
	comandFuncs["help"] = listCommands

	switch runtime.GOOS {
	case "windows":
		comandFuncs["clear"] = func(string) error {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
			return nil
		}

	case "linux":
		fallthrough
	default:
		comandFuncs["clear"] = func(string) error {
			print("\033[H\033[2J")
			return nil
		}
	}
}

func fetchMPEGCamLoop(addr string, outImg chan image.Image) {
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
			default:
				outImg <- img
			}
		}

	}
}

func main() {
	comandFuncs["clear"]("clear")
	flag.Parse()

	// Running Loop
	log.Println("Running Loop")
	b := startBot()

	for k, v := range camMap {
		outImgChan := make(chan image.Image)
		go fetchMPEGCamLoop(v, outImgChan)
		b.AddCamera(k, outImgChan)
	}

	lines := scanForInput()

	for {
		fmt.Println("Enter Command: ")
		select {
		case line := <-lines:
			cmd := strings.ToLower(strings.Split(line, " ")[0])

			valFunc, ok := comandFuncs[cmd]
			if ok {
				err := valFunc(line)

				if err != nil {
					log.Printf("Error [%s]: %s", line, err.Error())
				}
			} else if line == "quit" {
				return
			} else {
				log.Printf("Unknown command: %s", line)
				listCommands("")
			}

		}
	}
}

func scanForInput() chan string {
	lines := make(chan string)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	return lines
}

func listCommands(string) error {
	commandOut := "Commands: "
	for i := range comandFuncs {
		commandOut += i + ", "
	}

	fmt.Println(commandOut)
	return nil
}
