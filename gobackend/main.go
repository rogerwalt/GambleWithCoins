package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"

	"code.google.com/p/go.net/websocket"
)

func makeGame(ready chan *websocket.Conn, close chan bool) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Client connected")
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		checkError(err)

		// interpret message as json data
		// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
		var f map[string]interface{}
		err = json.Unmarshal([]byte(msg), &f)

		_, commandExists := f["command"]

		// remove client if sends invalid data
		if err != nil || !commandExists {
			log.Println("Remove client due to invalid requests.")
			ws.Close()
			return
		}

		// check what the command is; here only join is allowed
		if f["command"] == "join" {
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
				log.Println("Matching clients")

				cWaiting := waitingClients[len(waitingClients)-1]
				waitingClients = waitingClients[:len(waitingClients)-1]

				err := websocket.Message.Send(cWaiting, []byte(`{"command": "matched"}`))
				if err != nil {
					cWaiting.Close()
					waitingClients = append(waitingClients, c)
				}

				err = websocket.Message.Send(c, []byte(`{"command": "matched"}`))
				if err != nil {
					c.Close()
					waitingClients = append(waitingClients, cWaiting)
				}

				go handleGame(cWaiting, c)

			} else {
				log.Println("Appending client")
				waitingClients = append(waitingClients, c)
			}
		}
	}
}

func handleGame(conn1, conn2 *websocket.Conn) {
	var action1, action2 string
	err := websocket.Message.Receive(conn1, &action1)
	checkError(err)

	// interpret message as json data
	// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
	var f map[string]interface{}
	err = json.Unmarshal([]byte(action1), &f)

	_, commandExists := f["command"]

	// remove client if sends invalid data
	if err != nil || !commandExists {
		log.Println("Remove client due to invalid requests.")
		conn1.Close()
		return
	}

	// try to convert action1 to string
	if str, ok := f["command"].(string); ok {
	    action1 = str
	} else {
	    log.Println("Remove client due to invalid requests.")
		conn1.Close()
		return
	}

	err = websocket.Message.Receive(conn2, &action2)
	checkError(err)

	// interpret message as json data
	// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
	err = json.Unmarshal([]byte(action2), &f)

	_, commandExists = f["command"]

	// remove client if sends invalid data
	if err != nil || !commandExists {
		log.Println("Remove client due to invalid requests.")
		conn2.Close()
		return
	}

	// try to convert action2 to string
	if str, ok := f["command"].(string); ok {
	    action2 = str
	} else {
	    log.Println("Remove client due to invalid requests.")
		conn1.Close()
		return
	}

	log.Println("Received actions:")
	p1, p2 := masc.PrisonersDilemma(action1, action2)
	err = websocket.Message.Send(conn1, strconv.Itoa(p1))
	checkError(err)

	err = websocket.Message.Send(conn2, strconv.Itoa(p2))
	checkError(err)
	log.Println("Sent payoffs")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("index.html")
	checkError(err)

	fmt.Fprintf(w, "%s", page)
}

func main() {
	db, err := sql.Open("sqlite3", "./masc.db")
	checkError(err)
	masc.SetupDb(db)
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
		log.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
