package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
			// Save the video as a webm file on disk

			file, err := os.Create("video.webm")
			if err != nil {
				log.Printf("Error creating video file: %v", err)
			}

			if _, err := file.Write(message); err != nil {
				log.Printf("Error writing video data to file: %v", err)
			}

			if err := file.Close(); err != nil {
				log.Printf("Error closing video file: %v", err)
			}

			if err := processVideoData(message); err != nil {
				log.Printf("Error processing video data: %v", err)
			}
		}

		if messageType == websocket.TextMessage {
			fmt.Println("Received text message")
		}
	}
}

// Convert the bytes into frames of image.Image (image.NRGBA?)
func processVideoData(videoBytes []byte) error {

  fmt.Println("Processing video data")

	// Create the FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0", // Read from stdin
		"-vf", "fps=1", // Extract 1 frame per second
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-")

  fmt.Println("Created ffmpeg command")

	// Create a pipe for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %v", err)
	}

  fmt.Println("Created stdin pipe")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}

  fmt.Println("Created stdout pipe")

	// Start the FFmpeg process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting FFmpeg: %v", err)
	}

  fmt.Println("Started ffmpeg process")

	// Write the WebM data to stdin in a separate goroutine
	go func() {
		defer stdin.Close()
		_, err := stdin.Write(videoBytes)
		if err != nil {
			fmt.Printf("Error writing to stdin: %v\n", err)
		}
	}()

  fmt.Println("Wrote video data to stdin")

	// Process each JPEG frame from stdout
	buffer := bytes.NewBuffer(nil)
	chunk := make([]byte, 4096)

	for {
		n, err := stdout.Read(chunk)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading from stdout: %v", err)
		}
		if n == 0 {
			break
		}
		buffer.Write(chunk[:n])

		for {
			img, err := jpeg.Decode(buffer)
			if err != nil {
				if err == jpeg.FormatError("missing SOI marker") || err == io.ErrUnexpectedEOF || err == io.EOF {
					break // Partial image data, need to read more
				}
				return fmt.Errorf("error decoding JPEG: %v", err)
			}

			// Send the decoded image to the channel
			framesChan <- img

			// Remove the decoded image from the buffer
			imgSize := buffer.Len() - bytes.NewReader(buffer.Bytes()).Len()
			buffer = bytes.NewBuffer(buffer.Bytes()[imgSize:])
		}
	}

  fmt.Println("Processed video data")

	// Wait for the FFmpeg process to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for FFmpeg: %v", err)
	}

  fmt.Println("Waited for ffmpeg process")

	return nil
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
