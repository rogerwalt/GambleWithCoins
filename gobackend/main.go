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
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rogerwalt/GambleWithCoins/gobackend/bitcoin"

	"github.com/rogerwalt/GambleWithCoins/gobackend/masc"

	"github.com/gorilla/websocket"
)

type User struct {
	name          string
	conn          *websocket.Conn
	recvChan      chan string
	recvErrorChan chan error
	sendChan      chan string
	sendErrorChan chan error
}

func receiver(conn *websocket.Conn) (chan string, chan error) {
	ch, errCh := make(chan string), make(chan error)
	go func() {
		for {
			_, s, err := conn.ReadMessage()
			//pongWait := *time.Second
			//conn.SetReadDeadline(time.Now().Add(pongWait))
			if err != nil {
				log.Println(err.Error())
				errCh <- err
				close(ch)
				return
			}
			ch <- string(s)
		}
	}()
	return ch, errCh
}

func sender(conn *websocket.Conn) (chan string, chan error) {
	ch, errCh := make(chan string), make(chan error)
	go func() {
		for {
			s := <-ch
			if err := conn.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
				log.Println(err.Error())
				errCh <- err
				close(ch)
				return
			}
		}
	}()
	return ch, errCh
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ApiError struct {
	message string // api error message
	code    int    // api error code
	command string // which api command was executed while the error happened
}

func sendError(conn *websocket.Conn, err ApiError) {
	log.Printf("API Error: Code %i, \"%s\", Command \"%s\"", err.code, err.message, err.command)
	conn.WriteMessage(websocket.TextMessage, []byte(`{"command": "`+err.command+`", "result": { "errorCode": `+strconv.Itoa(err.code)+`, "errorMsg": "`+err.message+`"}}`))
}

func newUser(name string, conn *websocket.Conn) *User {
	recvChan, recvErrorChan := receiver(conn)
	sendChan, sendErrorChan := sender(conn)
	return &User{name, conn, recvChan, recvErrorChan, sendChan, sendErrorChan}
}

// returns a User if a user has successfully authenticated himself,
// otherwise returns an error
func authenticate(conn *websocket.Conn) (*User, *ApiError) {
	var e ApiError
	for i := 0; i < 3; i++ {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			e.message = "Could not receive data from client:" + err.Error()
			e.code = 98
			e.command = "_undefined"
			return nil, &e
		}
		log.Printf(string(msg))
		log.Printf("Authenticate: Received data from client with RemoteAddr: %v", conn.RemoteAddr())
		var m map[string]interface{}
		json.Unmarshal([]byte(msg), &m)

		log.Printf("Map: %v", m)
		log.Printf("Name: %v", m["name"])
		log.Printf("Password: %v", m["password"])
		log.Println("___________________________")

		if m["command"].(string) == "login" {
			if masc.Login(m["name"].(string), m["password"].(string)) {
				log.Println("Client logged in")
				b := []byte(`{"command": "login", "result" : "success"}`)
				err = conn.WriteMessage(messageType, b)
				if err != nil {
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					return nil, &e
				}
				log.Println("Authenticate:", m["name"], "logged in")
				return newUser(m["name"].(string), conn), nil
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
				b := []byte(`{"command": "register", "result" : "success"}`)
				err = conn.WriteMessage(messageType, b)
				if err != nil {
					e.message = "Could not send back data to client:" + err.Error()
					e.code = 99
					e.command = "login"
					log.Println("Authenticate:", m["name"], "did not receive data")
					return nil, &e
				}
				log.Println("Authenticate:", m["name"], "registered successfully")
				return newUser(m["name"].(string), conn), nil
			}
		}
	}
	disconnectClient(nil, conn)
	e.message = "Too many unsuccessful logins."
	e.code = 2
	/*if m["command"].(string) == "login" {*/
	//e.command = "login"
	//} else if m["command"].(string) == "register" {
	//e.command = "register"
	//} else {
	//e.command = "_undefined"
	/*}*/
	log.Println("Authenticate: Too many unsuccessful logins")
	return nil, &e
}

func makeGame(ready chan *User, hubDone chan chan int, close chan bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Client connected")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		user, e := authenticate(conn)
		if e != nil {
			sendError(conn, *e)
			conn.Close()
			return
		}

		log.Println("makeGame:", user.name, "Successfully authenticated")
		for {
			msg := <-user.recvChan

			// interpret message as json data
			// errors like "Fatal error  invalid character 'j' looking for beginning of value" are because of invalid JSON data
			var f map[string]interface{}
			err = json.Unmarshal([]byte(msg), &f)

			_, commandExists := f["command"]

			// remove client if sends invalid data
			if err != nil || !commandExists {
				disconnectClient(user, conn)
				return
			}

			// check what the command is;
			switch strings.Trim(f["command"].(string), "\"") {
			case "join":
				fmt.Println("Client to the hub")
				ready <- user
				done := make(chan int)
				hubDone <- done
				<-done
			case "getBalance":
				balance, _ := masc.GetBalance(user.name)
				b := fmt.Sprintf(`{"command" : "balance", "result" : %d}`, balance)
				user.sendChan <- b
			case "getDepositAddress":
				address, _ := masc.GetDepositAddress(user.name)
				b := fmt.Sprintf(`{"command" : "depositAddress", "result" : %d}`,
					address)
				user.sendChan <- b
			case "withdraw":
				address := f["address"].(string)
				amount := f["amount"].(int)
				err := masc.Withdraw(user.name, amount, address)
				var b string
				if err != nil {
					b = fmt.Sprintf(`{"command" : "withdraw", "result" : {"error": "%s"}}`, err.Error())
				} else {
					b = `{"command" : "withdraw", "result" : "success"}`
				}
				user.sendChan <- b
			}
		}
		conn.Close()
		log.Println("makeGame: End")
	}
}

