package main

import (
	"database/sql"
	"encoding/json"
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
	bet  int // bet of user in current round
	conn *websocket.Conn
}

type ApiError struct {
	message string 		// api error message
	code int 			// api error code
	command string		// which api command was executed while the error happened
}

func sendError(ws *websocket.Conn, err ApiError) {
	log.Printf("API Error: Code %i, \"%s\", Command \"%s\"", err.code, err.message, err.command)
	websocket.Message.Send(ws, []byte(`{"command": "` + err.command + `", "result": { "errorCode": ` + strconv.Itoa(err.code) + `, "errorMsg": "` + err.message + `"}}`))
}

// returns a User if a user has successfully authenticated himself,
// otherwise returns an error
func authenticate(ws *websocket.Conn) (*User, *ApiError) {
	var msg string
	var e ApiError
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		e.message = "Could not receive data from client:" + err.Error()
		e.code = 98
		e.command = "_undefined"
		return nil, &e
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
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					return nil, &e
				}
				return &User{m["name"].(string), 0, ws}, nil
			} else {
				e.message = "Wrong username or password."
				e.code = 1
				e.command = "login"
				return nil, &e
			}
		} else if m["command"].(string) == "register" {
			err := masc.Register(m["name"].(string), m["password"].(string))
			if err != nil {
				e.message = "Could not register new user:" + err.Error()
				e.code = 999
				e.command = "register"
				return nil, &e
			} else {
				log.Println("Client registered")
				b := []byte(`{"command": "register", "result" : "success"}`)
				err = websocket.Message.Send(ws, b)
				if err != nil {
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					return nil, &e
				}
				return &User{m["name"].(string), 0, ws}, nil
			}
		}
	}
	disconnectClient(nil, ws)
	e.message = "Too many unsuccessful logins."
	e.code = 2
	if m["command"].(string) == "login" {
		e.command = "login"
	} else if m["command"].(string) == "register" {
		e.command = "register"
	} else {
		e.command = "_undefined"
	}
	return nil, &e
}

func makeGame(ready chan *User, close chan bool) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		log.Println("Client connected")

		user, e := authenticate(ws)
		if e != nil {
			sendError(ws, *e)
			ws.Close()
			return
		}

		for {
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
				disconnectClient(user, ws)
				return
			}

			// check what the command is; here only join is allowed
			switch f["command"] {
			case "join":
				fmt.Println("Client wants to join")
				ready <- user
			case "getBalance":
				balance, _ := masc.GetBalance(user.name)
				b := []byte(fmt.Sprintf(`{"command" : "balance", "result" : %d}`, balance))
				err = websocket.Message.Send(user.conn, b)
			case "getDepositAddress":
				address, _ := masc.GetDepositAddress(user.name)
				b := []byte(fmt.Sprintf(`{"command" : "depositAddress", "result" : %d}`,
					address))
				err = websocket.Message.Send(user.conn, b)
			case "withdraw":
				address := f["address"].(string)
				amount := f["amount"].(int)
				err := masc.Withdraw(user.name, amount, address)
				var b []byte
				if err != nil {
					b = []byte(fmt.Sprintf(`{"command" : "withdraw", "result" : 
												{"error": "%s"}}`, err.Error()))
				} else {
					b = []byte(`{"command" : "withdraw", "result" : "success"}`)
				}
				err = websocket.Message.Send(user.conn, b)
			}
			<-close
			ws.Close()
		}
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

	log.Println("Received actions.")
	p1, p2 := masc.PrisonersDilemma(action1, action2)
	masc.AddAction(user1.name, action1)
	masc.AddAction(user2.name, action2)
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
	start("./masc.db", 8080, serverClose)
}

func start(dbName string, port int, serverClose chan int) {
	log.Println("Starting server")

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

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
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
	toSend, _ := json.Marshal(map[string]string{"errorMsg": "Disconnecting client due to invalid requests.", "errorCode": strconv.Itoa(10)})
	websocket.Message.Send(ws, toSend)
	ws.Close()
	if user != nil {
		masc.UpdateBalance(user.name, -user.bet)
		log.Println("User ", user.name, " loses his bet of ", user.bet)
	}
}
