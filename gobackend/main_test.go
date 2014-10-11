package main

import (
	"testing"

	"code.google.com/p/go.net/websocket"
)

func TestGame(t *testing.T) {
	service := "ws://localhost:8080/play/"

	conn, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	msg := "join"
	err = websocket.Message.Send(conn, msg)
	checkError(err)

	conn2, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	msg = "join"
	err = websocket.Message.Send(conn2, msg)
	checkError(err)

	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)
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
