package types

type Tile uint8
type GamePhase uint8

type TeamID uint8

type GameMap struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"`
}

type GameState struct {
	GameMap GameMap   `json:"map"`
	TeamA   int       `json:"teamA"`
	TeamB   int       `json:"teamB"`
	ScoreA  int       `json:"scoreA"`
	ScoreB  int       `json:"scoreB"`
	Phase   GamePhase `json:"phase"`
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
