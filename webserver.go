package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024*4,
	WriteBufferSize: 1024*4,
}

var lastRecievedFrame time.Time = time.Now()

func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if messageType == websocket.BinaryMessage {
			fmt.Println(string(p))
		}

		if messageType == websocket.TextMessage {

			// timeTaken := time.Since(lastRecievedFrame)
			lastRecievedFrame = time.Now()

			coI := strings.Index(string(p), ",")

			switch strings.TrimSuffix(string(p[5:coI]), ";base64") {
			case "image/png":
				reader := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(p[coI+1:]))
				img, _, err := image.Decode(reader)
				if err != nil {
					log.Println(err)
					return
				}
				// fmt.Println("Current frame rate: ", 1000/timeTaken.Milliseconds())
				framesChan <- img
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "screenshare.html")
}

func screenshare(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Connection established")

	go reader(conn)
}

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/screenshare", screenshare)
}
