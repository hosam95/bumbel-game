const game = { state: null, ctx: null, map: null };
let activeScreen = 0; // 0: Home, 1: Game
let lastTimestamp = 0;
const myData = {
    id: null,
    username: null,
};
let isServerUpdated = false;

let rendering = false;
let isAiming=false
let myLocation = { x: 0, y: 0 };
let targetLocation = { x: 0, y: 0 };
let startBuildingAt = null;
const rang_constA = 1
const rang_constB = 2
const max_cell_range = 5

// CONSTANTS
const playerSpeed = 10;
const gameDuration = 60 * 1000; // 1 minute
// Game States
const WaitingForPlayers = 0;
const Playing = 1;
const GameOver = 2;
// Tile Types
const EmptyTile = 0;
const TeamATile = 1;
const TeamBTile = 2;
const WallTile = 3;

function HomeScreen(root, handlers) {
    activeScreen = 0;

    const center = document.createElement("div");
    center.classList.add("center", "full-height");

    const container = document.createElement("div");
    center.appendChild(container);

    const h1 = document.createElement("h1");
    h1.textContent = "Online Game";
    container.appendChild(h1);

    const roomBox = document.createElement("div");
    roomBox.className = "row";
    container.appendChild(roomBox);

    const span = document.createElement("span");
    span.textContent = "Room ID:";
    roomBox.appendChild(span);

    const roomInput = document.createElement("input");
    roomInput.type = "text";
    roomBox.appendChild(roomInput);

    const joinBtn = document.createElement("button");
    joinBtn.classList.add("btn");
    joinBtn.textContent = "Join Room";
    roomBox.appendChild(joinBtn);

    const hostBtn = document.createElement("button");
    hostBtn.classList.add("btn", "secondary");
    hostBtn.textContent = "Host Room";
    roomBox.appendChild(hostBtn);

    root.replaceChildren(center);

    roomInput.addEventListener("keydown", (e) => {
        if (e.key === "Enter") {
            handlers.joinRoom(roomInput);
        }
    })

    joinBtn.addEventListener("click", () => {
        handlers.joinRoom(roomInput);
    });

    hostBtn.addEventListener("click", () => {
        handlers.hostRoom();
    });
}

function GameScreen(root, handlers) {
    activeScreen = 1;

    const container = document.createElement("div");
    container.classList.add("game-container", "full-height");

    const canvasContainer = document.createElement("div");
    canvasContainer.classList.add("canvas-container", "center");
    container.appendChild(canvasContainer);

    const canvas = document.createElement("canvas");
    canvas.id = "canvas";
    canvas.width = 1600;
    canvas.height = 900;
    canvas.tabIndex = 1;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("Failed to get 2d context");
    game.ctx = ctx;
    canvasContainer.appendChild(canvas);

    const chat = document.createElement("div");
    chat.classList.add("chat");
    container.appendChild(chat);

    const titleRow = document.createElement("div");
    titleRow.classList.add("row", "between");
    chat.appendChild(titleRow);

    const chatTitle = document.createElement("h2");
    chatTitle.classList.add("chat-title");
    chatTitle.textContent = "Chat";
    titleRow.appendChild(chatTitle);

    const leaveBtn = document.createElement("button");
    leaveBtn.classList.add("btn", "danger");
    leaveBtn.textContent = "Leave Room";
    titleRow.appendChild(leaveBtn);

    const chatBoxContainer = document.createElement("div");
    chatBoxContainer.classList.add("chat-box-container");
    chat.appendChild(chatBoxContainer);

    const chatBox = document.createElement("div");
    chatBox.id = "chatBox";
    chatBox.classList.add("chat-box");
    chatBoxContainer.appendChild(chatBox);

    const chatSend = document.createElement("div");
    chatSend.classList.add("chat-send");
    chat.appendChild(chatSend);

    const chatInput = document.createElement("input");
    chatInput.type = "text";
    chatSend.appendChild(chatInput);

    const chatBtn = document.createElement("button");
    chatBtn.type = "button";
    chatBtn.classList.add("btn");
    chatBtn.textContent = "Send";
    chatSend.appendChild(chatBtn);

    root.replaceChildren(container);

    handlers.setupGameControls(canvas);

    chatInput.addEventListener("keydown", (e) => {
        if (e.key === "Enter") {
            handlers.chat(chatInput);
        }
    });

    chatBtn.addEventListener("click", () => {
        handlers.chat(chatInput);
    });

    leaveBtn.addEventListener("click", () => {
        handlers.leaveRoom();
    });

    appendSystemMessage("SYS_MSG_INFO", "Welcome to the game");
    appendSystemMessage("SYS_MSG_SUCCESS", "Your username is " + myData.username);
    appendSystemMessage("SYS_MSG_INFO", "Use arrow keys to move");
    appendSystemMessage("SYS_MSG_INFO", "Use Z to shoot");
    appendSystemMessage("SYS_MSG_INFO", "Use T to change team");
    appendSystemMessage("SYS_MSG_INFO", "Use Q to start the game");
    appendSystemMessage("SYS_MSG_SUCCESS", "Have fun!");
}

