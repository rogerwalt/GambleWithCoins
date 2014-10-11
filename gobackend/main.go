package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
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
	conn *websocket.Conn
}

type ApiError struct {
	message string // api error message
	code    int    // api error code
	command string // which api command was executed while the error happened
}

func sendError(ws *websocket.Conn, err ApiError) {
	log.Printf("API Error: Code %i, \"%s\", Command \"%s\"", err.code, err.message, err.command)
	websocket.Message.Send(ws, string([]byte(`{"command": "`+err.command+`", "result": { "errorCode": `+strconv.Itoa(err.code)+`, "errorMsg": "`+err.message+`"}}`)))
}

// returns a User if a user has successfully authenticated himself,
// otherwise returns an error
func authenticate(ws *websocket.Conn) (*User, *ApiError) {
	var msg string
	var e ApiError
	err := websocket.Message.Receive(ws, &msg)
	fmt.Printf("Message: %v", msg)
	if err != nil {
		e.message = "Could not receive data from client:" + err.Error()
		e.code = 98
		e.command = "_undefined"
		return nil, &e
	}
	log.Printf("Authenticate: Received data from client with RemoteAddr: %v", ws.RemoteAddr())
	log.Printf(msg)
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
				b := string([]byte(`{"command": "login", "result" : "success"}`))
				err = websocket.Message.Send(ws, b)
				if err != nil {
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					return nil, &e
				}
				log.Println("Authenticate:", m["name"], "logged in")
				return &User{m["name"].(string), ws}, nil
			} else {
				e.message = "Wrong username or password."
				e.code = 1
				e.command = "login"
				log.Printf("Authenticate: %v entered wront username or password", m["name"].(string))
				return nil, &e
			}
		} else if m["command"].(string) == "register" {
			err := masc.Register(m["name"].(string), m["password"].(string))
			if err != nil {
				e.message = "Could not register new user:" + err.Error()
				e.code = 999
				e.command = "register"
				log.Println("Authenticate:", m["name"], "could not register")
				return nil, &e
			} else {
				log.Println("Client registered")
				b := string([]byte(`{"command": "register", "result" : "success"}`))
				err = websocket.Message.Send(ws, b)
				if err != nil {
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					log.Println("Authenticate:", m["name"], "did not receive data")
					return nil, &e
				}
				log.Println("Authenticate:", m["name"], "registered successfully")
				return &User{m["name"].(string), ws}, nil
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
	log.Println("Authenticate: Too many unsuccessful logins")
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

		log.Println("makeGame:", user.name, "Successfully authenticated")
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
				b := fmt.Sprintf(`{"command" : "balance", "result" : %d}`, balance)
				err = websocket.Message.Send(user.conn, string(b))
			case "getDepositAddress":
				address, _ := masc.GetDepositAddress(user.name)
				b := string([]byte(fmt.Sprintf(`{"command" : "depositAddress", "result" : %d}`,
					address)))
				err = websocket.Message.Send(user.conn, b)
			case "withdraw":
				address := f["address"].(string)
				amount := f["amount"].(int)
				err := masc.Withdraw(user.name, amount, address)
				var b []byte
				if err != nil {
					b = []byte(fmt.Sprintf(`{"command" : "withdraw", "result" : {"error": "%s"}}`, err.Error()))
				} else {
					b = []byte(`{"command" : "withdraw", "result" : "success"}`)
				}
				err = websocket.Message.Send(user.conn, string(b))
			}
			<-close
			ws.Close()
		}
		log.Println("makeGame: End")
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

func receiver(ws *websocket.Conn) (<-chan string, chan error) {
	ch, errCh := make(chan string), make(chan error)
	go func() {
		for {
			var s string
			err := websocket.Message.Receive(ws, &s)
			if err != nil {
				errCh <- err
				close(ch)
				return
			}
			ch <- s
		}
	}()
	return ch, errCh
}

func sender(ws *websocket.Conn) (chan string, chan error) {
	ch, errCh := make(chan string), make(chan error)
	go func() {
		for {
			s := <-ch
			if err := websocket.Message.Send(ws, s); err != nil {
				errCh <- err
				close(ch)
				return
			}
		}
	}()
	return ch, errCh
}

// takes on input channel the player a player that has chosen his action
// signals on channel ready when both players have chosen
func choseNotifier() (chan int, chan int) {
	ready := make(chan int)
	input := make(chan int)
	player1Chosen := false
	player2Chosen := false
	go func() {
		for {
			x := <-input
			if x == 1 {
				player1Chosen = true
			} else {
				player2Chosen = true
			}
			if player1Chosen && player2Chosen {
				ready <- 1
				return
			}
		}
	}()
	return input, ready
}

