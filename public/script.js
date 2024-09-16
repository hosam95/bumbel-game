const playerSpeed = 10;

const game = { state: null, ctx: null, map: null };
let activeScreen = 0; // 0: Home, 1: Game
let lastTimestamp = 0;
let rendering = false;
const myData = {
    id: null,
    username: null,
}
let isServerUpdated = false;

// CONSTANTS
// Game States
const WaitingForPlayers = 0
const Playing = 1
const GameOver = 2

// Tile Types
const EmptyTile = 0
const TeamATile = 1
const TeamBTile = 2
const WallTile = 3

function HomeScreen(root, _data, handlers) {
    activeScreen = 0;

    const center = document.createElement('div');
    center.classList.add('center', "full-height");

    const container = document.createElement('div');
    center.appendChild(container);

    const h1 = document.createElement('h1');
    h1.textContent = 'Online Game';
    container.appendChild(h1);

    const roomBox = document.createElement('div');
    roomBox.className = 'row';
    container.appendChild(roomBox);

    const span = document.createElement('span');
    span.textContent = 'Room ID:';
    roomBox.appendChild(span);

    const roomInput = document.createElement('input');
    roomInput.type = 'text';
    roomBox.appendChild(roomInput);

    const joinBtn = document.createElement('button');
    joinBtn.classList.add('btn');
    joinBtn.textContent = 'Join Room';
    roomBox.appendChild(joinBtn);

    const hostBtn = document.createElement('button');
    hostBtn.classList.add('btn', "secondary");
    hostBtn.textContent = 'Host Room';
    roomBox.appendChild(hostBtn);

    root.replaceChildren(center);

    joinBtn.addEventListener('click', () => {
        handlers.joinRoom(roomInput);
    });

    hostBtn.addEventListener('click', () => {
        handlers.hostRoom();
    })
}

