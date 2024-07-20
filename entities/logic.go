package entities

import "math/rand"

type Tile int

const (
	EmptyTile Tile = iota
	TeamATile Tile = iota
	TeamBTile Tile = iota
	WallTile  Tile = iota
)

type GamePhase int

const (
	WaitingForPlayers GamePhase = iota
	Playing           GamePhase = iota
	GameOver          GamePhase = iota
)

const MAX_WALLS = 3
const TEAM_A_COLOR = 0x6C946F
const TEAM_B_COLOR = 0xDC0083

func RandMN(m int, n int) int {
	return m + rand.Intn(n-m)
}

type Map struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}

type GameState struct {
	GameMap Map       `json:"map"`
	TeamA   int       `json:"teamA"`
	TeamB   int       `json:"teamB"`
	Phase   GamePhase `json:"phase"`
}

func NewGameState(width, height int) *GameState {
	gameMap := Map{
		Width:  width,
		Height: height,
		Tiles:  make([]Tile, height*width),
	}

	// Fill the map with empty tiles
	for i := range gameMap.Tiles {
		gameMap.Tiles[i] = EmptyTile
	}

	// walls
	for i := 0; i < MAX_WALLS; i++ {
		x := RandMN(1, width-2)
		w := RandMN(1, width-x)
		y := RandMN(1, height-2)
		h := RandMN(1, height-y)

		for j := y; j < y+h; j++ {
			gameMap.Tiles[j*width+x] = WallTile
		}

		for j := x; j < x+w; j++ {
			gameMap.Tiles[(y+h)*width+j] = WallTile
		}
	}

	return &GameState{
		GameMap: gameMap,
		TeamA:   TEAM_A_COLOR,
		TeamB:   TEAM_B_COLOR,
		Phase:   WaitingForPlayers,
	}
}

func RandomGameState(width, height int) *GameState {
	gameState := NewGameState(width, height)

	// Fill the map with random team tiles
	for i, tile := range gameState.GameMap.Tiles {
		if tile != WallTile && rand.Intn(100) < 50 {
			if rand.Intn(2) == 0 {
				gameState.GameMap.Tiles[i] = TeamATile
			} else {
				gameState.GameMap.Tiles[i] = TeamBTile
			}
		}
	}

	return gameState
}

func (gs *GameState) Stringify() map[string]interface{} {
	return map[string]interface{}{
		"map":   gs.GameMap,
		"teamA": gs.TeamA,
		"teamB": gs.TeamB,
		"phase": gs.Phase,
	}
}