func Hub(ready chan *User, hubDone chan chan int) {
	waitingUsers := make([]*User, 0, 2)
	for {
		select {
		case u := <-ready:
			if len(waitingUsers) > 0 {
				log.Println("Matching clients")

				uWaiting := waitingUsers[len(waitingUsers)-1]
				waitingUsers = waitingUsers[:len(waitingUsers)-1]

				select {
				case uWaiting.sendChan <- `{"command": "matched"}`:
				case <-uWaiting.sendErrorChan:
					waitingUsers = append(waitingUsers, u)
				}

				select {
				case u.sendChan <- `{"command": "matched"}`:
				case <-u.sendErrorChan:
					waitingUsers = append(waitingUsers, uWaiting)
				}

				done2 := <-hubDone
				done1 := <-hubDone
				go handleGame(uWaiting, u, done1, done2)

			} else {
				log.Println("Client joined")
				waitingUsers = append(waitingUsers, u)
			}
		}
	}
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
	// send startRound to players
	user1.sendChan <- `{"command" : "startRound"}`
	user2.sendChan <- `{"command" : "startRound"}`

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
			signal := f["signal"].(string)
			log.Println("Received signal from player", player, signal)
			sender <- fmt.Sprintf(`{"command" : "signal", "signal" : %s}`, signal)
		}
		return action
	}

	// wait for clients to send something
ActionLoop:
	for {
		select {
		case msg := <-user1.recvChan:
			action1 = receiveAction(msg, user2.sendChan, chose, 1, action1)
		case msg := <-user2.recvChan:
			action2 = receiveAction(msg, user1.sendChan, chose, 2, action2)
		case <-timer.C:
			log.Println("Time is up")
			user1.sendChan <- fmt.Sprintf(`{ "command": "timerEnd"}`)
			user2.sendChan <- fmt.Sprintf(`{ "command": "timerEnd"}`)
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

	user1.sendChan <- fmt.Sprintf(`{ "command": "endRound", 
						"outcome" : {"me" : %s, "other" : %s}, 
						"balanceDifference" : {"me" : %d, 
											   "other" : %d}}`,
		action1, action2, p1, p2)

	user2.sendChan <- fmt.Sprintf(`{ "command": "endRound", 
						"outcome" : {"me" : %s, "other" : %s}, 
						"balanceDifference" : {"me" : %d, 
											   "other" : %d}}`,
		action2, action1, p2, p1)
	log.Println("Sent payoffs", p1, p2)

	log.Println("Update balances")
	masc.UpdateBalance(user1.name, p1)
	masc.UpdateBalance(user2.name, p2)

	masc.AddAction(user1.name, action1)
	masc.AddAction(user2.name, action2)
	return
}

//TODO: handle more errors
func handleGame(user1, user2 *User, done1, done2 chan int) {
	bet := 1000
	p := 0.2
	E := int(p * 2 * float64(bet))

	cooperate1, defect1, err1 := masc.GetAction(user1.name)
	cooperate2, defect2, err2 := masc.GetAction(user2.name)
	if err1 != nil && err2 != nil {
	} else {
		user1.sendChan <- fmt.Sprintf(`{"command": "stats", 
			"result" : {"cooperate" : %d, "defect" : %d}}`,
			cooperate2, defect2)
		user2.sendChan <- fmt.Sprintf(`{"command": "stats", 
			"result" : {"cooperate" : %d, "defect" : %d}}`,
			cooperate1, defect1)
	}

RoundLoop:
	for {
		log.Println("New game round")

		// check if players have enough funds to play a round
		balance1, _ := masc.GetBalance(user1.name)
		balance2, _ := masc.GetBalance(user2.name)
		log.Println("Players balance: ", balance1, balance2)

		endGame := func() {
			user1.sendChan <- `{"command": "endGame"}`
			user2.sendChan <- `{"command": "endGame"}`
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
		log.Println("r:",r)
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

	done1 <- 1
	done2 <- 1

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
	hubDone := make(chan chan int)
	close := make(chan bool)

	go Hub(ready, hubDone)

	// index.html
	http.HandleFunc("/", staticHandler)
	// static files like js
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	// websocket endpoint
	http.HandleFunc("/play/", makeGame(ready, hubDone, close))
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// receives bitcoin callbacks
	bitcoin.Setup("./bitcoin/blockchain-conf.json")
	unconfirmed := make(chan *bitcoin.RecvTransaction)
	confirmed := make(chan *bitcoin.RecvTransaction)
	http.HandleFunc(fmt.Sprintf("/receive/%s/", bitcoin.Callback_secret),
		bitcoin.ReceiveCallback(unconfirmed, confirmed))
	go masc.InsertIncomingTransactionsInDb(confirmed)

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

func disconnectClient(user *User, conn *websocket.Conn) {
	log.Println("Disconnecting client due to invalid requests.")
	toSend := []byte(`{"result": {"errorMsg": "Disconnecting client due to invalid requests.", "errorCode": ` + strconv.Itoa(10) + `}}`)
	conn.WriteMessage(websocket.TextMessage, toSend)
	conn.Close()
}
