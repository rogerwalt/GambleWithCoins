package main

import (
	"log"
	"testing"
	"encoding/json"
	"code.google.com/p/go.net/websocket"
)

func TestGame(t *testing.T) {
	service := "ws://localhost:8080/play/"

	// open first connection (player 1)
	conn, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	// send join command from player 1
	b := []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn, b)
	checkError(err)

	// open second connection (player 2)
	conn2, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	// send join command from player 2
	b = []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn2, b)
	checkError(err)

	// receive message from server
	var msg string
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)

	// interpret message as JSON data
	var f interface{}
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	log.Printf("%v", f)

	if msg != "matched" {
		t.Error("Expected matched, got ", msg)
	}

	err = websocket.Message.Receive(conn, &msg)
	checkError(err)
	if msg != "matched" {
		t.Error("Expected matched, got ", msg)
	}

	msg = "C"
	err = websocket.Message.Send(conn, msg)
	checkError(err)

	msg = "D"
	err = websocket.Message.Send(conn2, msg)
	checkError(err)

	err = websocket.Message.Receive(conn, &msg)
	checkError(err)
	if msg != "0" {
		t.Error("Expected 0, got ", msg)
	}

	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)
	if msg != "3" {
		t.Error("Expected 3, got ", msg)
	}
}