let one = false;
function tick(ts) {
    if (!game.state) { console.log("no game state"); rendering = false; return; }
    if (!game.ctx) { console.log("no game ctx"); rendering = false; return; }
    const dt = (ts - lastTimestamp) / 1000;
    lastTimestamp = ts;
    const { width, height } = game.ctx.canvas;
    const wOffset = width * 0.1;
    const wRest = width - wOffset;
    const hOffset = height * 0.1;
    const hRest = height - hOffset;

    const { state: gameState, ctx, map } = game;

    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);

    // Update
    if (gameState.started && !isServerUpdated) {
        if (one) {
            console.log(gameState);
            one = false;
        }
        for (const player of gameState.players) {
            let newX = player.x + player.vx * dt * playerSpeed;
            let newY = player.y + player.vy * dt * playerSpeed;

            const { tile, bottom, right, bottomRight } = getAroundMap(map, Math.floor(newX), Math.floor(newY));
            const cornerX = newX - Math.floor(newX) > 0;
            const cornerY = newY - Math.floor(newY) > 0;

            if (player.vx > 0) {
                if (right === WallTile || (cornerY && bottomRight === WallTile)) {
                    newX = Math.floor(newX);
                }
            }

            if (player.vx < 0) {
                if (tile === WallTile || (cornerY && bottom === WallTile)) {
                    newX = Math.ceil(newX);
                }
            }

            if (player.vy > 0) {
                if (bottom === WallTile || (cornerX && bottomRight === WallTile)) {
                    newY = Math.floor(newY);
                }
            }

            if (player.vy < 0) {
                if (tile === WallTile || (cornerX && right === WallTile)) {
                    newY = Math.ceil(newY);
                }
            }

            player.x = newX;
            player.y = newY;
        }
    } else {
        isServerUpdated = false;
    }

    // Render
    const teamAColor = "#" + gameState.state.teamA.toString(16).padStart(6, "0");
    const teamBColor = "#" + gameState.state.teamB.toString(16).padStart(6, "0");

    // Bars
    ctx.fillStyle = "#353535";
    ctx.fillRect(0, 0, width, height);

    ctx.fillStyle = "#f0f0f0";
    ctx.font = "30px Arial";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(`Room: ${gameState.room}`, wOffset / 2, hOffset / 2, wOffset - 16);

    // Top bar
    if (gameState.started) {
        const started = new Date(gameState.startedAt);
        const now = new Date();
        const left = gameDuration - (now - started);
        ctx.fillStyle = "#f0f0f0";
        if (left <= 0) {
            ctx.fillText("Time is up", wOffset + wRest / 2, hOffset / 2);
        } else {
            const minutes = Math.floor(left / 1000 / 60).toString().padStart(2, "0");
            const seconds = (Math.floor(left / 1000) % 60).toString().padStart(2, "0");
            ctx.fillText(`${minutes}:${seconds}`, wOffset + wRest / 2, hOffset / 2);
        }
    } else {
        ctx.fillStyle = "#f0f0f0";
        ctx.fillText("01:00", width / 2, hOffset / 2);
    }

    // Sidebar
    const sidebarCenter = hOffset + hRest / 2;
    // score
    const score = gameState.state.phase === WaitingForPlayers ? "-" : `${gameState.state.scoreA} - ${gameState.state.scoreB}`;
    ctx.fillStyle = "#f0f0f0";
    ctx.fillText(score, wOffset / 2, sidebarCenter, wOffset - 20);

    // Teams
    const teamA = gameState.players.filter((p) => p.team === 0);
    const teamB = gameState.players.filter((p) => p.team === 1);
    const squareSize = wOffset - 20;

    // team A
    ctx.fillStyle = teamAColor;
    ctx.fillRect(10, sidebarCenter - 30 - squareSize, squareSize, squareSize);
    ctx.fillStyle = "#f0f0f0";
    ctx.fillText(`Team A (${teamA.length})`, wOffset / 2, sidebarCenter - 55 - squareSize, wOffset - 20);

    // team B
    ctx.fillStyle = teamBColor;
    ctx.fillRect(10, sidebarCenter + 30, squareSize, squareSize);
    ctx.fillStyle = "#f0f0f0";
    ctx.fillText(`Team B (${teamB.length})`, wOffset / 2, sidebarCenter + 55 + squareSize, wOffset - 20);

    // Map
    ctx.fillStyle = "#FFD35A";
    ctx.fillRect(wOffset, hOffset, wRest, hRest);

    // Render map
    if (gameState.started) {
        const { width: mapWidth, height: mapHeight } = map;
        const cellWidth = Math.floor(wRest / mapWidth);
        const mapWidthOffset = wOffset / cellWidth;
        const cellHeight = Math.floor(hRest / mapHeight);
        const mapHeightOffset = hOffset / cellHeight;

        for (let i = 0; i < mapWidth * mapHeight; i++) {
            const x = (i % mapWidth) + mapWidthOffset;
            const y = Math.floor(i / mapWidth) + mapHeightOffset;

            switch (map.tiles[i]) {
                case EmptyTile: {
                    // Empty
                } break;
                case TeamATile: {
                    ctx.fillStyle = teamAColor;
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                } break;
                case TeamBTile: {
                    ctx.fillStyle = teamBColor;
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                } break;
                case WallTile: {
                    ctx.fillStyle = "#FFA823";
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                } break;
            }
        }

        // Render players
        if (gameState.state.phase === 1) {
            for (const player of gameState.players) {
                const x = player.x + mapWidthOffset;
                const y = player.y + mapHeightOffset;
                const color = player.team === 0 ? teamAColor : teamBColor;

                if (player.user.id === myData.id) {
                    myLocation = { x:(x + 0.5)*cellWidth, y:(y + 0.5)*cellHeight};
                    ctx.fillStyle = myData.id === gameState.host ? "#fcbe03" : "#ffffff";
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                    ctx.fillStyle = color;
                    ctx.fillRect((x + 0.1) * cellWidth, (y + 0.1) * cellHeight, 0.8 * cellWidth, 0.8 * cellHeight);
                } else if (player.user.id === gameState.host) {
                    ctx.fillStyle = "#fcbe03";
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                    ctx.fillStyle = color;
                    ctx.fillRect((x + 0.1) * cellWidth, (y + 0.1) * cellHeight, 0.8 * cellWidth, 0.8 * cellHeight);
                } else {
                    ctx.fillStyle = color;
                    ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
                }
            }
        }

        // Render the aiming
        if(isAiming){
            ctx.beginPath();
            ctx.strokeStyle = "gray";
            ctx.moveTo(myLocation.x, myLocation.y);
            ctx.lineTo(targetLocation.x, targetLocation.y);
            ctx.stroke();  
            
            ctx.fillRect(targetLocation.x-5 , targetLocation.y-5, 10, 10);
        }
    } else {
        ctx.fillStyle = "#353535";
        ctx.font = "60px Arial";
        switch (gameState.state.phase) {
            case WaitingForPlayers: {
                ctx.fillText("Waiting for players", wOffset + wRest / 2, hOffset + hRest / 2);
            } break;
            case GameOver: {
                const winner = gameState.state.scoreA > gameState.state.scoreB ? "Team A Wins"
                    : gameState.state.scoreA === gameState.state.scoreB ? "It's a Tie" : "Team B Wins";
                ctx.fillText(`Game Over! ${winner}`, wOffset + wRest / 2, hOffset + hRest / 2);
            } break;
        }
    }

    requestAnimationFrame(tick);
}

