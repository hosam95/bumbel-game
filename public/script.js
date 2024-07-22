const playerSpeed = 10;

const game = { state: undefined, ctx: undefined };
let activeScreen = 0; // 0: Home, 1: Game
let lastTimestamp = 0;
let rendering = false;
const myData = {
    id: undefined,
    username: undefined,
}
let isServerUpdated = false;

function HomeScreen(root, _data, handlers) {
    activeScreen = 0;

    const center = document.createElement('div');
    center.className = 'center';

    const h1 = document.createElement('h1');
    h1.textContent = 'Online Game';
    center.appendChild(h1);

    const roomBox = document.createElement('div');
    roomBox.className = 'row';
    center.appendChild(roomBox);

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

    const dashboard = document.createElement('div');
    dashboard.classList.add('dashboard', "row");

    const roomId = document.createElement('span');
    roomId.textContent = `Room ID: ${data.room}`;
    dashboard.appendChild(roomId);

    const playerCount = document.createElement('span');
    playerCount.id = 'playerCount';
    playerCount.textContent = `Players: ${data?.players?.length ?? 0}`;
    dashboard.appendChild(playerCount);

    const leaveBtn = document.createElement('button');
    leaveBtn.classList.add("btn", "danger");
    leaveBtn.textContent = 'Leave Room';
    dashboard.appendChild(leaveBtn);

    const container = document.createElement('div');
    container.classList.add("game-container");

    const canvas = document.createElement('canvas');
    canvas.id = 'canvas';
    canvas.width = 800;
    canvas.height = 600;
    const ctx = canvas.getContext('2d');
    if (!ctx) throw new Error('Failed to get 2d context');
    game.ctx = ctx;
    container.appendChild(canvas);

    const chat = document.createElement('div');
    chat.classList.add("chat");
    container.appendChild(chat);

    const chatTitle = document.createElement('h2');
    chatTitle.classList.add("chat-title");
    chatTitle.textContent = 'Chat';
    chat.appendChild(chatTitle);

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

    root.replaceChildren(dashboard, container);

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

function draw(ts) {
    if (!game.state) { console.log("no game state"); rendering = false; return; }
    if (!game.ctx) { console.log("no game ctx"); rendering = false; return; }
    const dt = (ts - lastTimestamp) / 1000;
    lastTimestamp = ts;
    const { width, height } = game.ctx.canvas;
    
    const gameState = game.state;
    const ctx = game.ctx;

    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);

    // Update
    if (!isServerUpdated) {
        for (const player of gameState.players) {
            player.x += player.vx * dt * playerSpeed;
            player.y += player.vy * dt * playerSpeed;
        }
    } else {
        isServerUpdated = false;
    }

    // Render
    ctx.fillStyle = "#FFD35A";
    ctx.fillRect(0, 0, width, height);

    const { width: mapWidth, height: mapHeight } = gameState.state.map;
    const cellWidth = width / mapWidth;
    const cellHeight = height / mapHeight;

    for (let i = 0; i < mapWidth * mapHeight; i++) {
        const x = i % mapWidth;
        const y = Math.floor(i / mapWidth);

        switch (gameState.state.map.tiles[i]) {
            case 0: {
                // Empty
            } break;
            case 1: {
                ctx.fillStyle = "#" + gameState.state.teamA.toString(16).padStart(6, '0');
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
            case 2: {
                ctx.fillStyle = "#" + gameState.state.teamB.toString(16).padStart(6, '0');
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
            case 3: {
                ctx.fillStyle = "#FFA823";
                ctx.fillRect(x * cellWidth, y * cellHeight, cellWidth, cellHeight);
            } break;
        }
    }

    if (gameState.state.phase === 1) {
        for (let i = 0; i < gameState.players.length; i++) {
            const player = gameState.players[i];
            if (player.user.id === myData.id) {
                ctx.fillStyle = "#ffffff";
                ctx.fillRect(player.x * cellWidth, player.y * cellHeight, cellWidth, cellHeight);
                ctx.fillStyle = "#" + gameState.state[i === 0 ? "teamA" : "teamB"].toString(16).padStart(6, '0');
                ctx.fillRect(player.x * cellWidth + 0.1 * cellWidth, player.y * cellHeight + 0.1 * cellHeight, cellWidth * 0.8, cellHeight * 0.8);
            } else {
                ctx.fillStyle = "#" + gameState.state[i === 0 ? "teamA" : "teamB"].toString(16).padStart(6, '0');
                ctx.fillRect(player.x * cellWidth, player.y * cellHeight, cellWidth, cellHeight);
            }
        }
    }

    requestAnimationFrame(draw);
}

requestAnimationFrame(draw);

function appendMessage(from, message) {
    const chatBox = document.getElementById('chatBox');

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
        switch (e.key) {
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
            case " ": {
                ws.send(JSON.stringify({
                    type: 'action',
                    data: { action: "start" }
                }));
            } break;
        }
    })

    window.addEventListener("keyup", (e) => {
        switch (e.key) {
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
            draw(timestamp);
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
            } break;
            case "joined": {
                GameScreen(root, { room: msg.data.room }, {
                    leaveRoom,
                    chat
                });
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
                console.error('Error:', msg.data);
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