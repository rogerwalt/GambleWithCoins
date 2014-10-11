package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"

	"code.google.com/p/go.net/websocket"
)

type User struct {
	name string
	bet int					// bet of user in current round
	conn *websocket.Conn
}

func sendError(ws *websocket.Conn, err error) {
	log.Println("Sending error: ", err.Error())
	toSend, _ := json.Marshal(map[string]string{"error": err.Error()})
	websocket.Message.Send(ws, toSend)
}

// returns a User if a user has successfully authenticated himself,
// otherwise returns an error
func authenticate(ws *websocket.Conn) (*User, error) {
	var msg string
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	json.Unmarshal([]byte(msg), &m)

/*
	fmt.Printf("Map: %v", m)
	fmt.Printf("Name: %v", m["name"])
	fmt.Printf("Password: %v", m["password"])
	fmt.Println("___________________________")
*/

	for i := 0; i < 3; i++ {
		if m["command"].(string) == "login" {
			if masc.Login(m["name"].(string), m["password"].(string)) {
				log.Println("Client logged in")
				b := []byte(`{"command": "login", "result" : "success"}`)
				err = websocket.Message.Send(ws, b)
				if err != nil {
					return nil, err
				}
				return &User{m["name"].(string), 0, ws}, nil
			} else {
				return nil, errors.New("Wrong password")
			}
		} else if m["command"].(string) == "register" {
			err := masc.Register(m["name"].(string), m["password"].(string))
			if err != nil {
				return nil, err
			} else {
				log.Println("Client registered")
				b := []byte(`{"command": "register", "result" : "success"}`)
				err = websocket.Message.Send(ws, b)
				if err != nil {
					return nil, err
				}
				return &User{m["name"].(string), 0, ws}, nil
			}
		}
	}
	disconnectClient(nil, ws)
	return nil, errors.New("Too many unsuccessful logins")
}

func makeGame(ready chan *User, close chan bool) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Client connected")

		user, err := authenticate(ws)
		if err != nil {
			sendError(ws, err)
			ws.Close()
			return
		}

		var msg string
		err = websocket.Message.Receive(ws, &msg)
		checkError(err)

		// interpret message as json data
		// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
		var f map[string]interface{}
		err = json.Unmarshal([]byte(msg), &f)

		_, commandExists := f["command"]

		// remove client if sends invalid data
		if err != nil || !commandExists {
			disconnectClient(user, ws)
			return
		}

		// check what the command is; here only join is allowed
		if f["command"] == "join" {
			fmt.Println("Client wants to join")
			ready <- user
		}
		<-close
		ws.Close()
	}
}

func Hub(ready chan *User) {
	waitingUsers := make([]*User, 0, 2)
	for {
		select {
		case u := <-ready:
			if len(waitingUsers) > 0 {
				log.Println("Matching clients")

				uWaiting := waitingUsers[len(waitingUsers)-1]
				waitingUsers = waitingUsers[:len(waitingUsers)-1]

				err := websocket.Message.Send(uWaiting.conn, []byte(`{"command": "matched"}`))
				if err != nil {
					uWaiting.conn.Close()
					waitingUsers = append(waitingUsers, u)
				}

				err = websocket.Message.Send(u.conn, []byte(`{"command": "matched"}`))
				if err != nil {
					u.conn.Close()
					waitingUsers = append(waitingUsers, uWaiting)
				}

				go handleGame(uWaiting, u)

			} else {
				log.Println("Client joined")
				waitingUsers = append(waitingUsers, u)
			}
		}
	}
}

func handleGame(user1, user2 *User) {
	var action1, action2 string
	err := websocket.Message.Receive(user1.conn, &action1)
	checkError(err)

	// interpret message as json data
	// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
	var f map[string]interface{}
	err = json.Unmarshal([]byte(action1), &f)

	_, commandExists := f["command"]

	// remove client if sends invalid data
	if err != nil || !commandExists {
		disconnectClient(user1, user1.conn)
		return
	}

	// try to convert action1 to string
	if str, ok := f["command"].(string); ok {
		action1 = str
	} else {
		disconnectClient(user1, user1.conn)
		return
	}

	err = websocket.Message.Receive(user2.conn, &action2)
	checkError(err)

	// interpret message as json data
	// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
	err = json.Unmarshal([]byte(action2), &f)

	_, commandExists = f["command"]

	// remove client if sends invalid data
	if err != nil || !commandExists {
		disconnectClient(user2, user2.conn)
		return
	}

	// try to convert action2 to string
	if str, ok := f["command"].(string); ok {
		action2 = str
	} else {
		disconnectClient(user1, user1.conn)
		return
	}

	log.Println("Received actions:")
	p1, p2 := masc.PrisonersDilemma(action1, action2)
	err = websocket.Message.Send(user1.conn, strconv.Itoa(p1))
	checkError(err)

	err = websocket.Message.Send(user2.conn, strconv.Itoa(p2))
	checkError(err)
	log.Println("Sent payoffs")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("index.html")
	checkError(err)

	fmt.Fprintf(w, "%s", page)
}

func main() {
	serverClose := make(chan int)
	start("./masc.db", serverClose)
}

func start(dbName string, serverClose chan int) {
	log.Println("Starting server")

	db, err := sql.Open("sqlite3", dbName)
	checkError(err)
	masc.SetupDb(db)
	defer db.Close()

	ready := make(chan *User)
	close := make(chan bool)

	go Hub(ready)

	http.HandleFunc("/", staticHandler)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.Handle("/play/", websocket.Handler(makeGame(ready, close)))
	s := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	listener, err := net.Listen("tcp", ":8080")
	checkError(err)
	go s.Serve(listener)

	select {

	case <-serverClose:
		listener.Close()
	}
	return
}

func checkError(err error) {
	if err != nil {
		log.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func disconnectClient(user *User, ws *websocket.Conn) {
	log.Println("Disconnecting client due to invalid requests.")
	toSend, _ := json.Marshal(map[string]string{"error": "Disconnecting client due to invalid requests."})
	websocket.Message.Send(ws, toSend)
	ws.Close()
	if (user != nil) {
		masc.UpdateBalance(user.name, -user.bet)
		log.Println("User ", user.name, " loses his bet of ", user.bet)
	}
}