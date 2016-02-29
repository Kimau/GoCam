package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type CommandFunc func(string) error

const (
	MAX_IMAGE_PER_CAM = 10
	CAPTURE_FOLDER    = "./Capture"
)

var (
	addr         = flag.String("addr", "127.0.0.1:1667", "Web Address")
	db           = flag.String("db", "_data.db", "Database")
	staticFldr   = flag.String("static", "./static", "Static Folder")
	templateFldr = flag.String("template", "./templates", "Templates Folder")
	debug        = flag.Bool("debug", false, "enter debug mode")
	telegram     = flag.Bool("telegram", false, "telegram bot live")
	commandFuncs = make(map[string]CommandFunc)
)

func init() {
	commandFuncs["help"] = listCommands

	switch runtime.GOOS {
	case "windows":
		commandFuncs["clear"] = func(string) error {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
			return nil
		}

	case "linux":
		fallthrough
	default:
		commandFuncs["clear"] = func(string) error {
			print("\033[H\033[2J")
			return nil
		}
	}

}

func spltFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {

	var BYTE_SPLIT_TOKEN = []byte("--boundarydonotcross")
	var TokLen = len(BYTE_SPLIT_TOKEN)

	advance = bytes.Index(data, BYTE_SPLIT_TOKEN)
	if advance < 0 {
		return len(data), nil, nil
	}

	return advance + TokLen, data[:advance], nil
}

func main() {
	commandFuncs["clear"]("clear")
	flag.Parse()
	if *debug {
		log.Println("Debug Active")
	}

	//// Start Web Server
	//log.Println("Start Web Server")
	//wf := web.MakeWebFace(*addr, *staticFldr, *templateFldr)
	//
	//// Setup Webface with Database
	//SetupWebFace(wf)
	//wf.RedirectHandler = nil

	// Create Capture Folder
	FileErr := os.RemoveAll(CAPTURE_FOLDER)
	if FileErr != nil && !os.IsNotExist(FileErr) {
		log.Fatalln(FileErr)
		return
	}

	os.Mkdir(CAPTURE_FOLDER, os.ModePerm)

	camList := []*camObject{
		startCamCapture("camA", "http://admin:admin@192.168.1.99/goform/video"),
		startCamCapture("camB", "http://admin:admin@192.168.1.100/goform/video"),
	}

	// go startUploader(wf)
	commandFuncs["cam"] = func(string) error {
		saveJPEGToFolder("_mergecam.jpg", mergeCamFeeds(camList))
		return nil
	}

	commandFuncs["lum"] = func(string) error {
		saveGIFToFolder("_lum.gif", outlineImg(makeLumTimeline(camList), color.Black).(*image.Paletted), 255)
		return nil
	}

	commandFuncs["pal"] = func(line string) error {
		var numCol int
		var err error

		tok := strings.Split(line, " ")
		if len(tok) > 1 {
			numCol, err = strconv.Atoi(tok[1])
			if err != nil {
				numCol = 255
			}
		}

		saveGIFToFolder("_pal.gif", camList[0].lastImg, numCol)
		return nil
	}

	// Running Loop
	if *telegram {

		log.Println("Running Loop")
		go startBot(camList)
	}
	commandLoop()

	// Clean up
	log.Println("Clean up")
}

func commandLoop() {
	lines := scanForInput()

	for {
		fmt.Println("Enter Command: ")
		select {
		case line := <-lines:
			cmd := strings.ToLower(strings.Split(line, " ")[0])

			valFunc, ok := commandFuncs[cmd]
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
	for i := range commandFuncs {
		commandOut += i + ", "
	}

	fmt.Println(commandOut)
	return nil
}
