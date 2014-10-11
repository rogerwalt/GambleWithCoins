package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"

	"code.google.com/p/go.net/websocket"
)

// global database connection
var db *sql.DB

func makeGame(ready chan *websocket.Conn, close chan bool) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		fmt.Println("Client connected")
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		checkError(err)

		// interpret message as json data
		// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
		var f interface{}
		err = json.Unmarshal([]byte(msg), &f)
		checkError(err)

		fmt.Printf("json received: %v", f)

		msg = strings.Trim(msg, "\"")
		if msg == "join" {
			fmt.Println("Client wants to join")
			ready <- ws
		}
		<-close
		ws.Close()
	}
}

func Hub(ready chan *websocket.Conn) {
	waitingClients := make([]*websocket.Conn, 0, 2)
	for {
		select {
		case c := <-ready:
			if len(waitingClients) > 0 {
				fmt.Println("Matching clients")

				cWaiting := waitingClients[len(waitingClients)-1]
				waitingClients = waitingClients[:len(waitingClients)-1]

				err := websocket.Message.Send(cWaiting, "matched")
				if err != nil {
					cWaiting.Close()
					waitingClients = append(waitingClients, c)
				}

				err = websocket.Message.Send(c, "matched")
				if err != nil {
					c.Close()
					waitingClients = append(waitingClients, cWaiting)
				}

				go handleGame(cWaiting, c)

			} else {
				fmt.Println("Appending client")
				waitingClients = append(waitingClients, c)
			}
		}
	}
}

func handleGame(conn1, conn2 *websocket.Conn) {
	var action1, action2 string
	err := websocket.Message.Receive(conn1, &action1)
	checkError(err)
	err = websocket.Message.Receive(conn2, &action2)
	checkError(err)

	fmt.Println("Received actions:")
	fmt.Printf("%v", action1)
	p1, p2 := masc.PrisonersDilemma(action1, action2)
	err = websocket.Message.Send(conn1, strconv.Itoa(p1))
	checkError(err)

	err = websocket.Message.Send(conn2, strconv.Itoa(p2))
	checkError(err)
	fmt.Println("Sent payoffs")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("index.html")
	checkError(err)

	fmt.Fprintf(w, "%s", page)
}

func main() {
	db, err := sql.Open("sqlite3", "./masc.db")
	checkError(err)
	defer db.Close()

	ready := make(chan *websocket.Conn)
	close := make(chan bool)

	go Hub(ready)

	http.HandleFunc("/", staticHandler)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.Handle("/play/", websocket.Handler(makeGame(ready, close)))
	err = http.ListenAndServe(":8080", nil)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