requestAnimationFrame(tick);

function appendMessage(from, message) {
    const chatBox = document.getElementById("chatBox");
    if (!chatBox) return false;

    const chatMessage = document.createElement("div");
    chatMessage.classList.add("chat-message");
    chatBox.appendChild(chatMessage);

    const sender = document.createElement("span");
    sender.textContent = from;
    sender.classList.add("chat-sender");
    chatMessage.appendChild(sender);

    const msg = document.createElement("span");
    msg.textContent = message;
    msg.classList.add("chat-text");
    chatMessage.appendChild(msg);

    chatBox.scrollTop = chatBox.scrollHeight;

    return true;
}

function typeToString(type) {
    switch (type) {
        case "SYS_MSG_INFO": return "info";
        case "SYS_MSG_ERROR": return "error";
        case "SYS_MSG_SUCCESS": return "success";
        default: return "unknown";
    }
}

function usernameFromId(id) {
    for (const player of game.state.players) {
        if (player.user.id === id) {
            return player.user.username;
        }
    }
    return "Unknown";
}

function appendSystemMessage(type, message) {
    const chatBox = document.getElementById("chatBox");
    if (!chatBox) return false;

    const chatMessage = document.createElement("div");
    const t = typeToString(type);
    chatMessage.classList.add("chat-message", t);
    chatBox.appendChild(chatMessage);

    const msg = document.createElement("span");
    msg.textContent = message;
    msg.classList.add("chat-text");
    chatMessage.appendChild(msg);

    chatBox.scrollTop = chatBox.scrollHeight;

    return true;
}

