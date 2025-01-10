package entities

import (
	"online-game/msgs"
	"online-game/types"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type User struct {
	ID       int16
	Username string
	C        *websocket.Conn
	mu       sync.Mutex
}

var Users = map[int16]*User{}

func NewUser(c *websocket.Conn, id int16, username string) *User {
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
	return u.C.WriteMessage(websocket.BinaryMessage, msg)
}

// TODO: make this accept ServerMessage interface
func (u *User) SendMessage(msg msgs.ServerMessage) error {
	buf, ok := msg.Buffer()
	if !ok {
		return nil
	}
	return u.C.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}

func (u *User) Error(message string) {
	em := msgs.ErrorMessage{Message: message}
	u.SendMessage(em)
}

func (u *User) ToPlayer(team types.TeamID) *Player {
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