function GameScreen(root, data, handlers) {
    activeScreen = 1;

    const container = document.createElement('div');
    container.classList.add("game-container", "full-height");

    const canvasContainer = document.createElement('div');
    canvasContainer.classList.add("canvas-container", "center");
    container.appendChild(canvasContainer);

    const canvas = document.createElement('canvas');
    canvas.id = 'canvas';
    canvas.width = 1600;
    canvas.height = 900;
    const ctx = canvas.getContext('2d');
    if (!ctx) throw new Error('Failed to get 2d context');
    game.ctx = ctx;
    canvasContainer.appendChild(canvas);

    const chat = document.createElement('div');
    chat.classList.add("chat");
    container.appendChild(chat);

    const titleRow = document.createElement('div');
    titleRow.classList.add("row", "between");
    chat.appendChild(titleRow);

    const chatTitle = document.createElement('h2');
    chatTitle.classList.add("chat-title");
    chatTitle.textContent = 'Chat';
    titleRow.appendChild(chatTitle);

    const leaveBtn = document.createElement('button');
    leaveBtn.classList.add("btn", "danger");
    leaveBtn.textContent = 'Leave Room';
    titleRow.appendChild(leaveBtn);

    const chatBox = document.createElement('div');
    chatBox.id = 'chatBox';
    chatBox.classList.add("chat-box");
    chat.appendChild(chatBox);

    const chatSend = document.createElement('div');
    chatSend.classList.add("chat-send");
    chat.appendChild(chatSend);

    const chatInput = document.createElement('input');
    chatInput.type = 'text';
    chatSend.appendChild(chatInput);

    const chatBtn = document.createElement('button');
    chatBtn.type = 'button';
    chatBtn.classList.add("btn");
    chatBtn.textContent = 'Send';
    chatSend.appendChild(chatBtn);

    root.replaceChildren(container);

    chatInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            handlers.chat(chatInput);
        }
    });

    chatBtn.addEventListener('click', () => {
        handlers.chat(chatInput);
    });

    leaveBtn.addEventListener('click', () => {
        handlers.leaveRoom();
    });
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
    if (!isServerUpdated) {
        if (one) {
            console.log(gameState);
            one = false;
        }
        for (const player of gameState.players) {
            let newX = player.x + player.vx * dt * playerSpeed;
            let newY = player.y + player.vy * dt * playerSpeed;

            const { tile, bottom, right, bottomRight } = getAroundMap(map, Math.floor(newX), Math.floor(newY))
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

            player.x = newX
            player.y = newY
        }
    } else {
        isServerUpdated = false;
    }

    // Render
    const teamAColor = "#" + gameState.state.teamA.toString(16).padStart(6, '0');
    const teamBColor = "#" + gameState.state.teamB.toString(16).padStart(6, '0');

    // Top bar
    ctx.fillStyle = "#353535";
    ctx.fillRect(0, 0, width, hOffset); // top 10% of canvas
    // TODO: render timer when implemented

    // Sidebar
    ctx.fillStyle = "#353535";
    ctx.fillRect(0, 0, wOffset, height); // left 10% of canvas
    // TODO: render score when implemented

    // Map
    ctx.fillStyle = "#FFD35A";
    ctx.fillRect(wOffset, hOffset, wRest, hRest);

    const { width: mapWidth, height: mapHeight } = map;
    const cellWidth = Math.floor(wRest / mapWidth);
    const mapWidthOffset = wOffset / cellWidth;
    const cellHeight = Math.floor(hRest / mapHeight);
    const mapHeightOffset = hOffset / cellHeight;

    // Render map
    for (let i = 0; i < mapWidth * mapHeight; i++) {
        const x = i % mapWidth + mapWidthOffset;
        const y = Math.floor(i / mapWidth) + mapHeightOffset;

        switch (map.tiles[i]) {
            case 0: {
                // Empty
            } break;
            case 1: {
                ctx.fillStyle = teamAColor;
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
            case 2: {
                ctx.fillStyle = teamBColor;
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
            case 3: {
                ctx.fillStyle = "#FFA823";
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
        }
    }

    // Render players
    if (gameState.state.phase === 1) {
        for (const player of gameState.players) {
            if (player.user.id === myData.id) {
                ctx.fillStyle = "#ffffff";
                ctx.fillRect(player.x * cellWidth, player.y * cellHeight, cellWidth, cellHeight);
                ctx.fillStyle = player.team === 0 ? teamAColor : teamBColor;
                ctx.fillRect(player.x * cellWidth + 0.1 * cellWidth, player.y * cellHeight + 0.1 * cellHeight, cellWidth * 0.8, cellHeight * 0.8);
            } else {
                ctx.fillStyle = player.team === 0 ? teamAColor : teamBColor;
                ctx.fillRect(player.x * cellWidth, player.y * cellHeight, cellWidth, cellHeight);
            }
        }
    }

    requestAnimationFrame(tick);
}

requestAnimationFrame(tick);

function appendMessage(from, message) {
    const chatBox = document.getElementById('chatBox');
    if (!chatBox) return false;

    const chatMessage = document.createElement('div');
    chatMessage.classList.add('chat-message');
    chatBox.appendChild(chatMessage);

    const sender = document.createElement('span');
    sender.textContent = from;
    sender.classList.add('chat-sender');
    chatMessage.appendChild(sender);

    const msg = document.createElement('span');
    msg.textContent = message;
    msg.classList.add('chat-text');
    chatMessage.appendChild(msg);

    return true;
}

function appendSystemMessage(type, message) {
    const chatBox = document.getElementById('chatBox');
    if (!chatBox) return false;

    const chatMessage = document.createElement('div');
    chatMessage.classList.add('chat-message', type);
    chatBox.appendChild(chatMessage);

    const msg = document.createElement('span');
    msg.textContent = message;
    msg.classList.add('chat-text');
    chatMessage.appendChild(msg);

    return true;
}

(() => {
    const root = document.getElementById('root');

    let ws = new WebSocket('/ws');
    setupWSListeners(ws, {
        joinRoom,
        hostRoom,
        leaveRoom,
        startGame,
        chat
    }, root);

    window.addEventListener("keydown", (e) => {
        if (e.repeat) return;
        switch (e.code) {
            case "ArrowUp": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'up', start: true }
                }));
            } break;
            case "ArrowDown": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'down', start: true }
                }));
            } break;
            case "ArrowLeft": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'left', start: true }
                }));
            } break;
            case "ArrowRight": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'right', start: true }
                }));
            } break;
            case "KeyQ": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "start" }
                }));
            } break;
            case "KeyZ": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "shoot" }
                }));
            } break;
            case "KeyT": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "team" }
                }));
            } break;
            case "KeyR": {
                one = true;
            } break;
        }
    })

    window.addEventListener("keyup", (e) => {
        if (e.repeat) return;
        switch (e.code) {
            case "ArrowUp": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'up', start: false }
                }));
            } break;
            case "ArrowDown": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'down', start: false }
                }));
            } break;
            case "ArrowLeft": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'left', start: false }
                }));
            } break;
            case "ArrowRight": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "move", direction: 'right', start: false }
                }));
            } break;
        }
    })

    function joinRoom(roomInput) {
        const room = roomInput.value;
        ws.send(JSON.stringify({
            type: 'join',
            data: { room }
        }));
    }

    function hostRoom() {
        ws.send(JSON.stringify({
            type: 'host',
            data: {}
        }));
    }

    function leaveRoom() {
        ws.send(JSON.stringify({
            type: 'leave',
            data: {}
        }));
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
        ws.send(JSON.stringify({
            type: 'chat',
            data: { message }
        }));
        chatInput.value = '';
    }

    HomeScreen(root, {}, {
        joinRoom,
        hostRoom
    });
})()

