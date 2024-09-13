package entities

import (
	"errors"
	"math/rand"
	"time"
)

const TickRate = 30
const GameTick = time.Millisecond * 1000 / TickRate

const PlayerSpeed = 10

const MapWidth = 40
const MapHeight = 30

type Game struct {
	Players Players   `json:"players"`
	State   GameState `json:"state"`
	Host    *Player   `json:"host"`
	Room    string    `json:"room"`
}

var Games = []*Game{}

type CellResult struct {
	X     int  `json:"x"`
	Y     int  `json:"y"`
	State Tile `json:"state"`
}

func NewGame(host *User) string {
	room := "" // a random string
	for i := 0; i < 4; i++ {
		room += string(rune(65 + rand.Intn(26)))
	}
	player := host.ToPlayer(TeamA)
	game := &Game{
		Players: Players{
			player,
		},
		State: *NewGameState(MapWidth, MapHeight),
		Host:  player,
		Room:  room,
	}
	Games = append(Games, game)
	return room
}

func FindGameByRoom(room string) *Game {
	for _, game := range Games {
		if game.Room == room {
			return game
		}
	}
	return nil
}

func FindUserInfo(userId string) *Game {
	for _, game := range Games {
		for _, player := range game.Players {
			if player.User.ID == userId {
				return game
			}
		}
	}
	return nil
}

func (g *Game) Stringify() map[string]interface{} {
	gameMap := map[string]interface{}{
		"players": g.Players.Stringify(),
		"state":   g.State.Stringify(),
		"host":    g.Host.User.ID,
		"room":    g.Room,
	}

	return gameMap
}

func (g *Game) AddUser(user *User) {
	var newTeam TeamID
	if len(g.Players)&1 == 0 {
		newTeam = TeamB
	} else {
		newTeam = TeamA
	}

	player := user.ToPlayer(newTeam)
	g.Players = append(g.Players, player)
}

func (g *Game) RemovePlayer(userId string) {
	for i, p := range g.Players {
		if p.User.ID == userId {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			break
		}
	}
	if len(g.Players) == 0 {
		g.Finish()
	} else if len(g.Players) < 2 {
		g.State.Phase = WaitingForPlayers
		if g.Host.User.ID == userId {
			g.Host = g.Players[0]
		}
	}
}

func (g *Game) GetPlayer(userId string) *Player {
	for _, player := range g.Players {
		if player.User.ID == userId {
			return player
		}
	}
	return nil
}

func (g *Game) SwitchTeams(userId string) error {
	if g.State.Phase != WaitingForPlayers {
		return errors.New("game has already started")
	}

	player := g.GetPlayer(userId)
	if player == nil {
		return errors.New("player not found")
	}

	if player.Team == TeamA {
		player.Team = TeamB
	} else {
		player.Team = TeamA
	}

	return nil
}

func (g *Game) Start(userId string) error {
	if g.State.Phase != WaitingForPlayers {
		return errors.New("game has already started")
	}

	if g.Host.User.ID != userId {
		return errors.New("only the host can start the game")
	}

	if len(g.Players) < 2 {
		return errors.New("need at least 2 players to start the game")
	}

	teamA := 0
	teamB := 0

	for _, player := range g.Players {
		if player.Team == TeamA {
			teamA++
		} else {
			teamB++
		}
	}

	if teamA == 0 || teamB == 0 {
		return errors.New("need at least one player on each team")
	}

	g.State.Phase = Playing
	for _, player := range g.Players {
		for {
			player.X = rand.Float64() * float64(g.State.GameMap.Width)
			player.Y = rand.Float64() * float64(g.State.GameMap.Height)
			if g.State.GameMap.Get(int(player.X), int(player.Y)) != WallTile {
				break
			}
		}
	}
	return nil
}

func (g *Game) MovePlayer(userId string, direction string, start bool) {
	player := g.GetPlayer(userId)
	if player != nil {
		player.Move(direction, start)
	}
}

func (g *Game) Shoot(userId string) (CellResult, error) {
	player := g.GetPlayer(userId)
	if player == nil {
		return CellResult{}, errors.New("player not found")
	}

	x := int(player.X)
	y := int(player.Y)

	curr := g.State.GameMap.Get(x, y)
	if curr == WallTile {
		return CellResult{}, errors.New("cannot paint wall")
	}
	var newTile Tile
	switch player.Team {
	case TeamA:
		newTile = TeamATile
	case TeamB:
		newTile = TeamBTile
	}
	g.State.GameMap.Set(x, y, newTile)

	return CellResult{
		X:     x,
		Y:     y,
		State: newTile,
	}, nil
}

func (g *Game) Update() {
	gameMap := g.State.GameMap
	for _, player := range g.Players {
		player.Update(&gameMap)
	}
}

func (g *Game) Finish() {
	for i, game := range Games {
		if game == g {
			Games = append(Games[:i], Games[i+1:]...)
			break
		}
	}
}
