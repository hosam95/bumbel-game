package wepons

import (
	"encoding/binary"
	"errors"
	"math"
	"online-game/entities"
	"online-game/msgs"
	"online-game/types"
	"time"
)

const id = entities.GrenadeId
const name = "Grenade"
const rang_constA = 1
const rang_constB = 2
const max_range = 5
const hitBox = 3
const cooldown = 8

type Grenade struct {
	startBuildingAt time.Time
	startCoolDownAt time.Time
}

func (g *Grenade) Id() entities.WeaponId {
	return id
}

func (g *Grenade) GetCooldown() int {
	return cooldown
}

func (g *Grenade) GetCooldownLeft() float64 {
	return time.Since(g.startCoolDownAt).Seconds()
}

func (g *Grenade) Name() string {
	return name
}

func (g *Grenade) Stringify() map[string]interface{} {
	return map[string]interface{}{
		"type":     g.Name(),
		"cooldown": g.GetCooldown(),
	}
}

func (g *Grenade) OnWeaponDown(game *entities.Game, player *entities.Player, data map[string]interface{}) (map[string]interface{}, error) {
	//handle edge cases: (allredy building range, still cooling down)
	if !g.startBuildingAt.Equal(time.Time{}) {
		return nil, errors.New("allredy building range")
	}
	if time.Since(g.startCoolDownAt).Seconds() < cooldown {
		return nil, errors.New("still cooling down")
	}

	//save time
	g.startBuildingAt = time.Now()
	return nil, nil
}

func (g *Grenade) OnWeaponUpdate(game *entities.Game, player *entities.Player, data map[string]interface{}) (map[string]interface{}, error) {
	//TODO: implement the update
	return nil, nil
}

func (g *Grenade) OnWeaponUp(game *entities.Game, player *entities.Player, data map[string]interface{}) (map[string]interface{}, error) {
	x := data["x"].(float64)
	y := data["y"].(float64)
	//validate cooldown state
	if time.Since(g.startCoolDownAt).Seconds() < cooldown {
		return nil, errors.New("still cooling down")
	}

	//validate the x,y
	if y > entities.MapHeight || x > entities.MapWidth {
		return nil, errors.New("invalid coordinates")
	}

	//calculate the range
	buildTime := time.Since(g.startBuildingAt).Seconds()
	Range := (rang_constA * buildTime) + rang_constB

	if Range > max_range {
		Range = max_range
	}

	//validate the x,y are within range
	distance := math.Sqrt(((player.X - x) * (player.X - x)) + ((player.Y - y) * (player.Y - y)))
	if distance > float64(Range) {
		//if not project the x,y on the max inRange coordinates in the same direction
		seta := math.Atan2((player.Y + 0.5 - y), (player.X + 0.5 - x))
		y = player.Y - (float64(Range) * math.Sin(seta))
		x = player.X - (float64(Range) * math.Cos(seta))
	}

	//if x,y are out of map, project them on the map edge
	x, y = projectIntoMapIfOutside(player, x, y)

	//render the shoot to the map
	cornerX := int(x) - ((hitBox - 1) / 2)
	cornerY := int(y) - ((hitBox - 1) / 2)

	for i := 0; i < hitBox; i++ {
		for j := 0; j < hitBox; j++ {
			x := cornerX + i
			y := cornerY + j

			//continu if coordinates are out of map
			if y > entities.MapHeight || x > entities.MapWidth || y < 0 || x < 0 {
				continue
			}

			curr := entities.Get(&game.State.GameMap, x, y)

			//continu if the tile is a wall
			if curr == entities.WallTile {
				continue
			}

			//if the tile is allredy painted, decrement the score of the owner team
			switch curr {
			case entities.TeamATile:
				game.State.ScoreA--
			case entities.TeamBTile:
				game.State.ScoreB--
			}

			//get the tiles new team and increment the score of the new owner team
			var newTile types.Tile
			switch player.Team {
			case entities.TeamA:
				newTile = entities.TeamATile
				game.State.ScoreA++
			case entities.TeamB:
				newTile = entities.TeamBTile
				game.State.ScoreB++
			}

			//paint the tile
			entities.Set(&game.State.GameMap, x, y, newTile)
			data["updateMap"].(func(game *entities.Game, cellX, cellY int, state types.Tile))(game, x, y, newTile)
		}
	}

	//reset the time counter
	g.startBuildingAt = time.Time{}

	//start the cooldouwn
	g.startCoolDownAt = time.Now()

	return nil, nil
}

func projectIntoMapIfOutside(player *entities.Player, x float64, y float64) (float64, float64) {

	// the nonBorderCoordinate= samePlayerCoordinate + ( (sameTargetCoordinate - samePlayerCoordinate) * ( playersVerticalProjectionToBorder / playersProjectionToTargetLevelVerticalOnBorder ) );
	// the BorderCoordinate= borderCoordinate;

	if x < 0 {
		//if x<0, project the x,y on the x=0 line
		x = 0 + 0.5
		y = player.Y + 0.5 + ((y - (player.Y + 0.5)) * ((0 - (player.X + 0.5)) / (x - (player.X + 0.5))))
		if y > 0 && y < entities.MapHeight {
			return x, y
		}
	} else if x > entities.MapWidth {
		//if x>mapWidth, project the x,y on the x=mapWidth line
		x = entities.MapWidth - 0.5
		y = player.Y + 0.5 + ((y - (player.Y + 0.5)) * ((entities.MapWidth - (player.X + 0.5)) / (x - (player.X + 0.5))))
		if y > 0 && y < entities.MapHeight {
			return x, y
		}
	}

	if y < 0 {
		//if y<0, project the x,y on the y=0 line
		x = player.X + 0.5 + ((x - (player.X + 0.5)) * ((0 - (player.Y + 0.5)) / (y - (player.Y + 0.5))))
		y = 0 + 0.5

		return x, y
	} else if y > entities.MapHeight {
		//if y>mapHeight, project the x,y on the y=mapHeight line
		x = player.X + 0.5 + ((x - (player.X + 0.5)) * ((entities.MapHeight - (player.Y + 0.5)) / (y - (player.Y + 0.5))))
		y = entities.MapHeight - 0.5

		return x, y
	}

	return x, y
}

func (g *Grenade) ParseWeaponDownMessage(message msgs.GenericMessage) (map[string]interface{}, bool) {
	if message.Type != msgs.MSG_WEAPONDOWN {
		return map[string]interface{}{}, false
	}

	return map[string]interface{}{}, true
}

func (g *Grenade) ParseWeaponUpdateMessage(message msgs.GenericMessage) (map[string]interface{}, bool) {
	if message.Type != msgs.MSG_WEAPONUPDATE {
		return map[string]interface{}{}, false
	}

	return map[string]interface{}{}, true
}

func (g *Grenade) ParseWeaponUpMessage(message msgs.GenericMessage) (map[string]interface{}, bool) {
	if message.Type != msgs.MSG_WEAPONUP {
		return map[string]interface{}{}, false
	}

	args, ok := entities.CheckWeaponId(id, message.Args)
	if !ok {
		return map[string]interface{}{}, false
	}

	if len(args) != 16 {
		return map[string]interface{}{}, false
	}
	var x float64 = math.Float64frombits(binary.LittleEndian.Uint64(args[:8]))
	var y float64 = math.Float64frombits(binary.LittleEndian.Uint64(args[8:]))
	return map[string]interface{}{"x": x, "y": y}, true
}