func handleGameRound(user1, user2 *User, b, E int) {

	timer := time.NewTimer(time.Second * 30)
	chose, ready := choseNotifier()
	recv1, _ := receiver(user1.conn)
	sender1, _ := sender(user1.conn)
	recv2, _ := receiver(user2.conn)
	sender2, _ := sender(user2.conn)
	// send startRound to players
	sender1 <- `{"command" : "startRound"}`
	sender2 <- `{"command" : "startRound"}`

	log.Println("Waiting for players actions in round.")
	action1, action2 := "", ""

	receiveAction := func(msg string, sender chan string,
		chose chan int, player int, action string) string {
		log.Println("Received something")
		var f map[string]interface{}
		json.Unmarshal([]byte(msg), &f)

		if f["command"] == "action" {
			action = f["action"].(string)
			log.Println("Received action from player", player, action)
			sender <- `{"command" : "action", "action" : "chosen"}`
			chose <- player
			return action
		} else if f["command"] == "signal" {
			signal := f["signal"]
			log.Println("Received signal from player", player, signal)
			sender <- fmt.Sprintf(`{"command" : "signal", "signal" : %s}`, signal)
		}
		return action
	}

	// wait for clients to send something
ActionLoop:
	for {
		select {
		case msg := <-recv1:
			action1 = receiveAction(msg, sender2, chose, 1, action1)
		case msg := <-recv2:
			action2 = receiveAction(msg, sender1, chose, 2, action2)
		case <-timer.C:
			log.Println("Time is up")
			break ActionLoop
		case <-ready:
			log.Println("Both Players have chosen their actions")
			break ActionLoop
		}

	}

	// check if all players have chosen
	switch {
	case action1 == "" && action2 == "":
		log.Println("Nobody has chosen")
		// bank wins
	case action1 == "":
		log.Println("Player 1 has not chosen")
		//player2 wins
	case action2 == "":
		log.Println("Player 2 has not chosen")
		//player1 wins
	}

	log.Println("Received actions: ", action1, action2)
	p1, p2 := masc.PrisonersDilemma(action1, action2, b, E)

	err := websocket.Message.Send(user1.conn,
		fmt.Sprintf(`{ "command": "endRound", 
						"outcome" : {"me" : %s, "other" : %s}, 
						"balanceDifference" : {"me" : %d, 
											   "other" : %d}}`,
			action1, action2, p1, p2))
	checkError(err)

	err = websocket.Message.Send(user2.conn,
		fmt.Sprintf(`{ "command": "endRound", 
						"outcome" : {"me" : %s, "other" : %s}, 
						"balanceDifference" : {"me" : %d, 
											   "other" : %d}}`,
			action2, action1, p2, p1))
	checkError(err)
	log.Println("Sent payoffs", p1, p2)

	log.Println("Update balances")
	masc.UpdateBalance(user1.name, p1)
	masc.UpdateBalance(user2.name, p2)

	return
}

//TODO: handle more errors
func handleGame(user1, user2 *User) {
	bet := 1000
	p := 0.2
	E := int(p * 2 * float64(bet))

RoundLoop:
	for {
		log.Println("New game round")

		// check if players have enough funds to play a round
		balance1, _ := masc.GetBalance(user1.name)
		balance2, _ := masc.GetBalance(user2.name)
		log.Println("Players balance: ", balance1, balance2)

		endGame := func() {
			err := websocket.Message.Send(user1.conn, []byte(`{"command": "endGame"}`))
			checkError(err)
			err = websocket.Message.Send(user2.conn, []byte(`{"command": "endGame"}`))
			checkError(err)
		}
		switch {
		case balance1 < bet && balance2 < bet:
			log.Println("Both players have not enough funds left")
			endGame()
			break RoundLoop
		case balance1 < bet:
			// player 2 gets the remaining money and the game ends
			log.Println("Player 1 does not have enough funds left")
			endGame()
			break RoundLoop
		case balance2 < bet:
			//player 1 gets the remaining money and the game ends
			log.Println("Player 2 does not have enough funds left")
			endGame()
			break RoundLoop
		}

		// players have sufficient funds for the next round
		// check if game ends
		//TODO: TODO: TODO: TODO: TODO: TODO: use crypto secure rng
		r := rand.Intn(100)
		if float64(r) <= 100*p {
			log.Println("End game: bank wins bets.")
			masc.UpdateBalance(user1.name, -bet)
			masc.UpdateBalance(user2.name, -bet)
			//end game
			endGame()
			break RoundLoop
		}

		handleGameRound(user1, user2, bet, E)
	}

	return
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("index.html")
	checkError(err)

	fmt.Fprintf(w, "%s", page)
}

func main() {
	serverClose := make(chan int)
	start("./masc.db", 8080, serverClose, time.Now().UTC().UnixNano())
}

func start(dbName string, port int, serverClose chan int, seed int64) {
	log.Println("Starting server")

	rand.Seed(seed)
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
	toSend := string([]byte(`{"result": {"errorMsg": "Disconnecting client due to invalid requests.", "errorCode": ` + strconv.Itoa(10) + `}}`))
	websocket.Message.Send(ws, toSend)
	ws.Close()
}
