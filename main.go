package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	movieCmd     = flag.String("movieCmd", "./ffmpeg.exe -r 6 -f concat -i %s -c:v libx264 -pix_fmt yuv420p mov%s_%d_%d.mp4", "Set Movie Cmd")
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

func saveMovie(camName string) {
	prefix := fmt.Sprintf("_%s", camName)

	// Get File List
	rawfiles, _ := ioutil.ReadDir(CAPTURE_FOLDER)
	files := []string{}
	for _, f := range rawfiles {
		fn := f.Name()
		if strings.HasPrefix(fn, prefix) {
			fn, _ = filepath.Abs(CAPTURE_FOLDER + "\\" + fn)
			files = append(files, fn)
		}
	}

	if len(files) < 3 {
		fmt.Println("Not enough files", camName, files)
		return
	}

	// Write to Temp File
	tFilename := fmt.Sprintf("_templist_%s.txt", camName)
	tempFile, _ := os.Create(tFilename)
	for _, f := range files {
		fmt.Fprintf(tempFile, "file '%s'\n", f)
	}
	tempFile.Close()

	fullCmd := fmt.Sprintf(*movieCmd, tFilename, camName, time.Now().Day(), time.Now().Hour())
	fmt.Println(fullCmd)

	cmdList := strings.Split(fullCmd, " ")

	cmd := exec.Command(cmdList[0], cmdList[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Waiting for movie to finish...")
		cmd.Wait()
	}

	// Remove Files
	os.Remove(tFilename)
	for _, f := range files {
		os.Remove(f)
	}

	fmt.Println("Done")
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

	camAShutdown, camALastFile := captureFilterCameraPipe("http://admin:admin@192.168.1.99/goform/video", "lounge")
	camBShutdown, camBLastFile := captureFilterCameraPipe("http://admin:admin@192.168.1.100/goform/video", "kitchen")

	commandFuncs["lum"] = func(string) error {
		return nil
	}

	commandFuncs["movie"] = func(string) error {
		fmt.Println("Making Movie")
		go saveMovie("lounge")
		go saveMovie("kitchen")
		return nil
	}

	// Running Loop
	if *telegram {

		log.Println("Running Loop")
		b := startBot()
		b.AddCamera("camA", camALastFile)
		b.AddCamera("camB", camBLastFile)
	}
	commandLoop()

	// Clean up
	log.Println("Clean up")
	camAShutdown <- 1
	camBShutdown <- 1

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
