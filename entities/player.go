package entities

import (
	"math"
	"online-game/types"
)

const (
	TeamA types.TeamID = iota
	TeamB types.TeamID = iota
)

type Player struct {
	User   *User
	Team   types.TeamID
	X      float64
	Y      float64
	VX     int
	VY     int
	Weapon Weapon
}
type Players []*Player

// TODO: name
func (p Players) Foo() []types.StateMessagePlayer {
	smps := make([]types.StateMessagePlayer, len(p))
	for i, player := range p {
		smps[i] = player.ToStateMessagePlayer()
	}
	return smps
}

func (p *Player) ToStateMessagePlayer() types.StateMessagePlayer {
	return types.StateMessagePlayer{
		Team: p.Team,
		X:    p.X,
		Y:    p.Y,
		VX:   int32(p.VX),
		VY:   int32(p.VY),
		User: types.StateMessageUser{
			ID:       p.User.ID,
			Username: p.User.Username,
		},
	}
}

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

func (p *Player) Update(gameMap *types.GameMap) {
	newX := p.X + float64(p.VX)*PlayerSpeed*float64(GameTick.Seconds())
	newY := p.Y + float64(p.VY)*PlayerSpeed*float64(GameTick.Seconds())

	tile, bottom, right, bottomRight := GetAround(gameMap, int(math.Floor(newX)), int(math.Floor(newY)))
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
