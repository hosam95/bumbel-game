package msgs

import (
	"bytes"
	"encoding/binary"
	"online-game/types"
)

type ServerMessage interface {
	Buffer() (*bytes.Buffer, bool)
}

type GenericMessage struct {
	Type uint8
	Args []byte
}

type ConnectedMessage struct {
	ID       int16
	Username string
}

type HostMessage struct{}
type HostedMessage struct {
	Room string
}

type JoinMessage struct {
	Room string
}
type JoinedMessage struct {
	Room string
}

type LeaveMessage struct{}
type LeftMessage struct{}

type StartMessage struct{}
type StartedMessage struct{}

type TeamMessage struct{}

type MoveMessage struct {
	Up    bool
	Down  bool
	Right bool
	Left  bool
	Start bool
}
type MovedMessage struct{}

type ShootMessage struct{}
type ShotMessage struct {
	X     int
	Y     int
	State types.Tile
}

type ChatMessage struct {
	Message string
}
type ChattedMessage struct {
	Message string
	From    int16
}

type MapMessage struct {
	Map types.GameMap
}

type StateMessage struct {
	Host      int16
	Room      string
	Started   bool
	StartedAt int32
	State     types.StateMessageState
	Players   []types.StateMessagePlayer
}

type SystemMessage struct {
	Type    uint8
	Message string
}
type ErrorMessage struct {
	Message string
}

const (
	MSG_CNCT    uint8 = iota
	MSG_HOST    uint8 = iota
	MSG_HOSTED  uint8 = iota
	MSG_JOIN    uint8 = iota
	MSG_JOINED  uint8 = iota
	MSG_LEAVE   uint8 = iota
	MSG_LEFT    uint8 = iota
	MSG_START   uint8 = iota
	MSG_STARTED uint8 = iota
	MSG_TEAM    uint8 = iota
	MSG_TEAMED  uint8 = iota
	MSG_MOVE    uint8 = iota
	MSG_MOVED   uint8 = iota
	MSG_SHOOT   uint8 = iota
	MSG_SHOT    uint8 = iota
	MSG_CHAT    uint8 = iota
	MSG_CHATTED uint8 = iota
	MSG_MAP     uint8 = iota
	MSG_STATE   uint8 = iota
	MSG_SYSTEM  uint8 = iota
	MSG_ERROR   uint8 = iota
	MSG_LEN     uint8 = iota
)

const (
	SYS_MSG_INFO uint8 = iota
)

type MessageError int8

const (
	MessageNoError     MessageError = iota
	MessageTooShort    MessageError = iota
	MessageInvalidType MessageError = iota
)

func ParseMessage(buf []byte) (GenericMessage, MessageError) {
	if len(buf) < 1 {
		return GenericMessage{}, MessageTooShort
	}

	t := buf[0]
	if t >= MSG_LEN {
		return GenericMessage{}, MessageInvalidType
	}

	return GenericMessage{
		Type: t,
		Args: buf[1:],
	}, MessageNoError
}

func (cm ConnectedMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{}
	buf.WriteByte(MSG_CNCT)
	// binary.Write(buf, binary.LittleEndian, cm)
	binary.Write(buf, binary.LittleEndian, cm.ID)
	buf.WriteString(cm.Username)

	return buf, true
}

func (gm GenericMessage) ParseHostMessage() (HostMessage, bool) {
	if gm.Type != MSG_HOST {
		return HostMessage{}, false
	}

	if len(gm.Args) > 0 {
		return HostMessage{}, false
	}

	return HostMessage{}, true
}

func (hm HostedMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{} // type[1] room[4]
	buf.WriteByte(MSG_HOSTED)
	buf.WriteString(hm.Room[:4])

	return buf, true
}

func (gm GenericMessage) ParseJoinMessage() (JoinMessage, bool) {
	if gm.Type != MSG_JOIN {
		return JoinMessage{}, false
	}

	if len(gm.Args) != 4 {
		return JoinMessage{}, false
	}

	return JoinMessage{Room: string(gm.Args)}, true
}

func (jm JoinedMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{}
	buf.WriteByte(MSG_JOINED)
	buf.WriteString(jm.Room[:4])

	return buf, true
}

func (gm GenericMessage) ParseLeaveMessage() (LeaveMessage, bool) {
	if gm.Type != MSG_LEAVE {
		return LeaveMessage{}, false
	}

	if len(gm.Args) > 0 {
		return LeaveMessage{}, false
	}

	return LeaveMessage{}, true
}

func (lm LeftMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{}
	buf.WriteByte(MSG_LEFT)

	return buf, true
}

func (gm GenericMessage) ParseStartMessage() (StartMessage, bool) {
	if gm.Type != MSG_START {
		return StartMessage{}, false
	}

	if len(gm.Args) > 0 {
		return StartMessage{}, false
	}

	return StartMessage{}, true
}

// func (sm StartedMessage) Buffer() (*bytes.Buffer, bool) {

// }

