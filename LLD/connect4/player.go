package connect4

type Player struct {
	name  string
	color DiscColor
}

func NewPlayer(name string, color DiscColor) *Player {
	return &Player{
		name:  name,
		color: color,
	}
}

func (p *Player) GetName() string {
	return p.name
}

func (p *Player) GetColor() DiscColor {
	return p.color
}
