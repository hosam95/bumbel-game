package entities

import "online-game/msgs"

type WeaponId uint8

type Weapon interface {
	Id() WeaponId
	Name() string
	GetCooldown() int
	GetCooldownLeft() float64
	OnWeaponDown(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	OnWeaponUpdate(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	OnWeaponUp(game *Game, player *Player, data map[string]interface{}) (map[string]interface{}, error)
	ParseWeaponDownMessage(message msgs.GenericMessage) (map[string]interface{}, bool)
	ParseWeaponUpMessage(message msgs.GenericMessage) (map[string]interface{}, bool)
	ParseWeaponUpdateMessage(message msgs.GenericMessage) (map[string]interface{}, bool)
	Stringify() map[string]interface{}
}

func CheckWeaponId(id WeaponId, buf []byte) ([]byte, bool) {
	if len(buf) < 1 {
		return buf, false
	}
	return buf[1:], buf[0] == uint8(id)
}

const (
	GrenadeId WeaponId = iota
)
