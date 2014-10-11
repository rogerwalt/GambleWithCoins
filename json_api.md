# JSON API

## Authentification
### login and register
player has to be logged in before he can join a game.


`{"command" : "register", "name" : ?, "password" : ?}`
`{"command" : "login", "name" : ?, "password" : ?}`

returns either

`{"command" : "register", "result" : "success"}`

`{"command" : "login", "result" : "success"}`

or `{"command" : "login", "result" {"error" : error_message}}`

## general
### balance
`{"command" : "getBalance"}` is sent from the client to server to retrieve the current balance
`{"command" : "balance", "result" : ?}` is sent from the server whenever requested or a balance update occurs

`{"command" : "getDepositAddress"}` client -> server
`{"command" : "depositAddress", "result" : ?}` server -> client

`{"command" : "withdraw", "address" : ?}` client -> server
`{"command" : "withdraw", "result" : ?}` server -> client


### setup game
player sends`{"command" : "join"}` if he's ready to start a game,
server replies `{"command" : "matched"}` as soon as the game is ready to begin

#### When round is running
player sends `{"command" : "action", "action" : "cooperate"}` or `{"command" : "action", "action" : "defect"}`
Smileys for other player with `{"command" : "signal", "signal" : ?}`

#### After every round server replies
`{ "command": "outcome", ... }`

