package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"code.google.com/p/go.net/websocket"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"
)

// TODO: shutting down the server gracefully does not yet work
func startServer() chan int {
	time.Sleep(100 * time.Millisecond)
	// use fresh db
	os.Remove("./test.db")
	serverClose := make(chan int)
	go start("./test.db", 8080, serverClose)

	time.Sleep(100 * time.Millisecond)
	return serverClose
}

func TestGame(t *testing.T) {
	log.Println("Test: Starting server")
	serverClose := startServer()

	loginAndRegister(t)
	testDatabase(t)
	game(t)

	serverClose <- 1
}

func testDatabase(t *testing.T) {
	log.Println("Test: ----- testingDatabase() procedure -----")
	balance, err := masc.GetBalance("foo1")
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}
	log.Println("Test: Balance of user foo1:", balance)

	log.Println("Test: Depositing 100 to user foo1")
	err = masc.UpdateBalance("foo1", 100)
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}

	balance, err = masc.GetBalance("foo1")
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}
	if balance == 100 {
		log.Println("Test: Balance of user foo1:", balance)
	} else {
		log.Println("Test: ERROR: Expected balance of 100")
	}

	log.Println("Test: Depositing -50 to user foo1")
	err = masc.UpdateBalance("foo1", -50)
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}

	balance, err = masc.GetBalance("foo1")
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}
	if balance == 50 {
		log.Println("Test: Balance of user foo1:", balance)
	} else {
		log.Println("Test: ERROR: Expected balance of 50")
	}

	log.Println("Test: Registering two users with two different names")
	masc.Register("foo1", "bar1")
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}
	masc.Register("foo2", "bar2")
	if err != nil {
		log.Println("Test: ERROR: ", err.Error())
	}

	log.Println("Test: ----- testingDatabase() ended -----")
}

func loginAndRegister(t *testing.T) {
	log.Println("Test: ----- loginAndRegister() procedure----- ")
	service := "ws://localhost:8080/play/"

	// open connection as player 1
	log.Println("Test: Connecting as player 1")
	conn, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	log.Println("Test: Registering as player 1")
	// send "join" command as player 1
	b := []byte(`{"command": "register", "name" : "foo1", "password" : "bar1"}`)
	err = websocket.Message.Send(conn, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	var msg string
	err = websocket.Message.Receive(conn, &msg)
	checkError(err)

	var f map[string]interface{}
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)
	if f["command"] != "register" {
		log.Println("Test: ERROR: Expected register, got", msg)
	}
	if f["result"] != "success" {
		log.Println("Test: ERROR: Expected success, got ", msg)
	}

	// open connection as player 1
	log.Println("Test: Open second Connection")
	conn2, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	log.Println("Test: Login as player 1")
	// send "join" command as player 1
	b = []byte(`{"command": "login", "name" : "foo1", "password" : "bar1"}`)
	err = websocket.Message.Send(conn2, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)

	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)
	if f["command"] != "login" {
		log.Println("Test: ERROR: Expected login, got", msg)
	}
	if f["result"] != "success" {
		log.Println("Test: ERROR: Expected success, got ", msg)
	}

	// open connection as player 2
	log.Println("Test: Open third Connection")
	conn3, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	log.Println("Test: Registering as player 2")
	// send "join" command as player 1
	b = []byte(`{"command": "register", "name" : "foo2", "password" : "bar2"}`)
	err = websocket.Message.Send(conn3, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	err = websocket.Message.Receive(conn3, &msg)
	checkError(err)

	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)
	if f["command"] != "register" {
		log.Println("Test: ERROR: Expected register, got", msg)
	}
	if f["result"] != "success" {
		log.Println("Test: ERROR: Expected success, got ", msg)
	}

	log.Println("Test: ----- loginAndRegister() ended -----")
}

func game(t *testing.T) {
	log.Println("Test: ----- game() procedure -----")

	service := "ws://localhost:8080/play/"

	// open connection as player 1
	log.Println("Test: Connecting as player 1")
	conn, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	log.Println("Test: Loggin in as player 1")
	// send "join" command as player 1
	b := []byte(`{"command": "login", "name" : "foo1", "password" : "bar1"}`)
	err = websocket.Message.Send(conn, b)
	checkError(err)

	// receive answer from registration
	var msg string
	err = websocket.Message.Receive(conn, &msg)
	checkError(err)

	var f map[string]interface{}
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)
	if f["command"] != "login" {
		log.Println("Test: ERROR: Expected register, got", msg)
	}
	if f["result"] != "success" {
		log.Println("Test: ERROR: Expected success, got ", msg)
	}
	log.Println("Test: Player 1 logged in")

	// open connection as player 2
	log.Println("Test: Connecting as player 2")
	conn2, err := websocket.Dial(service, "", "http://localhost")
	checkError(err)

	log.Println("Test: Loggin in as player 2")
	// send "join" command as player 2
	b = []byte(`{"command": "login", "name" : "foo2", "password" : "bar2"}`)
	err = websocket.Message.Send(conn2, b)
	checkError(err)

	// receive answer from registration
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)

	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)
	if f["command"] != "login" {
		log.Println("Test: ERROR: Expected register, got", msg)
	}
	if f["result"] != "success" {
		log.Println("Test: ERROR: Expected success, got ", msg)
	}
	log.Println("Test: Player 2 logged in")

	// send "join" command as player 1
	log.Println("Test: Joining as player 1")
	b = []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn, b)
	checkError(err)

	// send "join" command as player 2
	log.Println("Test: Joining as player 2")
	b = []byte(`{"command": "join"}`)
	err = websocket.Message.Send(conn2, b)
	checkError(err)

	// receive "matched" message from server as player 1
	err = websocket.Message.Receive(conn, &msg)
	checkError(err)

	// receive "matched" message from server as player 2
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)

	// interpret message as JSON data
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	if f["command"] != "matched" {
		t.Error("Expected matched, got ", msg)
	}

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
	log.Println("Player one gets 0")

	// receive game result as player 2
	err = websocket.Message.Receive(conn2, &msg)
	checkError(err)
	if msg != "3" {
		t.Error("Expected 3, got ", msg)
	}
	log.Println("Player one gets 3")

	log.Println("Test: ----- game() ended -----")
}
