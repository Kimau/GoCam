package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	google "./google"
	mjpeg "./mjpeg"
	web "./web"
)

type CommandFunc func() error

const ()

var (
	addr         = flag.String("addr", "127.0.0.1:1667", "Web Address")
	db           = flag.String("db", "_data.db", "Database")
	staticFldr   = flag.String("static", "./static", "Static Folder")
	templateFldr = flag.String("template", "./templates", "Templates Folder")
	debug        = flag.Bool("debug", false, "show HTTP traffic")
	commandFuncs = make(map[string]CommandFunc)
)

func init() {
	commandFuncs["help"] = listCommands

	switch runtime.GOOS {
	case "windows":
		commandFuncs["clear"] = func() error {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
			return nil
		}

	case "linux":
		fallthrough
	default:
		commandFuncs["clear"] = func() error {
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
	commandFuncs["clear"]()
	flag.Parse()
	if *debug {
		log.Println("Debug Active")
	}

	// Start Web Server
	log.Println("Start Web Server")
	wf := web.MakeWebFace(*addr, *staticFldr, *templateFldr)
	wf.RedirectHandler = func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "Starting Server on %s", *addr)
	}

	// Login
	log.Println("Login")
	Tok, cErr := google.Login(wf, google.GetClientScope())
	if cErr != nil {
		log.Fatalln("Login Error:", cErr)
	}

	iTok, iErr := google.GetIdentity(Tok)
	if iErr != nil {
		log.Fatalln("Identity Error:", iErr)
	}
	fmt.Println(iTok)

	b := new(bytes.Buffer)
	google.EncodeToken(Tok, b)

	// Setup Webface with Database
	SetupWebFace(wf)
	wf.RedirectHandler = nil

	log.Println("First Capture")
	{
		respA, errA := http.Get("http://admin:admin@192.168.1.99/goform/video")

		if errA != nil {
			log.Println("http://192.168.1.99/", errA)
		} else {
			// Stuff

			var img image.Image
			d, err := mjpeg.NewDecoderFromResponse(respA)
			i := 0

			if err != nil {
				log.Println("Failed to Decode", err)
			} else {
				d.Decode(&img)
				fmt.Println(img.Bounds())
			}

			f, e := os.Create(fmt.Sprintf("CameraA %d.jpeg", i))
			i = (i + 1) % 50
			if e != nil {
				log.Println("Failed to Write")
			} else {
				jpeg.Encode(f, img, &jpeg.Options{80})
				f.Close()
			}

			return

			/*

				reader := bufio.NewScanner(respA.Body)
				reader.Split(spltFunc)
				i := 0

				for reader.Scan() {

					b := reader.Bytes()
					bl := len(b)

					if bl > 0 {
						f, e := os.Create(fmt.Sprintf("CameraA %d.jpeg", i))
						i = (i + 1) % 50
						if e != nil {
							log.Println("Failed to Write")
						} else {
							f.Write(b)
							f.Close()
						}
					}

					log.Println("Snap", bl)
				}

				respA.Body.Close()
			*/
		}

		log.Println("Upload File A...")
		google.InsertFile("Camera A", "", "", "image/jpeg", respA.Body)
	}
	/*
		respB, errB := http.Get("http://admin:admin@192.168.1.100/goform/video")

		if errB != nil {
			log.Println("http://192.168.1.100/", errB)
		} else {
			// Stuff
			log.Println("Upload File B...")
			f, e := os.Create("CameraB.webp")
			if e != nil {
				log.Println("Failed to Write")
			} else {
				io.Copy(f, respB.Body)
				respB.Body.Close()
				f.Close()
			}

			google.InsertFile("Camera B", "", "", "image/webp", respB.Body)
		}*/

	// Running Loop
	log.Println("Running Loop")
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
			line = strings.ToLower(line)

			valFunc, ok := commandFuncs[line]
			if ok {
				err := valFunc()

				if err != nil {
					log.Printf("Error [%s]: %s", line, err.Error())
				}
			} else if line == "quit" {
				return
			} else {
				log.Printf("Unknown command: %s", line)
				listCommands()
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

func listCommands() error {
	commandOut := "Commands: "
	for i, _ := range commandFuncs {
		commandOut += i + ", "
	}

	fmt.Println(commandOut)
	return nil
}
