package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 16,
	WriteBufferSize: 1024 * 16,
}

var lastRecievedFrame time.Time = time.Now()

func reader(conn *websocket.Conn) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if messageType == websocket.BinaryMessage {
      fmt.Println("Received binary message", message)
		}

		if messageType == websocket.TextMessage {
      img, err := jpeg.Decode(bytes.NewReader([]byte(message)))
      if err != nil {
        log.Println(err)
        return
      }

      framesChan <- img
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
