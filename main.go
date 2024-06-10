package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	ddp "github.com/milindmadhukar/ddp-go"
)

func main() {
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

	fps := 45
	brightness := 0.1
	delay := 1000 / fps

	ticker := time.NewTicker(time.Millisecond * time.Duration(delay))

	// frames, err := GetMp4Frames("oops.mp4")
	frames, err := GetGIFFrames("./gifs/spiderverse.gif")
	// leftFrames, rightFrames := SplitFrames(frames)
	frameCount := len(frames)

	if err != nil {
		log.Fatal(err)
	}

	// Save all frames as pngs into gifparts folder

	// os.RemoveAll("./gifparts")
	// os.Mkdir("./gifparts", 0755)
	//
	// for i, frame := range frames {
	// 	out, err := os.Create(fmt.Sprintf("./gifparts/frame%d.png", i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	png.Encode(out, frame)
	// 	out.Close()
	// }

	idx := 0
	for range ticker.C {
		go func() {

			frame := frames[idx]


			plusData := ImageToPixelData(&plusMap, frame, plusPixelCount, brightness)
			plusClient.Write(plusData)

			crossData := ImageToPixelData(&crossMap, frame, crossPixelCount, brightness)
			crossClient.Write(crossData)
		}()

		idx = (idx + 1) % frameCount
	}

}
