package entities

type WeponId int

type Wepon interface {
	Id() WeponId
	Name() string
	GetCooldown() int
	GetCooldownLeft() float64
	OnPress(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	Update(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	OnRelease(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	Stringify() map[string]interface{}
}
