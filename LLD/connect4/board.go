package connect4

type DiscColor string

const (
	RED    DiscColor = "RED"
	YELLOW DiscColor = "YELLOW"
)

type Board struct {
	rows    int
	columns int
	grid    [][]*DiscColor
}

func NewBoard(rows, cols int) *Board {
	grid := make([][]*DiscColor, rows)

	for i := 0; i < rows; i++ {
		grid[i] = make([]*DiscColor, cols)
	}

	return &Board{
		rows:    rows,
		columns: cols,
		grid:    grid,
	}
}

func (b *Board) GetRows() int {
	return b.rows
}

func (b *Board) GetCols() int {
	return b.columns
}

func (b *Board) CanPlaceDisc(col int, color DiscColor) bool {
	if col < 0 || col >= b.columns {
		return false
	}

	return b.grid[0][col] == nil
}

func (b *Board) PlaceDisc(col int, color DiscColor) int {
	if !b.CanPlaceDisc(col, color) {
		return -1
	}

	for row := b.rows - 1; row >= 0; row-- {
		if b.grid[row][col] == nil {
			b.grid[row][col] = &color
			return row
		}
	}
	return -1
}

func (b *Board) IsFull() bool {
	for col := 0; col < b.columns; col++ {
		if b.grid[b.rows-1][col] == nil {
			return false
		}
	}

	return true
}

func (b *Board) GetCell(row, col int) *DiscColor {
	if row < 0 || row >= b.rows || col < 0 || col > b.columns {
		return nil
	}
	return b.grid[row][col]
}

func (b *Board) CheckWin(row, col int, color DiscColor) bool {
	if row < 0 || row >= b.rows || col < 0 || col > b.columns {
		return false
	}

	cell := b.GetCell(row, col)
	if cell == nil || cell != &color {
		return false
	}

	directions := [][]int{
		{1, 0},
		{0, 1},
		{1, 1},
		{-1, 1},
	}

	for _, dir := range directions {
		count := 1
		count += b.countInDirection(row, col, dir[0], dir[1], color)
		count += b.countInDirection(row, col, -dir[0], -dir[1], color)

		if count >= 4 {
			return true
		}
	}
	return false
}

func (b *Board) countInDirection(row, col, dr, dc int, color DiscColor) int {
	count := 0
	r := row + dr
	c := col + dc
	for r < b.rows {
		for c < b.columns {
			if b.grid[r][c] == &color {
				count++
			} else {
				return count
			}
		}
		r += dr
		c += dc
	}
	return count
}
