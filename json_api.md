# JSON API

### login
player has to be logged in before he can join a game.


`{"command" : "register", "name" : ?, "password" : ?}`
`{"command" : "login", "name" : ?, "password" : ?}`

returns either

`{"command" : "register", "result" : "success"}`

`{"command" : "login", "result" : "success"}`

or `{"error" : errormessage}`

### general
`{"command" : "balance", "balance" : ?}` is sent from the server whenever a balance update occurs

### setup game
player sends`{"command" : "join"}` if he's ready to start a game,
server replies `{"command" : "matched"}` as soon as the game is ready to begin

#### When round is running
player sends `{"command" : "action", "action" : "cooperate"}` or `{"command" : "action", "action" : "defect"}`
Smileys for other player with `{"command" : "signal", "signal" : ?}`

#### After every round server replies
`{ "command": "outcome", ... }`

