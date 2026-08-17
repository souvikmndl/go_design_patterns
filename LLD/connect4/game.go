package connect4

type GameState string

const (
	InProgress GameState = "IN_PROGRESS"
	Draw       GameState = "DRAW"
	Won        GameState = "Won"
)

type Game struct {
	player1       *Player
	player2       *Player
	board         *Board
	currentPlayer *Player
	state         GameState
	winner        *Player
}

func NewGame(pl1, pl2 *Player) *Game {
	return &Game{
		player1:       pl1,
		player2:       pl2,
		board:         NewBoard(6, 7),
		currentPlayer: pl1,
		state:         InProgress,
	}
}

func (g *Game) MakeMove(player *Player, col int) bool {
	if g.state != InProgress {
		return false
	}

	if player == nil || player != g.currentPlayer {
		return false
	}

	if !g.board.CanPlaceDisc(col, player.GetColor()) {
		return false
	}

	row := g.board.PlaceDisc(col, player.GetColor())
	if row == -1 {
		return false
	}

	if g.board.CheckWin(row, col, player.GetColor()) {
		g.state = Won
		g.winner = player
	} else if g.board.IsFull() {
		g.state = Draw
	} else {
		if g.currentPlayer == g.player1 {
			g.currentPlayer = g.player2
		} else {
			g.currentPlayer = g.player1
		}
	}
	return true
}

func (g *Game) GetGameState() GameState {
	return g.state
}

func (g *Game) GetCurrentPlayer() *Player {
	return g.currentPlayer
}

func (g *Game) GetWinner() *Player {
	return g.winner
}

func (g *Game) GetBoard() *Board {
	return g.board
}
