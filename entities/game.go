package entities

import (
	"encoding/json"
	"errors"
	"math/rand"
	"online-game/structs"
	"time"
)

const TickRate = 30
const GameTick = time.Millisecond * 1000 / TickRate
const MapTick = TickRate * 5 // every 5 seconds

const MaxPlayers = 8
const GameDuration = 60 * time.Second
const PlayerSpeed = 10

const MapWidth = 48
const MapHeight = 27

type Game struct {
	Players Players   `json:"players"`
	State   GameState `json:"state"`
	Host    string    `json:"host"`
	Room    string    `json:"room"`
	LC      bool      `json:"-"` // large change

	Started   bool      `json:"started"`
	StartedAt time.Time `json:"startedAt"`
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
		Host:  host.ID,
		Room:  room,
		LC:    true,
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
		"players":   g.Players.Stringify(),
		"state":     g.State.Stringify(),
		"host":      g.Host,
		"room":      g.Room,
		"started":   g.Started,
		"startedAt": g.StartedAt,
	}

	return gameMap
}

func (g *Game) AddUser(user *User) error {
	if len(g.Players) >= MaxPlayers {
		return errors.New("game is full")
	}

	if g.State.Phase != WaitingForPlayers {
		return errors.New("game has already started")
	}
	var newTeam TeamID
	if len(g.Players)&1 == 0 { // even number of players
		newTeam = TeamA
	} else {
		newTeam = TeamB
	}

	player := user.ToPlayer(newTeam)
	g.Players = append(g.Players, player)
	g.LC = true

	return nil
}

func (g *Game) RemovePlayer(userId string) {
	for i, p := range g.Players {
		if p.User.ID == userId {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			break
		}
	}
	if len(g.Players) == 0 {
		g.Terminate()
	} else if len(g.Players) < 2 {
		g.State.Phase = WaitingForPlayers
		g.State.GameMap.Clear()
		if g.Host == userId {
			g.Host = g.Players[0].User.ID
		}
	}
	g.LC = true
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
	g.LC = true

	return nil
}

func (g *Game) Start(userId string) error {
	if g.State.Phase == Playing {
		return errors.New("game has already started")
	}

	if g.Host != userId {
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

	g.State = *NewGameState(MapWidth, MapHeight)
	g.BroadcastTD("map", g.State.GameMap.Serialize())

	g.State.Phase = Playing
	g.Started = true
	g.StartedAt = time.Now()

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

	x := int(player.X + 0.5)
	y := int(player.Y + 0.5)

	curr := g.State.GameMap.Get(x, y)
	if curr == WallTile {
		return CellResult{}, errors.New("cannot paint wall")
	}
	switch curr {
	case TeamATile:
		g.State.ScoreA--
	case TeamBTile:
		g.State.ScoreB--
	}

	var newTile Tile
	switch player.Team {
	case TeamA:
		newTile = TeamATile
		g.State.ScoreA++
	case TeamB:
		newTile = TeamBTile
		g.State.ScoreB++
	}
	g.State.GameMap.Set(x, y, newTile)

	return CellResult{
		X:     x,
		Y:     y,
		State: newTile,
	}, nil
}

func (g *Game) Update() {
	if g.State.Phase != Playing {
		return
	}

	gameMap := g.State.GameMap
	for _, player := range g.Players {
		player.Update(&gameMap)
	}
	if time.Since(g.StartedAt) > GameDuration {
		g.Finish()
	}
}

func (g *Game) Finish() {
	g.State.Phase = GameOver
	g.Started = false
	for _, player := range g.Players {
		player.Reset()
	}
	g.BroadcastSystem("info", "Game over")
	g.LC = true
}

func (g *Game) Terminate() {
	for i, game := range Games {
		if game == g {
			Games = append(Games[:i], Games[i+1:]...)
			break
		}
	}
}

func (g *Game) Broadcast(message structs.Message, exclude ...string) {
	jsonMsg, _ := json.Marshal(message)

PlayerLoop:
	for _, player := range g.Players {
		for _, ex := range exclude {
			if player.User.ID == ex {
				continue PlayerLoop
			}
		}
		player.User.Send(jsonMsg)
	}
}

func (g *Game) BroadcastTD(t string, d map[string]interface{}, exclude ...string) {
	mapped := structs.Message{
		Type: t,
		Data: d,
	}
	g.Broadcast(mapped, exclude...)
}

func (g *Game) BroadcastState(exclude ...string) {
	g.BroadcastTD("state", g.Stringify(), exclude...)
}

func (g *Game) BroadcastSystem(msgType, msg string, exclude ...string) {
	mapped := structs.Message{
		Type: "system",
		Data: map[string]interface{}{"msg": msg, "type": msgType},
	}
	jsonMsg, _ := json.Marshal(mapped)

PlayerLoop:
	for _, player := range g.Players {
		for _, ex := range exclude {
			if player.User.ID == ex {
				continue PlayerLoop
			}
		}
		player.User.Send(jsonMsg)
	}
}