(() => {
    const root = document.getElementById("root");

    let ws = new WebSocket("/ws");
    ws.binaryType = "arraybuffer";
    setupWSListeners(ws, {
        joinRoom,
        hostRoom,
        leaveRoom,
        startGame,
        chat,
        setupGameControls,
    },
        root
    );

    function setupGameControls(canvas) {
        canvas.addEventListener("keydown", (e) => {
            if (e.repeat) return;
            switch (e.code) {
                case "KeyW":
                case "ArrowUp":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "up",
                                    start: true,
                                },
                            })
                        );
                    }
                    break;
                case "KeyS":
                case "ArrowDown":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "down",
                                    start: true,
                                },
                            })
                        );
                    }
                    break;
                case "KeyA":
                case "ArrowLeft":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "left",
                                    start: true,
                                },
                            })
                        );
                    }
                    break;
                case "KeyD":
                case "ArrowRight":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "right",
                                    start: true,
                                },
                            })
                        );
                    }
                    break;
                case "KeyQ":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_START",
                            })
                        );
                    }
                    break;
                case "Space":
                case "KeyZ":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_SHOOT",
                            })
                        );
                    }
                    break;
                case "KeyT":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_TEAM",
                            })
                        );
                    }
                    break;
                case "KeyR":
                    {
                        one = true;
                    }
                    break;
            }
        });

        canvas.addEventListener("keyup", (e) => {
            if (e.repeat) return;
            switch (e.code) {
                case "KeyW":
                case "ArrowUp":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "up",
                                    start: false,
                                },
                            })
                        );
                    }
                    break;
                case "KeyS":
                case "ArrowDown":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "down",
                                    start: false,
                                },
                            })
                        );
                    }
                    break;
                case "KeyA":
                case "ArrowLeft":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "left",
                                    start: false,
                                },
                            })
                        );
                    }
                    break;
                case "KeyD":
                case "ArrowRight":
                    {
                        ws.send(
                            encodeMsg({
                                type: "MSG_MOVE",
                                data: {
                                    direction: "right",
                                    start: false,
                                },
                            })
                        );
                    }
                    break;
            }
        });

        canvas.addEventListener("mousedown",(e)=>{
            if(!game.state || !game.state.started){
                return;
            }

            ws.send(
                encodeMsg({
                    type: "action",
                    data: {
                        action: "powerupPressed",
                    },
                })
            );
            /**@todo:vesioalize the aiming */
            isAiming=true;
            startBuildingAt = startBuildingAt?? Date.now();
            projectMouseLocationToTarget()
        })

        canvas.addEventListener("mouseup",(e)=>{
            if(!game.state || !game.state.started){
                return;
            }

            let rect=canvas.getBoundingClientRect()
            
            const get_mouseX= ()=>{return e.clientX - (((rect.right-rect.left)*0.1)+rect.left)}
            const get_mouseY= ()=>{ return e.clientY - (((rect.bottom-rect.top)*0.1)+rect.top)}
            const get_mapWidth=()=>{ return ((rect.right-rect.left)*0.9)}
            const get_mapHeight=()=>{return ((rect.bottom-rect.top)*0.9)}
            
            let cellX= (get_mouseX()/get_mapWidth())*game.map.width
            let cellY= (get_mouseY()/get_mapHeight())*game.map.height;
            
            ws.send(
                encodeMsg({
                    type: "action",
                    data: {
                        action: "powerupReleased",
                        x: cellX,
                        y:cellY,
                    },
                })
            );

            //stop rendering the aiming
            isAiming=false;
            startBuildingAt=null;

            /**@todo: render a throwing action */
        })

        document.getElementById("root").addEventListener("mousemove",(e)=>{
            if(!isAiming){
                return;
            }

            let rect=canvas.getBoundingClientRect()
            
            const get_mouseX= ()=>{return e.clientX - (((rect.right-rect.left)*0.1)+rect.left)}
            const get_mouseY= ()=>{ return e.clientY - (((rect.bottom-rect.top)*0.1)+rect.top)}
            const get_mapWidth=()=>{ return ((rect.right-rect.left)*0.9)}
            const get_mapHeight=()=>{return ((rect.bottom-rect.top)*0.9)}
            
            let cellX= (get_mouseX()/get_mapWidth())*game.map.width
            let cellY= (get_mouseY()/get_mapHeight())*game.map.height;
            
            
            targetLocation = { x: cellX, y: cellY };
            projectMouseLocationToTarget();
        })
    }

    function joinRoom(roomInput) {
        const room = roomInput.value;
        const buf = encodeMsg({
            type: "MSG_JOIN",
            data: { room },
        })
        if (buf.error) {
            alert(buf.error);
            return;
        }
        ws.send(buf);
    }

    function hostRoom() {
        ws.send(
            encodeMsg({
                type: "MSG_HOST",
            })
        );
    }

    function leaveRoom() {
        ws.send(
            encodeMsg({
                type: "MSG_LEAVE",
            })
        );
    }

    function startGame() {
        rendering = true;
        requestAnimationFrame((timestamp) => {
            lastTimestamp = timestamp;
            tick(timestamp);
        });
    }

    function chat(chatInput) {
        const message = chatInput.value;
        if (!message) return;
        const buf = encodeMsg({
            type: "MSG_CHAT",
            data: { message },
        });

        if (buf.error) {
            appendSystemMessage("SYS_MSG_ERROR", buf.error);
            return;
        }

        ws.send(buf);
        chatInput.value = "";
    }

    HomeScreen(root, {
        joinRoom,
        hostRoom,
    });
})();

