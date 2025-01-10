package entities

import (
	"encoding/json"
	"online-game/structs"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// User represents a connected player
type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	C        *websocket.Conn `json:"-"`
	mu       sync.Mutex
}

var Users = map[string]*User{}

// NewUser creates a new user and adds it to the global map
func NewUser(c *websocket.Conn, id, username string) *User {
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
	return u.C.WriteMessage(websocket.TextMessage, msg)
}

// SendMessage sends a structured message to the user
func (u *User) SendMessage(t string, data map[string]any) {
	msg := structs.Message{
		Type: t,
		Data: data,
	}
	jsonMsg, _ := json.Marshal(msg)
	u.Send(jsonMsg)
}

// Error sends an error message to the user
func (u *User) Error(message string) {
	u.SendMessage("error", map[string]any{"message": message})
}

// ToPlayer converts the user to a player for game participation
func (u *User) ToPlayer(team TeamID, powerup *Wepon) *Player {
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
