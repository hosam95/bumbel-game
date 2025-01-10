package types

type Tile uint8
type GamePhase uint8

type TeamID uint8

type GameMap struct {
	Width  int
	Height int
	Tiles  []Tile
}

type GameState struct {
	GameMap GameMap
	TeamA   int
	TeamB   int
	ScoreA  int
	ScoreB  int
	Phase   GamePhase
}

type StateMessageState struct {
	TeamA  int32
	TeamB  int32
	ScoreA int32
	ScoreB int32
	Phase  GamePhase
}

type StateMessagePlayer struct {
	Team TeamID
	X    float64
	Y    float64
	VX   int32
	VY   int32
	User StateMessageUser
}

type StateMessageUser struct {
	ID       int16
	Username string
}
