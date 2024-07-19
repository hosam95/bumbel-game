package entities

import (
	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	C        *websocket.Conn `json:"-"`
}

var Users = map[string]*User{}

func NewUser(c *websocket.Conn, id, username string) *User {
	user := &User{
		ID:       id,
		Username: username,
		C:        c,
	}
	Users[id] = user
	return user
}

// func (u *User) Stringify() string {
// 	return
// }

func (u *User) Join(roomId string) {
	// Join the room
}

func (u *User) ToPlayer() *Player {
	return &Player{
		User: u,
		// Location: loc,
	}
}

func (u *User) Cleanup() {
	game := FindUserInfo(u.ID)
	game.RemovePlayer(u.ID)
}
