# JSON API

## The game

### Players -> Server

player sends`{"command" : "join"}` if he's ready to start a game
server replies `{"command" : "matched"}` as soon as the game is ready to begin

#### When round is running
player sends `{"command" : "action", "action" : "cooperate"}` or `{"command" : "action", "action" : "defect"}`
Smileys for other player with `{"command" : "signal", "signal" : ?}`

#### After every round server replies
`{ "result": "cooperate" }` or `{ "result": "defect" }`.

### Server -> Players

#### Anytime
Smileys from other player.

#### When round finished
Result of the round.

#### If no round is running
Start of next round.
