package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"online-game/entities"
	"online-game/structs"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func ParseMessage(msg []byte) (*structs.Message, error) {
	message := &structs.Message{}
	err := json.Unmarshal(msg, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func randomName() string {
	first := []string{"The"}
	second := []string{"Fiery", "Icy", "Electric", "Magnetic", "Toxic", "Radioactive", "Mystic", "Dark", "Light", "Wind", "Water", "Earth", "Fire"}
	third := []string{"Chicken", "Duck", "Geese", "Pigeon", "Eagle", "Falcon", "Hawk", "Owl", "Parrot", "Penguin", "Robin", "Sparrow", "Swan", "Turkey"}

	return fmt.Sprintf("%s %s %s", first[rand.Intn(len(first))], second[rand.Intn(len(second))], third[rand.Intn(len(third))])
}

func UpdateState() {
	for _, game := range entities.Games {
		game.Update()
	}
}

func BroadcastState() {
	for _, game := range entities.Games {
		state := game.Stringify()
		msg := structs.Message{
			Type: "state",
			Data: state,
		}
		jsonState, _ := json.Marshal(msg)
		for _, player := range game.Players {
			player.User.Send(jsonState)
		}
	}
}

func updateMap(game *entities.Game, cellX, cellY int, state entities.Tile) {
	msg := structs.Message{
		Type: "map",
		Data: map[string]any{
			"x":     cellX,
			"y":     cellY,
			"state": state,
		},
	}
	jsonMsg, _ := json.Marshal(msg)
	for _, player := range game.Players {
		player.User.Send(jsonMsg)
	}
}

func main() {
	app := fiber.New()

	app.Use(filesystem.New(filesystem.Config{
		Root: http.Dir("./public"),
	}))

	go func() {
		for range time.Tick(entities.GameTick) {
			UpdateState()
			BroadcastState()
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
		user.SendMessage("connected", map[string]any{"id": id, "username": user.Username})

		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		for {
			message := structs.Message{}
			if err := c.ReadJSON(&message); err != nil {
				log.Println("read:", err)
				break
			}

			game := entities.FindUserInfo(id)

			switch message.Type {
			case "host":
				if game != nil {
					user.SendMessage("error", map[string]any{"message": "You are already in a game"})
				} else {
					room := entities.NewGame(user)
					user.SendMessage("hosted", map[string]any{"room": room})
				}
			case "join":
				if game != nil {
					user.SendMessage("error", map[string]any{"message": "You are already in a game"})
				} else {
					room := strings.ToUpper(message.Data["room"].(string))
					game = entities.FindGameByRoom(room)
					if game == nil {
						user.SendMessage("error", map[string]any{"message": "Room not found"})
					} else {
						game.AddUser(user)
						user.SendMessage("joined", map[string]any{"room": room})
					}
				}
			case "leave":
				game.RemovePlayer(id)
				user.SendMessage("left", map[string]any{})
			case "action":
				if game == nil {
					user.SendMessage("error", map[string]any{"message": "You are not in a game"})
					continue
				}

				switch message.Data["action"] {
				case "start":
					err := game.Start(id)
					if err != nil {
						user.Error(err.Error())
					}
				case "team":
					err := game.SwitchTeams(id)
					if err != nil {
						user.Error(err.Error())
					}
				case "move":
					if game.State.Phase != entities.Playing {
						continue
					}
					direction := message.Data["direction"].(string)
					start := message.Data["start"].(bool)
					game.MovePlayer(id, direction, start)
				case "shoot":
					if game.State.Phase != entities.Playing {
						continue
					}
					cell, error := game.Shoot(id)
					if error != nil {
						user.Error(error.Error())
					} else {
						updateMap(game, cell.X, cell.Y, cell.State)
					}
				}
			case "chat":
				// TODO: Add support for commands
				msg := map[string]any{
					"message": message.Data["message"],
					"from":    user.Username,
				}

				for _, player := range game.Players {
					player.User.SendMessage("chat", msg)
				}
			default:
				fmt.Println("Unknown message type", message.Type)
			}
		}

		c.Close()
		user.Cleanup()
	}))

	log.Fatal(app.Listen(":3000"))
}
