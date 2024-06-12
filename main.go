package main

import (
	"encoding/json"
	"image"
	"log"
	"net/http"
	"os"
	"time"

	ddp "github.com/milindmadhukar/ddp-go"
	"golang.org/x/image/draw"
)

var framesChan = make(chan image.Image, 100)

func main() {
	setupRoutes()
	go http.ListenAndServe(":8069", nil)

	plusClient, err := ddp.DefaultDDPConnection("192.168.1.41", 4048)
	if err != nil {
		log.Fatal(err)
	}

	defer plusClient.Close()

	crossClient, err := ddp.DefaultDDPConnection("192.168.1.42", 4048)
	if err != nil {
		log.Fatal(err)
	}

	defer crossClient.Close()

	var plusMap MatrixMap
	plusMapFile, err := os.Open("./plusledmap.json")
	if err != nil {
		log.Fatal(err)
	}

	plusMapDecoder := json.NewDecoder(plusMapFile)
	err = plusMapDecoder.Decode(&plusMap)
	if err != nil {
		log.Fatal(err)
	}

	plusMapFile.Close()

	var crossMap MatrixMap
	crossMapFile, err := os.Open("./crossledmap.json")
	if err != nil {
		log.Fatal(err)
	}

	crossMapDecoder := json.NewDecoder(crossMapFile)
	err = crossMapDecoder.Decode(&crossMap)
	if err != nil {
		log.Fatal(err)
	}

	crossMapFile.Close()

	plusPixelCount := GetPixelCount(&plusMap)
	crossPixelCount := GetPixelCount(&crossMap)

	fps := 40
	brightness := 0.1
	delay := 1000 / fps
	split := true

	ticker := time.NewTicker(time.Millisecond * time.Duration(delay))

	frames, err := GetGIFFrames("./gifs/spiderverse.gif")
  // frames, err := GetMp4Frames("./oops.mp4")
	frameCount := len(frames)

	if err != nil {
		log.Fatal(err)
	}

	// Generate frames from the GIF
	go func() {
		idx := 0
		for range ticker.C {
			// framesChan <- frames[idx]
			idx = (idx + 1) % frameCount
		}
	}()

	lastRecievedFrame := time.Now()

	for frame := range framesChan {
		go func(frame image.Image) {
			var leftFrame, rightFrame *image.RGBA
			if split {
				leftFrame = image.NewRGBA(image.Rect(0, 0, frame.Bounds().Dx()/2, frame.Bounds().Dy()))
				rightFrame = image.NewRGBA(image.Rect(0, 0, frame.Bounds().Dx()/2, frame.Bounds().Dy()))

				draw.Draw(leftFrame, leftFrame.Bounds(), frame, image.Point{0, 0}, draw.Over)
				draw.Draw(rightFrame, rightFrame.Bounds(), frame, image.Point{frame.Bounds().Dx() / 2, 0}, draw.Over)

			} else {
				leftFrame = frame.(*image.RGBA)
				rightFrame = frame.(*image.RGBA)
			}

			plusData := ImageToPixelData(&plusMap, leftFrame, plusPixelCount, brightness)
			plusClient.Write(plusData)

			crossData := ImageToPixelData(&crossMap, rightFrame, crossPixelCount, brightness)
			crossClient.Write(crossData)

			timeTaken := time.Since(lastRecievedFrame)
			lastRecievedFrame = time.Now()

			log.Println("Current frame rate: ", 1E9/timeTaken.Nanoseconds())
		}(frame)
	}

}
