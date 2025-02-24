// messages
const MESSAGES = {};

MESSAGES[MESSAGES["MSG_CNCT"] = 0] = "MSG_CNCT";
MESSAGES[MESSAGES["MSG_HOST"] = 1] = "MSG_HOST";
MESSAGES[MESSAGES["MSG_HOSTED"] = 2] = "MSG_HOSTED";
MESSAGES[MESSAGES["MSG_JOIN"] = 3] = "MSG_JOIN";
MESSAGES[MESSAGES["MSG_JOINED"] = 4] = "MSG_JOINED";
MESSAGES[MESSAGES["MSG_LEAVE"] = 5] = "MSG_LEAVE";
MESSAGES[MESSAGES["MSG_LEFT"] = 6] = "MSG_LEFT";
MESSAGES[MESSAGES["MSG_START"] = 7] = "MSG_START";
MESSAGES[MESSAGES["MSG_STARTED"] = 8] = "MSG_STARTED";
MESSAGES[MESSAGES["MSG_TEAM"] = 9] = "MSG_TEAM";
MESSAGES[MESSAGES["MSG_TEAMED"] = 10] = "MSG_TEAMED";
MESSAGES[MESSAGES["MSG_MOVE"] = 11] = "MSG_MOVE";
MESSAGES[MESSAGES["MSG_MOVED"] = 12] = "MSG_MOVED";
MESSAGES[MESSAGES["MSG_SHOOT"] = 13] = "MSG_SHOOT";
MESSAGES[MESSAGES["MSG_SHOT"] = 14] = "MSG_SHOT";
MESSAGES[MESSAGES["MSG_CHAT"] = 15] = "MSG_CHAT";
MESSAGES[MESSAGES["MSG_CHATTED"] = 16] = "MSG_CHATTED";
MESSAGES[MESSAGES["MSG_MAP"] = 17] = "MSG_MAP";
MESSAGES[MESSAGES["MSG_STATE"] = 18] = "MSG_STATE";
MESSAGES[MESSAGES["MSG_SYSTEM"] = 19] = "MSG_SYSTEM";
MESSAGES[MESSAGES["MSG_ERROR"] = 20] = "MSG_ERROR";
MESSAGES[MESSAGES["MSG_WEAPONDOWN"] = 21] = "MSG_WEAPONDOWN";
MESSAGES[MESSAGES["MSG_WEAPONUPDATE"] = 22] = "MSG_WEAPONUPDATE";
MESSAGES[MESSAGES["MSG_WEAPONUP"] = 23] = "MSG_WEAPONUP";
MESSAGES[MESSAGES["MSG_WEAPONPRESSED"] = 24] = "MSG_WEAPONPRESSED";
MESSAGES[MESSAGES["MSG_WEAPONUPDATED"] = 25] = "MSG_WEAPONUPDATED";
MESSAGES[MESSAGES["MSG_WEAPONRELEASED"] = 26] = "MSG_WEAPONRELEASED";
MESSAGES[MESSAGES["MSG_LEN"] = 27] = "MSG_LEN";


// system messages
const SYSTEM_MESSAGES = {};
SYSTEM_MESSAGES[SYSTEM_MESSAGES["SYS_MSG_INFO"] = 0] = "SYS_MSG_INFO";

// helpers
function getUint8(data, state) {
    const uint8 = data.getUint8(state.i);
    state.i += 1;
    return uint8
}

function getBoolean(data, state) {
    const boolean = data.getUint8(state.i) === 1;
    state.i += 1;
    return boolean
}

function getInt16(data, state) {
    const int16 = data.getInt16(state.i, true);
    state.i += 2;
    return int16
}

function getInt32(data, state) {
    const int32 = data.getInt32(state.i, true);
    state.i += 4;
    return int32
}

function getFloat64(data, state) {
    const float64 = data.getFloat64(state.i, true);
    state.i += 8;
    return float64
}

