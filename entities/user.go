package entities

import (
	"online-game/msgs"
	"online-game/types"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// User represents a connected player
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

// Send sends a message to the user
func (u *User) Send(msg []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.C.WriteMessage(websocket.BinaryMessage, msg)
}

func (u *User) SendMessage(msg msgs.ServerMessage) error {
	buf, ok := msg.Buffer()
	if !ok {
		return nil
	}
	return u.C.WriteMessage(websocket.BinaryMessage, buf.Bytes())
}

// Error sends an error message to the user
func (u *User) Error(message string) {
	em := msgs.ErrorMessage{Message: message}
	u.SendMessage(em)
}

func (u *User) ToPlayer(team types.TeamID, powerup *Wepon) *Player {
	return &Player{
		User:  u,
		Team:  team,
		Wepon: *powerup,
	}
}

// Cleanup removes the user from the game and global map
func (u *User) Cleanup() {
	game := FindUserInfo(u.ID)
	if game != nil {
		game.RemovePlayer(u.ID)
	}
	delete(Users, u.ID)
}
