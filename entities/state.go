package entities

import (
	"math/rand"
	"online-game/consts"
	"online-game/types"
)

const (
	EmptyTile types.Tile = iota
	TeamATile types.Tile = iota
	TeamBTile types.Tile = iota
	WallTile  types.Tile = iota
)

const (
	WaitingForPlayers types.GamePhase = iota
	Playing           types.GamePhase = iota
	GameOver          types.GamePhase = iota
)

func RandMN(m int, n int) int {
	return m + rand.Intn(n-m)
}

func NewGameState(width, height int) *types.GameState {
	gameMap := types.GameMap{
		Width:  width,
		Height: height,
		Tiles:  make([]types.Tile, height*width),
	}

	// Fill the map with empty tiles
	for i := range gameMap.Tiles {
		gameMap.Tiles[i] = EmptyTile
	}

	// walls
	generateWalls(&gameMap, consts.MAP_DIVISIONS)

	return &types.GameState{
		GameMap: gameMap,
		TeamA:   consts.TEAM_A_COLOR,
		TeamB:   consts.TEAM_B_COLOR,
		Phase:   WaitingForPlayers,
	}
}

func RandomGameState(width, height int) *types.GameState {
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

func Get(m *types.GameMap, x, y int) types.Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return WallTile
	}

	return m.Tiles[y*m.Width+x]
}

func GetAround(m *types.GameMap, x, y int) (tile, bottom, right, bottomRight types.Tile) {
	tile = Get(m, x, y)
	bottom = Get(m, x, y+1)
	right = Get(m, x+1, y)
	bottomRight = Get(m, x+1, y+1)
	return
}

func Set(m *types.GameMap, x, y int, tile types.Tile) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}

	m.Tiles[y*m.Width+x] = tile
}

func Clear(m *types.GameMap) {
	for i := range m.Tiles {
		if m.Tiles[i] != WallTile {
			m.Tiles[i] = EmptyTile
		}
	}
}

func generateWalls(m *types.GameMap, divisions int) {
	generateWallsInRange(m, 0, 0, m.Width-1, m.Height-1, divisions)
}

func generateWallsInRange(m *types.GameMap, x1, y1, x2, y2 int, divisions int) {
	// use BSP to generate walls
	if divisions == 0 {
		return
	}

	width := x2 - x1
	height := y2 - y1

	if width < 2*consts.ROOM_PADDING+1 || height < 2*consts.ROOM_PADDING+1 {
		return
	}

	dir := RandMN(0, 2) // 0: horizontal, 1: vertical
	px := RandMN(x1+consts.ROOM_PADDING, x2-consts.ROOM_PADDING)
	py := RandMN(y1+consts.ROOM_PADDING, y2-consts.ROOM_PADDING)

	if dir == 0 { // horizontal
		// draw a horizontal line
		for x := x1 + consts.ROOM_PADDING; x <= x2-consts.ROOM_PADDING; x++ {
			Set(m, x, py, WallTile)
		}
		// divide the map into two parts
		generateWallsInRange(m, x1, y1, x2, py-1, divisions-1)
		generateWallsInRange(m, x1, py+1, x2, y2, divisions-1)
	} else { // vertical
		// draw a vertical line
		for y := y1 + consts.ROOM_PADDING; y <= y2-consts.ROOM_PADDING; y++ {
			Set(m, px, y, WallTile)
		}
		// divide the map into two parts
		generateWallsInRange(m, x1, y1, px-1, y2, divisions-1)
		generateWallsInRange(m, px+1, y1, x2, y2, divisions-1)
	}
}
