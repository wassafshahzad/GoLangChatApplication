package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type ServerRoom struct {
	connectionMap map[*websocket.Conn]bool
	sync.RWMutex
}

func (room *ServerRoom) addConnection(ws *websocket.Conn) {
	room.Lock()
	room.connectionMap[ws] = true
	room.Unlock()
}

func (room *ServerRoom) deleteConnection(ws *websocket.Conn) {
	room.Lock()
	delete(room.connectionMap, ws)
	room.Unlock()
}

func (room *ServerRoom) connect(ws *websocket.Conn) {
	fmt.Printf("Incoming connection from %v \n", ws.RemoteAddr())

	room.addConnection(ws)
	room.mainLoop(ws)
}

func (room *ServerRoom) mainLoop(ws *websocket.Conn) {
	for {
		buff := make([]byte, 1024)
		fmt.Println("Waiting for message")
		n, err := ws.Read(buff)

		if err != nil {
			if err == io.EOF {
				ws.Close()
				room.deleteConnection(ws)
				return
			}
			fmt.Println("Error:", err)
			continue
		}
		room.broadcast(buff[:n], ws)
	}
}

func (s *ServerRoom) broadcast(message []byte, ws *websocket.Conn) {
	s.RLock()
	for connection := range s.connectionMap {
		if connection == ws {
			continue
		} else {
			_, err := connection.Write(message)
			if err != nil {
				fmt.Println("Error in writing")
			}
		}
	}
	s.RUnlock()
}

func printHelloWorld(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func main() {

	server := ServerRoom{
		connectionMap: make(map[*websocket.Conn]bool),
	}
	http.HandleFunc("/", printHelloWorld)
	http.Handle("/ws", websocket.Handler(server.connect))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
