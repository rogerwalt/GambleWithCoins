# JSON API

## The game

### Players -> Server

#### When round is running
Smileys for other player.

#### Once every round
`{ "result": "cooperate" }` or `{ "result": "defect" }`.

### Server -> Players

#### Anytime
Smileys from other player.

#### When round finished
Result of the round.

#### If no round is running
Start of next round.