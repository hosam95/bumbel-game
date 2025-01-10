package entities

import "math"

type TeamID int

const (
	TeamA TeamID = iota
	TeamB TeamID = iota
)

type Player struct {
	User  *User   `json:"user"`
	Team  TeamID  `json:"team"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	VX    int     `json:"vx"`
	VY    int     `json:"vy"`
	Wepon Wepon
}
type Players []*Player

func (p *Player) Stringify() map[string]interface{} {
	playerMap := map[string]interface{}{
		"user": p.User,
		"team": p.Team,
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

// Move updates the player's velocity based on the direction and start/stop flag
func (p *Player) Move(direction string, start bool) {
	if start {
		switch direction {
		case "up":
			p.VY += -1
		case "down":
			p.VY += 1
		case "left":
			p.VX += -1
		case "right":
			p.VX += 1
		}
	} else {
		switch direction {
		case "up":
			p.VY -= -1
		case "down":
			p.VY -= 1
		case "left":
			p.VX -= -1
		case "right":
			p.VX -= 1
		}
	}
	// Clamp the direction
	p.VX = int(math.Min(math.Max(float64(p.VX), -1), 1))
	p.VY = int(math.Min(math.Max(float64(p.VY), -1), 1))
}

// Update updates the player's position based on the game map
func (p *Player) Update(gameMap *Map) {
	newX := p.X + float64(p.VX)*PlayerSpeed*float64(GameTick.Seconds())
	newY := p.Y + float64(p.VY)*PlayerSpeed*float64(GameTick.Seconds())

	tile, bottom, right, bottomRight := gameMap.GetAround(int(math.Floor(newX)), int(math.Floor(newY)))
	cornerX := newX-math.Floor(newX) > 0
	cornerY := newY-math.Floor(newY) > 0

	// Check for collisions
	if p.VX > 0 { // Moving right
		if right == WallTile || (cornerY && bottomRight == WallTile) {
			newX = math.Floor(newX)
		}
	}

	if p.VX < 0 { // Moving left
		if tile == WallTile || (cornerY && bottom == WallTile) {
			newX = math.Ceil(newX)
		}
	}

	if p.VY > 0 { // Moving down
		if bottom == WallTile || (cornerX && bottomRight == WallTile) {
			newY = math.Floor(newY)
		}
	}

	if p.VY < 0 { // Moving up
		if tile == WallTile || (cornerX && right == WallTile) {
			newY = math.Ceil(newY)
		}
	}

	p.X = newX
	p.Y = newY
}

// Reset resets the player's velocity
func (p *Player) Reset() {
	p.VX = 0
	p.VY = 0
}
