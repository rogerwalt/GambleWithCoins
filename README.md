GambleWithCoins
===============

Prisoner's Dilemma
---
A standard prisoner's dilemma looks like this (example from wikipedia):

| -        | C           | D  |
| ------------- |:-------------:| -----:|
| C      | -1, -1 | -3, 0 |
| D      | 0, -3      |   -2, -2 |

In general, the requirement for being a prisoner's dilemma is that `(D,C) > (C,C) > (D,D) > (C,D)` (from the view of player 1).

In order to be fun, the game has to be turned around such that you can win something, therefore, there need to be some positive values in the table. However, the bank can not give out more than it can earn and the game has to be resistant against colluding players.
Therefore, the players have to bet a certain amount `b` before the start of the game. These considerations result in

| -        | C           | D  |
| ------------- |:-------------:| -----:|
| C      | 0, 0 | -b, b |
| D      | b, -b       |   -a, -a |
where `a < b`. The only way to make a profit is playing `D` while tricking your opponent to play `C`.
However, there are some problems with this. First, if you play this game only once there is one strict Nash equilibrium at `(D,D)`, therefore it is strongly recommended to pick `D`.
In contrast, iterated games allow for cooperation equilibria. 
But only if the total number of game rounds is not known. Because if the number of rounds is fixed, you can prove by induction that the the outcome `(D,D)` in every round is the only equilibrium.
Therefore denote with `p` the probability that a game ends after a round. Note that the players have to bet `b` in *each* round.

The second problem is that the players can't win something if they cooperate. Therefore, change the game rules such that the game does not only end with probability `p`, but the bank will also collect the bets this last round. 
Denote with `E` the expected payoff of the bank in each round if both players cooperate. It holds that `E = p*2*b`.
Then we can change the game to 

| -        | C           | D  |
| ------------- |:-------------:| -----:|
| C      | E/2, E/2 | -b, b |
| D      | b, -b       |   0, 0 |

This leads to the following process:
```
each player bets b
with probability p the game ends and the bank collects the bets
both players submit their actions and receive their payoffs
repeat
```

Hawk-Dove Game (aka Game of Chicken)
---
In a hawk-dove game both players fight for a resource. Two doves will share the resource, two hawks will have a costly conflict.

| -        | `H`           | `D`  |
| ------------- |:-------------:| -----:|
| `H`      | `-b, -b` | `E, 0` |
| `D`      | `0, E`   | `E/2, E/2` |
This game becomes especially interesting if you introduce *signaling* (chat). If you can convince your opponent that you are playing hawk in any case, your opponent has no other sensible option than picking dove.

TODO
---
* how to ensure that the player's funds are sufficient for a unknown number of rounds?
* does the bank make enough profit from `(D,D)` and `(H,H)` outcomes in prisoner's dilemma and hawk-dove respectively to pay for server cost?




