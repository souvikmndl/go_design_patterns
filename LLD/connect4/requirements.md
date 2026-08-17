# Connect 4 — Design Requirements

## Functional Requirements

1. Two players take turns dropping discs into a 7-column, 6-row board.
2. A disc falls to the lowest available row in the chosen column.
3. The game ends when:
   - A player gets four discs in a row (vertical, horizontal, or diagonal) — **they win**.
   - The board is full — **it's a draw**.
4. Invalid moves are rejected clearly:
   - Dropping in a full column.
   - Moving out of turn.
   - Moving after the game is over.

## Out of Scope

- UI support
- Concurrent games
- Move history
- Undo
- Board size configuration

## Core Entities and Relationships

| Entity | Responsibility |
|--------|-----------------|
| **Game** | Orchestrator. Holds `Board` and `Player`s. Tracks whose turn it is, manages game status, and enforces rules. |
| **Board** | The 7x6 grid where discs live. Owns grid state and handles disc placement — checking if a column is full, which row a disc falls to, and whether 4 discs are connected. |
| **Player** | Represents a person in the game — holds the player's name and disc color. |

## Class Design

### Enums

**GameState**
- `IN_PROGRESS`
- `WON`
- `DRAW`

**DiscColor**
- `RED`
- `YELLOW`

### Game

```
class Game:
    - board: Board
    - player1: Player
    - player2: Player
    - currentPlayer: Player
    - state: GameState
    - winner: Player?

    + Game(player1, player2)
    + makeMove(player, col) -> bool
    + getGameState() -> GameState
    + getCurrentPlayer() -> Player
    + getWinner() -> Player
    + getBoard() -> Board
```

### Board

```
class Board:
    - rows: int
    - cols: int
    - grid: DiscColor[rows][cols]

    + placeDisc(column, color) -> row
    + canPlaceDisc(column) -> bool
    + isFull() -> bool
    + checkWin(row, column, color) -> bool
    + getCell(row, column) -> DiscColor
```

### Player

```
class Player:
    - name: string
    - color: DiscColor
```
