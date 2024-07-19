package entities

type Player struct {
	User *User `json:"user"`
	// Location string `json:"location"`
}
type Players []*Player

func (p *Player) Stringify() map[string]interface{} {
	playerMap := map[string]interface{}{
		"user": p.User,
		// "location": p.Location,
	}

	return playerMap
}

func (p Players) Stringify() []map[string]interface{} {
	pArr := make([]map[string]interface{}, len(p))

	for i, player := range p {
		pArr[i] = player.Stringify()
	}

	return pArr
}
