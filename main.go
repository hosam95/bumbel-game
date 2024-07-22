package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"online-game/entities"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

const TickRate = 30
const GameTick = time.Millisecond * 1000 / TickRate

const PlayerSpeed = 10

var connections = map[string]*entities.User{}

type Message struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func ParseMessage(msg []byte) (*Message, error) {
	message := &Message{}
	err := json.Unmarshal(msg, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func randomName() string {
	first := []string{"The"}
	second := []string{"Fiery", "Icy", "Electric", "Magnetic", "Toxic", "Radioactive", "Mystic", "Dark", "Light", "Wind", "Water", "Earth", "Fire"}
	third := []string{"Chickens", "Ducks", "Geese", "Pigeons", "Eagles", "Falcons", "Hawks", "Owls", "Parrots", "Penguins", "Robins", "Sparrows", "Swans", "Turkeys"}

	return fmt.Sprintf("%s %s %s", first[rand.Intn(len(first))], second[rand.Intn(len(second))], third[rand.Intn(len(third))])
}

func UpdateState() {
	for _, game := range entities.Games {
		for _, player := range game.Players {
			player.X += float32(player.VX) * PlayerSpeed * float32(GameTick.Seconds())
			player.Y += float32(player.VY) * PlayerSpeed * float32(GameTick.Seconds())
		}
	}
}

func BroadcastState() {
	for _, game := range entities.Games {
		state := game.Stringify()
		msg := Message{
			Type: "state",
			Data: state,
		}
		jsonState, _ := json.Marshal(msg)
		for _, player := range game.Players {
			conn := connections[player.User.ID]
			if conn == nil {
				continue
			}
			player.User.C.WriteMessage(websocket.TextMessage, jsonState)
		}
	}
}

func SendMessage(t string, data map[string]any, to ...*websocket.Conn) {
	msg := Message{
		Type: t,
		Data: data,
	}
	jsonMsg, _ := json.Marshal(msg)
	for _, conn := range to {
		conn.WriteMessage(websocket.TextMessage, jsonMsg)
	}
}

func main() {
	app := fiber.New()

	app.Use(filesystem.New(filesystem.Config{
		Root: http.Dir("./public"),
	}))

	go func() {
		// broadcast game state every GameTick
		for {
			UpdateState()
			BroadcastState()
			time.Sleep(GameTick)
		}
	}()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		id := fmt.Sprintf("%X", rand.Int63())
		user := entities.NewUser(c, id, randomName())
		connections[id] = user
		SendMessage("connected", map[string]any{"id": id, "username": connections[id].Username}, c)

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		for {
			message := Message{}
			if err := c.ReadJSON(&message); err != nil {
				log.Println("read:", err)
				break
			}

			game := entities.FindUserInfo(id)

			switch message.Type {
			case "host":
				if game != nil {
					SendMessage("error", map[string]any{"message": "You are already in a game"}, c)
				} else {
					room := entities.NewGame(connections[id])
					SendMessage("hosted", map[string]any{"room": room}, c)
				}
			case "join":
				if game != nil {
					SendMessage("error", map[string]any{"message": "You are already in a game"}, c)
				} else {
					room := strings.ToUpper(message.Data["room"].(string))
					game = entities.FindGameByRoom(room)
					if game == nil {
						SendMessage("error", map[string]any{"message": "Room not found"}, c)
					} else {
						game.Players = append(game.Players, connections[id].ToPlayer())
						SendMessage("joined", map[string]any{"room": room}, c)
					}
				}
			case "leave":
				game.RemovePlayer(id)
				SendMessage("left", map[string]any{}, c)
			case "action":
				if game == nil {
					SendMessage("error", map[string]any{"message": "You are not in a game"}, c)
					continue
				}

				switch message.Data["action"] {
				case "start":
					game.Start(id)
				case "move":
					direction := message.Data["direction"].(string)
					start := message.Data["start"].(bool)
					game.MovePlayer(id, direction, start)
				}
			case "chat":
				// TODO: Add support for commands
				to := []*websocket.Conn{}
				for _, player := range game.Players {
					conn := connections[player.User.ID]
					if conn != nil {
						to = append(to, conn.C)
					}
				}
				SendMessage("chat", map[string]any{
					"message": message.Data["message"],
					"from":    user.Username},
					to...)
			default:
				fmt.Println("Unknown message type", message.Type)
			}
		}

		c.Close()
		user.Cleanup()
		connections[id] = nil
	}))

	log.Fatal(app.Listen(":3000"))
}