function setupWSListeners(ws, handlers, root) {
    const {
        joinRoom,
        hostRoom,
        leaveRoom,
        chat,
        startGame,
        setupGameControls,
    } = handlers;
    ws.addEventListener("open", () => {
        console.log("Connected");
    });
    ws.addEventListener("message", (event) => {
        const msg = decodeMsg(event.data);
        switch (msg.type) {
            case "MSG_CNCT":
                {
                    myData.id = msg.data.id;
                    myData.username = msg.data.username;
                }
                break;
            case "MSG_HOSTED":
                {
                    GameScreen(root, {
                        leaveRoom,
                        chat,
                        setupGameControls,
                    });
                }
                break;
            case "MSG_JOINED":
                {
                    GameScreen(root, {
                        leaveRoom,
                        chat,
                        setupGameControls,
                    });
                }
                break;
            case "MSG_STATE":
                {
                    if (!game.state || !rendering) {
                        game.state = msg.data;
                        startGame();
                    } else {
                        game.state = msg.data;
                    }
                    isServerUpdated = true;

                    if (activeScreen !== 1) {
                        console.error("should be unreachable");
                    }
                }
                break;
            case "MSG_MAP":
                {
                    game.map = msg.data;
                }
                break;
            case "MSG_LEFT":
                {
                    HomeScreen(root, {
                        joinRoom,
                        hostRoom,
                    });
                    gameState = undefined;
                }
                break;
            case "MSG_CHATTED":
                {
                    appendMessage(usernameFromId(msg.data.from), msg.data.message);
                }
                break;
            case "MSG_ERROR":
                {
                    if (!appendSystemMessage("SYS_MSG_ERROR", msg.data.message)) {
                        // TODO: find a better way to display error messages
                        alert(msg.data.message);
                    }
                }
                break;
            case "MSG_SYSTEM":
                {
                    if (!appendSystemMessage(msg.data.type, msg.data.message)) {
                        console.log(msg.data.type, msg.data.msg);
                    }
                }
                break;
            case "MSG_SHOT":
                {
                    const { x, y, state } = msg.data;
                    game.map.tiles[y * game.map.width + x] = state;
                }
                break;
            default: {
                console.error("Unknown message type:", msg.type);
            }
        }
    });
    ws.addEventListener("close", () => {
        console.log("Disconnected");
        HomeScreen(
            root,
            {},
            {
                joinRoom,
                hostRoom,
            }
        );
        game.state = undefined;
        myData.id = undefined;
        myData.username = undefined;
        // TODO: reconnect
    });
}