func (gm GenericMessage) ParseTeamMessage() (TeamMessage, bool) {
	if gm.Type != MSG_TEAM {
		return TeamMessage{}, false
	}

	if len(gm.Args) > 0 {
		return TeamMessage{}, false
	}

	return TeamMessage{}, true
}

// func (tm TeamedMessage) Buffer() (*bytes.Buffer, bool) {

// }

func (gm GenericMessage) ParseMoveMessage() (MoveMessage, bool) {
	if gm.Type != MSG_MOVE {
		return MoveMessage{}, false
	}

	if len(gm.Args) != 1 {
		return MoveMessage{}, false
	}

	flags := gm.Args[0]

	return MoveMessage{
		Up:    (flags & (1 << 0)) > 0,
		Down:  (flags & (1 << 1)) > 0,
		Left:  (flags & (1 << 2)) > 0,
		Right: (flags & (1 << 3)) > 0,
		Start: (flags & (1 << 4)) > 0,
	}, true
}

// func (mm MovedMessage) Buffer() (*bytes.Buffer, bool) {

// }

func (gm GenericMessage) ParseShootMessage() (ShootMessage, bool) {
	if gm.Type != MSG_SHOOT {
		return ShootMessage{}, false
	}

	if len(gm.Args) > 0 {
		return ShootMessage{}, false
	}

	return ShootMessage{}, true
}

func (sm ShotMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{}

	buf.WriteByte(MSG_SHOT)
	binary.Write(buf, binary.LittleEndian, int32(sm.X))
	binary.Write(buf, binary.LittleEndian, int32(sm.Y))
	buf.WriteByte(byte(sm.State))

	return buf, true
}

func (gm GenericMessage) ParseChatMessage() (ChatMessage, bool) {
	if gm.Type != MSG_CHAT {
		return ChatMessage{}, false
	}

	if len(gm.Args) < 2 || len(gm.Args) > 256 {
		return ChatMessage{}, false
	}

	sz := gm.Args[0]
	str := string(gm.Args[1 : sz+1])

	return ChatMessage{Message: str}, true
}

func (cm ChattedMessage) Buffer() (*bytes.Buffer, bool) {
	buf := &bytes.Buffer{}
	buf.WriteByte(MSG_CHATTED)
	binary.Write(buf, binary.LittleEndian, cm.From)
	binary.Write(buf, binary.LittleEndian, uint8(len(cm.Message)))
	buf.WriteString(cm.Message)

	return buf, true
}

func (mm MapMessage) Buffer() (*bytes.Buffer, bool) {
	buf := new(bytes.Buffer)

	buf.WriteByte(MSG_MAP)
	binary.Write(buf, binary.LittleEndian, int32(mm.Map.Width))
	binary.Write(buf, binary.LittleEndian, int32(mm.Map.Height))
	binary.Write(buf, binary.LittleEndian, mm.Map.Tiles)

	return buf, true
}

func (sm StateMessage) Buffer() (*bytes.Buffer, bool) {
	buf := new(bytes.Buffer)

	buf.WriteByte(MSG_STATE)
	binary.Write(buf, binary.LittleEndian, sm.Host)
	if len(sm.Room) != 4 {
		return nil, false
	}
	buf.WriteString(sm.Room)
	binary.Write(buf, binary.LittleEndian, sm.Started)
	binary.Write(buf, binary.LittleEndian, sm.StartedAt)
	binary.Write(buf, binary.LittleEndian, sm.State.TeamA)
	binary.Write(buf, binary.LittleEndian, sm.State.TeamB)
	binary.Write(buf, binary.LittleEndian, sm.State.ScoreA)
	binary.Write(buf, binary.LittleEndian, sm.State.ScoreB)
	buf.WriteByte(byte(sm.State.Phase))
	buf.WriteByte(uint8(len(sm.Players)))

	for _, player := range sm.Players {
		binary.Write(buf, binary.LittleEndian, player.User.ID)
		binary.Write(buf, binary.LittleEndian, player.Team)
		binary.Write(buf, binary.LittleEndian, player.X)
		binary.Write(buf, binary.LittleEndian, player.Y)
		binary.Write(buf, binary.LittleEndian, player.VX)
		binary.Write(buf, binary.LittleEndian, player.VY)
		binary.Write(buf, binary.LittleEndian, uint8(len(player.User.Username)))
		buf.WriteString(player.User.Username)
	}

	return buf, true
}

func (sm SystemMessage) Buffer() (*bytes.Buffer, bool) {
	sz := len(sm.Message)
	if sz > 255 {
		return nil, false
	}

	buf := &bytes.Buffer{}

	buf.WriteByte(MSG_SYSTEM)
	buf.WriteByte(sm.Type)
	buf.WriteByte(uint8(sz))
	buf.WriteString(sm.Message)

	return buf, true
}

func (em ErrorMessage) Buffer() (*bytes.Buffer, bool) {
	sz := len(em.Message)
	if sz > 255 {
		return nil, false
	}

	buf := &bytes.Buffer{}

	buf.WriteByte(MSG_ERROR)
	buf.WriteByte(uint8(sz))
	buf.WriteString(em.Message)

	return buf, true
}
