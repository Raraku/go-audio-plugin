package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

const samepleRate = 44100
const seconds = 1

func main() {
	flag.Parse()
	log.SetFlags(0)
	portaudio.Initialize()
	defer portaudio.Terminate()
	// numDevices,err := portaudio.Devices()
	// chk(err,7)
	// for _, dev := range numDevices{
	// 	fmt.Println(dev.Name)
	// 	portaudio.
	// }

	buffer2 := &bytes.Buffer{}
	buffer := make([]float32, samepleRate*seconds)
	stream, _ := portaudio.OpenDefaultStream(1, 0, samepleRate, len(buffer), func(in []float32) {
		for i := range buffer {
			buffer[i] = in[i]
		}
	})
	// chk(err,0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	log.Printf(c.LocalAddr().String())
	chk(err, 1)
	defer c.Close()
	done := make(chan struct{})

	defer close(done)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	chk(err, 2)
	chk(stream.Start(), 3)

	defer stream.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:

			fmt.Println("doing this again")
			// err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			// if err != nil{
			// 	log.Println("write:", err)
			// 	return
			// }]s
			stream.Read()
			err = binary.Write(buffer2, binary.BigEndian, &buffer)
			chk(err, 4)
			// fmt.Println(buffer2.Bytes())
			sEnc := base64.StdEncoding.EncodeToString([]byte(buffer2.Bytes()))
			// fmt.Println(sEnc)
			message := fmt.Sprintf(`{"event":"media", "media":{"payload":"%s"}}`, sEnc)
			err = c.WriteMessage(websocket.TextMessage, []byte(message))
			buffer2.Reset()
			chk(err, 5)

		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			chk(err, 6)
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func chk(err error, n int64) {
	if err != nil {
		log.Println(n)
		panic(err)
	}
}
