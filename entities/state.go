package entities

import (
	"math/rand"
)

type Tile int
type GamePhase int

type Map struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}

type GameState struct {
	GameMap Map       `json:"map"`
	TeamA   int       `json:"teamA"`
	TeamB   int       `json:"teamB"`
	ScoreA  int       `json:"scoreA"`
	ScoreB  int       `json:"scoreB"`
	Phase   GamePhase `json:"phase"`
}

const (
	EmptyTile Tile = iota
	TeamATile Tile = iota
	TeamBTile Tile = iota
	WallTile  Tile = iota
)

const (
	WaitingForPlayers GamePhase = iota
	Playing           GamePhase = iota
	GameOver          GamePhase = iota
)

const (
	MAP_DIVISIONS = 6
	ROOM_PADDING  = 2
	TEAM_A_COLOR  = 0x6C946F
	TEAM_B_COLOR  = 0xDC0083
)

func RandMN(m int, n int) int {
	return m + rand.Intn(n-m)
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
	gameMap.generateWalls(MAP_DIVISIONS)

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
		"teamA":  gs.TeamA,
		"teamB":  gs.TeamB,
		"scoreA": gs.ScoreA,
		"scoreB": gs.ScoreB,
		"phase":  gs.Phase,
	}
}

func (m *Map) Serialize() map[string]interface{} {
	return map[string]interface{}{
		"width":  m.Width,
		"height": m.Height,
		"tiles":  m.Tiles,
	}
}

func (m *Map) Get(x, y int) Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return WallTile
	}

	return m.Tiles[y*m.Width+x]
}

func (m *Map) GetAround(x, y int) (tile Tile, bottom Tile, right Tile, bottomRight Tile) {
	tile = m.Get(x, y)
	bottom = m.Get(x, y+1)
	right = m.Get(x+1, y)
	bottomRight = m.Get(x+1, y+1)
	return
}

func (m *Map) Set(x, y int, tile Tile) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}

	m.Tiles[y*m.Width+x] = tile
}

func (m *Map) Clear() {
	for i := range m.Tiles {
		if m.Tiles[i] != WallTile {
			m.Tiles[i] = EmptyTile
		}
	}
}

func (m *Map) generateWalls(divisions int) {
	m.generateWallsInRange(0, 0, m.Width-1, m.Height-1, divisions)
}

func (m *Map) generateWallsInRange(x1, y1, x2, y2 int, divisions int) {
	// use BSP to generate walls
	if divisions == 0 {
		return
	}

	width := x2 - x1
	height := y2 - y1

	if width < 2*ROOM_PADDING+1 || height < 2*ROOM_PADDING+1 {
		return
	}

	dir := RandMN(0, 2) // 0: horizontal, 1: vertical
	px := RandMN(x1+ROOM_PADDING, x2-ROOM_PADDING)
	py := RandMN(y1+ROOM_PADDING, y2-ROOM_PADDING)

	if dir == 0 { // horizontal
		// draw a horizontal line
		for x := x1 + ROOM_PADDING; x <= x2-ROOM_PADDING; x++ {
			m.Set(x, py, WallTile)
		}
		// divide the map into two parts
		m.generateWallsInRange(x1, y1, x2, py-1, divisions-1)
		m.generateWallsInRange(x1, py+1, x2, y2, divisions-1)
	} else { // vertical
		// draw a vertical line
		for y := y1 + ROOM_PADDING; y <= y2-ROOM_PADDING; y++ {
			m.Set(px, y, WallTile)
		}
		// divide the map into two parts
		m.generateWallsInRange(x1, y1, px-1, y2, divisions-1)
		m.generateWallsInRange(px+1, y1, x2, y2, divisions-1)
	}
}
