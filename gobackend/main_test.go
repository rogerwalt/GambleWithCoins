package main

import (
	"testing"
	"encoding/json"
	"code.google.com/p/go.net/websocket"
)

func TestGame(t *testing.T) {
	service := "ws://localhost:8080/play/"

	// open connection as player 1
	conn, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	// send "join" command as player 1
	b := []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn, b)
	checkError(err)

	// open connection as player 2
	conn2, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	// send "join" command as player 2
	b = []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn2, b)
	checkError(err)

	// receive "matched" message from server as player 2
	var msg string
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)

	// interpret message as JSON data
	var f map[string]interface{}
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	if f["command"] != "matched" {
		t.Error("Expected matched, got ", msg)
	}

	// receive "matched" message from server as player 1
	err = websocket.Message.Receive(conn, &msg)
	checkError(err)

	// interpret message as JSON data
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	if f["command"] != "matched" {
		t.Error("Expected matched, got ", msg)
	}

	// send "cooperate" command as player 1
	b = []byte(`{"command": "cooperate"}`)
	err = websocket.Message.Send(conn, b)
	checkError(err)

	// send "defect" command as player 2
	b = []byte(`{"command": "defect"}`)
	err = websocket.Message.Send(conn2, b)
	checkError(err)

	// receive game result as player 1
	err = websocket.Message.Receive(conn, &msg)
	checkError(err)
	if msg != "0" {
		t.Error("Expected 0, got ", msg)
	}

	// receive game result as player 2
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)
	if msg != "3" {
		t.Error("Expected 3, got ", msg)
	}
}
