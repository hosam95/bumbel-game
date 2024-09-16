package entities

import (
	"encoding/json"
	"online-game/structs"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	C        *websocket.Conn `json:"-"`
	mu       sync.Mutex
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

func (u *User) Send(msg []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.C.WriteMessage(websocket.TextMessage, msg)
}

func (u *User) SendMessage(t string, data map[string]any) {
	msg := structs.Message{
		Type: t,
		Data: data,
	}
	jsonMsg, _ := json.Marshal(msg)
	u.Send(jsonMsg)
}

func (u *User) Error(message string) {
	u.SendMessage("error", map[string]any{"message": message})
}

func (u *User) ToPlayer(team TeamID) *Player {
	return &Player{
		User: u,
		Team: team,
	}
}

func (u *User) Cleanup() {
	game := FindUserInfo(u.ID)
	if game != nil {
		game.RemovePlayer(u.ID)
	}
	delete(Users, u.ID)
}
