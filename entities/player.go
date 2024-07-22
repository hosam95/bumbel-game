package entities

type Player struct {
	User *User   `json:"user"`
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
	VX   int     `json:"vx"`
	VY   int     `json:"vy"`
}
type Players []*Player

func (p *Player) Stringify() map[string]interface{} {
	playerMap := map[string]interface{}{
		"user": p.User,
		"x":    p.X,
		"y":    p.Y,
		"vx":   p.VX,
		"vy":   p.VY,
	}

	return playerMap
}

func (p Players) Stringify() []map[string]interface{} {
	pArr := make([]map[string]interface{}, len(p))

	for i, player := range p {
		pArr[i] = player.Stringify()
	}

	return pArr
}

func (p *Player) Move(direction string, start bool) {
	if start {
		switch direction {
		case "up":
			p.VY = -1
		case "down":
			p.VY = 1
		case "left":
			p.VX = -1
		case "right":
			p.VX = 1
		}
	} else {
		switch direction {
		case "up":
			p.VY = 0
		case "down":
			p.VY = 0
		case "left":
			p.VX = 0
		case "right":
			p.VX = 0
		}
	}
}
