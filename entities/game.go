package entities

import (
	"errors"
	"fmt"
	"math/rand"
	"online-game/msgs"
	"online-game/types"
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

// Game represents a game session
type Game struct {
	Players Players
	State   types.GameState
	Host    int16
	Room    string
	LC      bool // large change

	Started   bool
	StartedAt time.Time
}

var Games = []*Game{}

type CellResult struct {
	X     int
	Y     int
	State types.Tile
}

// NewGame creates a new game and returns the room code
func NewGame(host *User, wepon *Weapon) string {
	room := "" // a random string
	for i := 0; i < 4; i++ {
		room += string(rune(65 + rand.Intn(26)))
	}
	player := host.ToPlayer(TeamA, wepon)
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

// FindGameByRoom finds a game by its room code
func FindGameByRoom(room string) *Game {
	for _, game := range Games {
		if game.Room == room {
			return game
		}
	}
	return nil
}

// FindUserInfo finds the game a user is in
func FindUserInfo(userId int16) *Game {
	for _, game := range Games {
		for _, player := range game.Players {
			if player.User.ID == userId {
				return game
			}
		}
	}
	return nil
}

func (g *Game) AddUser(user *User, weapon *Weapon) error {
	if len(g.Players) >= MaxPlayers {
		return errors.New("game is full")
	}

	if g.State.Phase != WaitingForPlayers {
		return errors.New("game has already started")
	}
	var newTeam types.TeamID
	if len(g.Players)&1 == 0 { // even number of players
		newTeam = TeamA
	} else {
		newTeam = TeamB
	}

	player := user.ToPlayer(newTeam, weapon)
	g.Players = append(g.Players, player)
	g.LC = true

	return nil
}

func (g *Game) RemovePlayer(userId int16) {
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
		Clear(&g.State.GameMap)
		if g.Host == userId {
			g.Host = g.Players[0].User.ID
		}
	}
	g.LC = true
}

func (g *Game) GetPlayer(userId int16) *Player {
	for _, player := range g.Players {
		if player.User.ID == userId {
			return player
		}
	}
	return nil
}

func (g *Game) SwitchTeams(userId int16) error {
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

func (g *Game) Start(userId int16) error {
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

	if g.State.Phase == WaitingForPlayers { // First game
		Clear(&g.State.GameMap)
	} else {
		g.State = *NewGameState(MapWidth, MapHeight)
	}

	g.BroadcastMap()

	g.State.Phase = Playing
	g.Started = true
	g.StartedAt = time.Now()

	for _, player := range g.Players {
		for {
			player.X = rand.Float64() * float64(g.State.GameMap.Width)
			player.Y = rand.Float64() * float64(g.State.GameMap.Height)
			if Get(&g.State.GameMap, int(player.X), int(player.Y)) != WallTile {
				break
			}
		}
	}

	return nil
}

func (g *Game) MovePlayer(userId int16, direction string, start bool) {
	player := g.GetPlayer(userId)
	if player != nil {
		player.Move(direction, start)
	}
}

func (g *Game) Shoot(userId int16) (CellResult, error) {
	player := g.GetPlayer(userId)
	if player == nil {
		return CellResult{}, errors.New("player not found")
	}

	x := int(player.X + 0.5)
	y := int(player.Y + 0.5)

	curr := Get(&g.State.GameMap, x, y)
	if curr == WallTile {
		return CellResult{}, errors.New("cannot paint wall")
	}
	switch curr {
	case TeamATile:
		g.State.ScoreA--
	case TeamBTile:
		g.State.ScoreB--
	}

	var newTile types.Tile
	switch player.Team {
	case TeamA:
		newTile = TeamATile
		g.State.ScoreA++
	case TeamB:
		newTile = TeamBTile
		g.State.ScoreB++
	}
	Set(&g.State.GameMap, x, y, newTile)

	return CellResult{
		X:     x,
		Y:     y,
		State: newTile,
	}, nil
}

// Update updates the game state
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

// Finish finishes the game
func (g *Game) Finish() {
	g.State.Phase = GameOver
	g.Started = false
	for _, player := range g.Players {
		player.Reset()
	}
	g.BroadcastSystem(msgs.SYS_MSG_INFO, "Game over")
	g.LC = true
}

// Terminate terminates the game
func (g *Game) Terminate() {
	for i, game := range Games {
		if game == g {
			Games = append(Games[:i], Games[i+1:]...)
			break
		}
	}
}

func (g *Game) Broadcast(message msgs.ServerMessage, exclude ...int16) {
	b, _ := message.Buffer()
	buf := b.Bytes()

PlayerLoop:
	for _, player := range g.Players {
		for _, ex := range exclude {
			if player.User.ID == ex {
				continue PlayerLoop
			}
		}
		player.User.Send(buf)
	}
}

func (g *Game) BroadcastMap(exclude ...int16) {
	g.Broadcast(msgs.MapMessage{
		Map: g.State.GameMap,
	})
}

func (g *Game) BroadcastState(exclude ...int16) {
	// fmt.Printf("Started At: %v\r\nUnix: %d\r\n int32: %d\r\n", g.StartedAt, int32(g.StartedAt.Unix()), int32(g.StartedAt.Unix()))
	g.Broadcast(msgs.StateMessage{
		Host:      g.Host,
		Room:      g.Room,
		Started:   g.Started,
		StartedAt: int32(g.StartedAt.Unix()),
		State: types.StateMessageState{
			TeamA:  int32(g.State.TeamA),
			TeamB:  int32(g.State.TeamB),
			ScoreA: int32(g.State.ScoreA),
			ScoreB: int32(g.State.ScoreB),
			Phase:  g.State.Phase,
		},
		Players: g.Players.Foo(),
	})
}

func (g *Game) BroadcastSystem(msgType uint8, msg string, exclude ...int16) {
	mapped := msgs.SystemMessage{
		Type:    msgs.SYS_MSG_INFO,
		Message: msg,
	}
	b, ok := mapped.Buffer()

	if !ok {
		fmt.Println("Failed to marshal system message")
	}
	buf := b.Bytes()

PlayerLoop:
	for _, player := range g.Players {
		for _, ex := range exclude {
			if player.User.ID == ex {
				continue PlayerLoop
			}
		}
		player.User.Send(buf)
	}
}