function getString(data, length, state) {
    const string = new TextDecoder("utf-8").decode(data.buffer.slice(state.i + data.byteOffset, state.i + length + data.byteOffset));
    state.i += length;
    return string
}

// decode

/**
 * 
 * @param {ArrayBuffer} msg 
 */
function decodeMsg(msg) {
    const type = new Uint8Array(msg, 0)[0];   
    if (!type in MESSAGES) {
        return null;
    }

    const view = new DataView(msg, 1);
    const state = { i: 0 };
    const data = {};

    switch (MESSAGES[type]) {
        case "MSG_CNCT":
            data.id = view.getInt16(0, true);
            data.username = new TextDecoder("utf-8").decode(msg.slice(3));
            break;
        case "MSG_HOSTED":
            data.room = new TextDecoder("utf-8").decode(msg.slice(1));
            break;
        case "MSG_JOINED":
            data.room = getString(view, 4, state);
            break;
        case "MSG_LEFT":
            break;
        case "MSG_SHOT":
            data.x = getInt32(view, state);
            data.y = getInt32(view, state);
            data.state = getUint8(view, state);
            break;
        case "MSG_CHATTED":
            data.from = getInt16(view, state);
            const sz = getUint8(view, state);
            data.message = getString(view, sz, state);
            break;
        case "MSG_MAP":
            data.width = getInt32(view, state);
            data.height = getInt32(view, state);
            data.tiles = Array.from(new Uint8Array(view.buffer, state.i + view.byteOffset, data.width * data.height));
            break;
        case "MSG_STATE": {
            data.host = getInt16(view, state);
            data.room = getString(view, 4, state);
            data.started = getBoolean(view, state);

            const unix = getInt32(view, state);
            if (unix <= 0) {
                data.startedAt = null;
            } else {
                data.startedAt = new Date(unix * 1000);
            }

            data.state = {};
            data.state.teamA = getInt32(view, state);
            data.state.teamB = getInt32(view, state);
            data.state.scoreA = getInt32(view, state);
            data.state.scoreB = getInt32(view, state);
            data.state.phase = getUint8(view, state);

            const playersLen = getUint8(view, state);
            data.players = [];
            for (let i = 0; i < playersLen; i++) {
                var weapon=undefined;
                try{
                    weapon = game.state.players[i].weapon;
                }catch(err){}
                
                data.players.push({
                    user: {
                        id: getInt16(view, state),
                    },
                    team: getUint8(view, state),
                    x: getFloat64(view, state),
                    y: getFloat64(view, state),
                    vx: getInt32(view, state),
                    vy: getInt32(view, state),
                    weapon: weapon?? Weapon.getWeaponObject(getUint8(view, state),i)
                });
                //if the wepon is not null read the weponId byte to skip it
                if(weapon)getUint8(view, state);

                const usernameLen = getUint8(view, state);
                data.players[i].user.username = getString(view, usernameLen, state);
            }
        } break;
        case "MSG_SYSTEM":
            const sysType = getUint8(view, state);
            if (sysType in SYSTEM_MESSAGES) {
                data.type = SYSTEM_MESSAGES[sysType];
            } else {
                throw new Error("Unknown System Message " + sysType);
            }

            const messageLen = getUint8(view, state);
            data.message = getString(view, messageLen, state);
            break;
        case "MSG_ERROR":
            const size = getUint8(view, state);
            data.message = getString(view, size, state);
            break;
        case "MSG_WEAPONPRESSED":
            data.data=Weapon.decodeWeaponPressedMSG(msg);
            break;
        case "MSG_WEAPONUPDATED":
            data.data=Weapon.decodeWeaponUpdatedMSG(msg);
            break;
        case "MSG_WEAPONRELEASED":
            data.data=Weapon.decodeWeaponReleasedMSG(msg);
            break;

        case "MSG_HOST":
        case "MSG_JOIN":
        case "MSG_LEAVE":
        case "MSG_START":
        case "MSG_TEAM":
        case "MSG_MOVE":
        case "MSG_SHOOT":
        case "MSG_CHAT":
            throw new Error("Not Recivable " + MESSAGES[type]);
    }

    return {
        type: MESSAGES[type],
        data,
    };
}

