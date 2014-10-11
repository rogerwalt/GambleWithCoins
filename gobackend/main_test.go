package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"

	"github.com/gorilla/websocket"
)

// TODO: shutting down the server gracefully does not yet work
func startServer() chan int {
	time.Sleep(100 * time.Millisecond)
	// use fresh db
	os.Remove("./test.db")
	serverClose := make(chan int)
	go start("./test.db", 8080, serverClose, 0)

	time.Sleep(100 * time.Millisecond)
	return serverClose
}

func TestGame(t *testing.T) {
	log.Println("Test: Starting server")
	serverClose := startServer()

	// do not change order: probabilistic/seed dependent!!
	loginAndRegister(t)
	testDatabase(t)
	game(t)
	/* gameB(t)*/
	//gameC(t)
	/*gameSlow(t)*/

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

	var DefaultDialer *websocket.Dialer
	// open connection as player 1
	log.Println("Test: Connecting as player 1")

	conn, _, err := DefaultDialer.Dial(service, nil)
	checkError(err)

	log.Println("Test: Registering as player 1")
	// send "join" command as player 1
	b := []byte(`{"command": "register", "name" : "foo1", "password" : "bar1"}`)
	err = conn.WriteMessage(websocket.TextMessage, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	messageType, msg, err := conn.ReadMessage()
	//err = conn.WriteMessage(websocket.TextMessage, b)
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
	conn2, _, err := DefaultDialer.Dial(service, nil)
	checkError(err)

	log.Println("Test: Login as player 1")
	// send "join" command as player 1
	b = []byte(`{"command": "login", "name" : "foo1", "password" : "bar1"}`)
	err = conn2.WriteMessage(messageType, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	messageType, msg, err = conn2.ReadMessage()
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
	conn3, _, err := DefaultDialer.Dial(service, nil)
	checkError(err)

	log.Println("Test: Registering as player 2")
	// send "join" command as player 1
	b = []byte(`{"command": "register", "name" : "foo2", "password" : "bar2"}`)
	err = conn3.WriteMessage(messageType, b)
	// fatal error?
	checkError(err)

	// receive answer from registration
	messageType, msg, err = conn3.ReadMessage()
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

func joinGame(name1, pass1 string, balance1 int,
	name2, pass2 string, balance2 int) (*websocket.Conn, *websocket.Conn) {
	log.Println("Test: ----- game() procedure -----")

	service := "ws://localhost:8080/play/"

	// open connection as player 1
	log.Println("Test: Connecting as player 1")
	var DefaultDialer *websocket.Dialer
	conn, _, err := DefaultDialer.Dial(service, nil)
	checkError(err)

	log.Println("Test: Loggin in as player 1")
	// send "join" command as player 1
	b := []byte(fmt.Sprintf(
		`{"command": "login", "name" : "%s", "password" : "%s"}`, name1, pass1))
	err = conn.WriteMessage(websocket.TextMessage, b)
	checkError(err)

	// receive answer from registration
	messageType, msg, err := conn.ReadMessage()
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
	conn2, _, err := DefaultDialer.Dial(service, nil)
	checkError(err)

	log.Println("Test: Loggin in as player 2")
	// send "join" command as player 2
	b = []byte(fmt.Sprintf(`{"command": "login", "name" : "%s", "password" : "%s"}`,
		name2, pass2))
	err = conn2.WriteMessage(messageType, b)
	checkError(err)

	// receive answer from registration
	messageType, msg, err = conn2.ReadMessage()
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

	// add funds
	masc.UpdateBalance(name1, balance1)
	masc.UpdateBalance(name2, balance2)

	// send "join" command as player 1
	log.Println("Test: Joining as player 1")
	b = []byte(`{"command": "join"}`)
	err = conn.WriteMessage(messageType, b)
	checkError(err)

	// send "join" command as player 2
	log.Println("Test: Joining as player 2")
	b = []byte(`{"command": "join"}`)
	err = conn2.WriteMessage(messageType, b)
	checkError(err)

	// receive "matched" message from server as player 1
	messageType, msg, err = conn.ReadMessage()
	checkError(err)

	// receive "matched" message from server as player 2
	messageType, msg, err = conn2.ReadMessage()
	checkError(err)

	// interpret message as JSON data
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	if f["command"] != "matched" {
		log.Println("Expected matched, got ", msg)
	}

	// interpret message as JSON data
	err = json.Unmarshal([]byte(msg), &f)
	checkError(err)

	if f["command"] != "matched" {
		log.Println("Expected matched, got ", msg)
	}

	return conn, conn2
}

func game(t *testing.T) {
	//var conn1, conn2 *websocket.Conn
	conn1, conn2 := joinGame("foo1", "bar1", 10000, "foo2", "bar2", 10000)

	for i := 0; i < 2; i++ {
		// receive start round as player 1
		messageType, msg, err := conn1.ReadMessage()
		checkError(err)
		log.Println(string(msg))

		// receive start round as player 2
		messageType, msg, err = conn2.ReadMessage()
		checkError(err)
		log.Println(string(msg))

		// send "cooperate" command as player 1
		b := []byte(`{"command": "action", "action" : "cooperate"}`)
		err = conn1.WriteMessage(messageType, b)
		checkError(err)

		// receive notification as player 2
		messageType, msg, err = conn2.ReadMessage()
		checkError(err)
		log.Println(string(msg))

		// send "defect" command as player 2
		b = []byte(`{"command": "action", "action": "defect"}`)
		err = conn2.WriteMessage(messageType, b)
		checkError(err)

		// receive notification as player 1
		messageType, msg, err = conn1.ReadMessage()
		checkError(err)
		log.Println(string(msg))

		// receive game notification as player 1
		messageType, msg, err = conn1.ReadMessage()
		checkError(err)
		log.Println(string(msg))

		// receive game notification as player 2
		messageType, msg, err = conn2.ReadMessage()
		checkError(err)
		log.Println(string(msg))
	}

	log.Println("Test: ----- game() ended -----")
}

func gameSlow(t *testing.T) {
	//var conn1, conn2 *websocket.Conn
	conn1, conn2 := joinGame("foo1", "bar1", 10000, "foo2", "bar2", 10000)

	// receive start round as player 1
	messageType, msg, err := conn1.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	// receive start round as player 2
	messageType, msg, err = conn2.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	// send "cooperate" command as player 1
	b := []byte(`{"command": "action", "action" : "cooperate"}`)
	err = conn1.WriteMessage(messageType, b)
	checkError(err)

	// receive notification as player 2
	messageType, msg, err = conn2.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	// send "defect" command as player 2
	b = []byte(`{"command": "action", "action": "defect"}`)
	err = conn2.WriteMessage(messageType, b)
	checkError(err)

	// receive notification as player 1
	messageType, msg, err = conn1.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	// receive game notification as player 1
	messageType, msg, err = conn1.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	// receive game notification as player 2
	messageType, msg, err = conn2.ReadMessage()
	checkError(err)
	log.Println(string(msg))

	log.Println("Test: ----- game() ended -----")

}