function getFromMap(map, x, y) {
    if (x < 0 || y < 0 || x >= map.width || y >= map.height) {
        return WallTile;
    }
    return map.tiles[y * map.width + x];
}

function getAroundMap(map, x, y) {
    const tile = getFromMap(map, x, y);
    const bottom = getFromMap(map, x, y + 1);
    const right = getFromMap(map, x + 1, y);
    const bottomRight = getFromMap(map, x + 1, y + 1);
    return { tile, bottom, right, bottomRight };
}

function projectMouseLocationToTarget() {
    let x= targetLocation.x
    let y= targetLocation.y
    
    let mapWidth=game.map.width, mapHeight = game.map.height;

    let cellWidth = 1600*0.9/mapWidth, cellHeight = 900*0.9/mapHeight
    
    let px= (myLocation.x- 1600*0.1)/cellWidth
    let py= (myLocation.y- 900*0.1)/cellHeight

    //calculate the range
	let buildTime = (Date.now() - startBuildingAt)/1000

	let range = (rang_constA * buildTime) + rang_constB

	if (range > max_cell_range ) {
		range = max_cell_range 
	}
    
	//validate the x,y are within range
	let distance = Math.sqrt(((px - x) * (px - x)) + ((py - y) * (py - y)))
	if (distance > range){
		//if not project the x,y on the max inRange coordinates in the same direction
		let seta = Math.atan2((py + 0.5 - y), (px + 0.5 - x))
        console.log("seta:",seta)
		y = py - (range * Math.sin(seta)) 
		x = px - (range * Math.cos(seta))
	}

    targetLocation = { x: x, y: y }

	//if x,y are out of map, project them on the map edge
	projectIntoMapIfOutside( mapWidth, mapHeight)

    targetLocation = { x: (targetLocation.x * cellWidth) + (1600 * 0.1), y: (targetLocation.y * cellHeight) + (900 * 0.1) }
}

function projectIntoMapIfOutside( mapWidth, mapHeight) {
	// the nonBorderCoordinate= samePlayerCoordinate + ( (sameTargetCoordinate - samePlayerCoordinate) * ( playersVerticalProjectionToBorder / playersProjectionToTargetLevelVerticalOnBorder ) );
	// the BorderCoordinate= borderCoordinate;

    let x= targetLocation.x
    let y= targetLocation.y

	if (x < 0) {
		//if x<0, project the x,y on the x=0 line
		x = 0 
		y = myLocation.y + ((y - (myLocation.y )) * ((0 - (myLocation.x )) / (x - (myLocation.x ))))
		if (y > 0 && y < mapHeight) {
			targetLocation = { x: x, y: y }
            return;
		}
	} else if (x > mapWidth) {
		//if x>mapWidth, project the x,y on the x=mapWidth line
		x = mapWidth 
		y = myLocation.y + ((y - (myLocation.y)) * ((mapWidth - (myLocation.x )) / (x - (myLocation.x))))
		if (y > 0 && y < mapHeight) {
			targetLocation = { x: x, y: y }
            return;
		}
	}

	if (y < 0) {
		//if y<0, project the x,y on the y=0 line
		x = myLocation.x + 0.5 + ((x - (myLocation.x + 0.5)) * ((0 - (myLocation.y + 0.5)) / (y - (myLocation.y + 0.5))))
		y = 0 + 0.5

		targetLocation = { x: x, y: y }
        return;
	} else if (y > mapHeight) {
		//if y>mapHeight, project the x,y on the y=mapHeight line
		x = myLocation.x + 0.5 + ((x - (myLocation.x + 0.5)) * ((mapHeight - (myLocation.y + 0.5)) / (y - (myLocation.y + 0.5))))
		y = mapHeight - 0.5

		targetLocation = { x: x, y: y }
        return;
	}
}