function setupWSListeners(ws, handlers, root) {
    const { joinRoom, hostRoom, leaveRoom, chat, startGame } = handlers;
    ws.addEventListener("open", () => {
        console.log('Connected');
    })
    ws.addEventListener("message", (event) => {
        const msg = JSON.parse(event.data);
        switch (msg.type) {
            case "connected": {
                myData.id = msg.data.id;
                myData.username = msg.data.username;
            } break;
            case "hosted": {
                GameScreen(root, { room: msg.data.room }, {
                    leaveRoom,
                    chat
                });
                game.map = msg.data.map;
            } break;
            case "joined": {
                GameScreen(root, { room: msg.data.room }, {
                    leaveRoom,
                    chat
                });
                game.map = msg.data.map;
            } break;
            case "state": {
                if (!game.state || !rendering) {
                    game.state = msg.data;
                    startGame();
                } else {
                    game.state = msg.data;
                }
                isServerUpdated = true;

                const playerCount = document.getElementById('playerCount');
                playerCount.textContent = `Players: ${msg.data.players.length}`;

                if (activeScreen !== 1) {
                    console.error("should be unreachable");
                    ws.send(JSON.stringify({
                        type: 'return',
                        data: {}
                    }));
                }
            } break;
            case "map": {
                game.map = msg.data;
            } break;
            case "left": {
                HomeScreen(root, {}, {
                    joinRoom,
                    hostRoom
                });
                gameState = undefined;
            } break;
            case "chat": {
                appendMessage(msg.data.from, msg.data.message);
            } break;
            case "error": {
                if (!appendSystemMessage('error', msg.data.message)) {
                    // TODO: find a better way to display error messages
                    alert(msg.data.message);
                }
            } break;
            case "system": {
                if (!appendSystemMessage(msg.data.type, msg.data.msg)) {
                    console.log(msg.data.type, msg.data.msg);
                }
            } break;
            case "attack": {
                const { x, y, state } = msg.data;
                game.map.tiles[y * game.map.width + x] = state;
            } break;
            default: {
                console.error('Unknown message type:', msg.type);
            }
        }
    })
    ws.addEventListener("close", () => {
        console.log('Disconnected');
        HomeScreen(root, {}, {
            joinRoom,
            hostRoom
        });
        game.state = undefined;
        myData.id = undefined;
        myData.username = undefined
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