// encode

/**
 * 
 * @param {{type: string, data: any}} msg
 */
function encodeMsg(msg) {
    const type = MESSAGES[msg.type];
    if (!type) {
        throw new Error("Unknown message type", msg.type)
    }

    let buf = null;

    switch (msg.type) {
        case "MSG_CNCT":
        case "MSG_HOSTED":
        case "MSG_JOINED":
        case "MSG_LEFT":
        case "MSG_STARTED":
        case "MSG_SHOT":
        case "MSG_CHATTED":
        case "MSG_MAP":
        case "MSG_STATE":
        case "MSG_SYSTEM":
        case "MSG_ERROR":
            throw new Error("Not Sendable " + msg.type);
        case "MSG_HOST":
            buf = new Uint8Array(1);
            buf[0] = type;
            break;
        case "MSG_JOIN":
            buf = new Uint8Array(5);
            buf[0] = type;
            const { room } = msg.data;
            if (room.length > 4) {
                return { error: "Room name too long" };
            }
            const roomBuf = new TextEncoder("utf-8").encode(room);
            buf.set(roomBuf, 1);
            break;
        case "MSG_LEAVE":
        case "MSG_START":
        case "MSG_TEAM":
        case "MSG_SHOOT":
            buf = new Uint8Array(1);
            buf[0] = type;
            break;
        case "MSG_MOVE":
            const { direction, start } = msg.data;
            buf = new Uint8Array(2);
            buf[0] = type;
            let flags = 0;
            switch (direction) {
                case "up": flags |= 1 << 0; break;
                case "down": flags |= 1 << 1; break;
                case "left": flags |= 1 << 2; break;
                case "right": flags |= 1 << 3; break;
            }
            if (start) {
                flags |= 1 << 4;
            }
            buf[1] = flags;
            break;
        case "MSG_CHAT":
            const sz = msg.data.message.length;
            if (sz > 255) {
                return { error: "Message too long" };
            }
            buf = new Uint8Array(sz + 1 + 1); // for type and size
            buf[0] = type;
            buf[1] = sz;
            const msgBuf = new TextEncoder("utf-8").encode(msg.data.message);
            buf.set(msgBuf, 2);
            break;
        case "MSG_WEAPONDOWN":
            buffer = new DataView(new ArrayBuffer(10));
            // Combine type, weapon_id, and seta byte arrays
            buffer.setUint8(0,type);
            buffer.setUint8(1,WEAPONS["WEAPON_Grenade"]);
            buffer.setFloat64(2,msg.data.seta, true);
            
            // Concatenate the byte arrays
            buf = new Uint8Array(buffer.buffer);
            break;
        case "MSG_WEAPONUPDATE":
            buffer = new DataView(new ArrayBuffer(10));
            // Combine type, weapon_id, and seta byte arrays
            buffer.setUint8(0,type);
            buffer.setUint8(1,WEAPONS["WEAPON_Grenade"]);
            buffer.setFloat64(2,msg.data.seta, true);
            
            // Concatenate the byte arrays
            buf = new Uint8Array(buffer.buffer);
            break;
        case "MSG_WEAPONUP":
            buffer = new DataView(new ArrayBuffer(18));
            // Combine x and y byte arrays
            buffer.setUint8(0,type);
            buffer.setUint8(1,WEAPONS["WEAPON_Grenade"]);
            buffer.setFloat64(2,msg.data.x, true);
            buffer.setFloat64(10,msg.data.y, true);
            
            // Concatenate the byte arrays
            buf = new Uint8Array(buffer.buffer);
            break;
        
    }

    return buf;
}