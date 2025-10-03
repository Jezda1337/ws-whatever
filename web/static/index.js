let ws = null;
let currentRoomID = null;
let currentUserID = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;

const rooms = new Map();

const loginContainer = document.getElementById("login-container");
const appContainer = document.getElementById("app-container");
const loginForm = document.getElementById("login-form");
const userIdInput = document.getElementById("user-id-input");
const statusDiv = document.getElementById("status");
const userAvatar = document.getElementById("user-avatar");
const userName = document.getElementById("user-name");
const roomsList = document.getElementById("rooms-list");
const chatMessages = document.getElementById("chat-messages");
const chatHeader = document.getElementById("chat-header");
const chatInput = document.getElementById("chat-input");
const messageInput = document.getElementById("message-input");
const sendButton = document.getElementById("send-button");
const chatRoomName = document.getElementById("chat-room-name");
const chatRoomAvatar = document.getElementById("chat-room-avatar");
const typingIndicator = document.getElementById("typing-indicator");

loginForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const userId = parseInt(userIdInput.value, 10);

  if (isNaN(userId) || userId < 1) {
    alert("Please enter a valid user ID");
    return;
  }

  currentUserID = userId;
  userName.textContent = `User ${userId}`;
  userAvatar.textContent = `U${userId}`;

  loginContainer.style.display = "none";
  appContainer.classList.add("active");

  connect();
  loadRooms();
});

function connect() {
  ws = new WebSocket(
    `ws://${document.location.host}/ws?user_id=${currentUserID}`
  );

  ws.onopen = () => {
    console.log("WebSocket connected");
    statusDiv.textContent = "Connected";
    statusDiv.className = "status-indicator connected";
    reconnectAttempts = 0;
  };

  ws.onmessage = (e) => {
    try {
      const event = JSON.parse(e.data);
      handleEvent(event);
    } catch (err) {
      console.error("Failed to parse message:", err);
    }
  };

  ws.onerror = (err) => {
    console.error("WebSocket error:", err);
  };

  ws.onclose = () => {
    console.log("WebSocket disconnected");
    statusDiv.textContent = "Disconnected";
    statusDiv.className = "status-indicator disconnected";

    messageInput.disabled = true;
    sendButton.disabled = true;

    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
      console.log(
        `Reconnecting in ${delay}ms... (attempt ${reconnectAttempts})`
      );
      setTimeout(connect, delay);
    }
  };
}

function handleEvent(event) {
  switch (event.type) {
    case "new_message":
      handleNewMessage(event.payload);
      break;
    case "history":
      handleHistory(event.payload);
      break;
    case "error":
      handleError(event.payload);
      break;
    case "typing":
      handleTyping(event.payload);
      break;
    default:
      console.log("Unknown event type:", event.type, event);
  }
}

function handleNewMessage(msg) {
  updateRoomPreview(msg.room_id, msg.content, msg.created_at);

  if (currentRoomID !== msg.room_id) return;

  const messageDiv = document.createElement("div");
  messageDiv.className = `message ${msg.sender_id === currentUserID ? "own" : "other"}`;

  const bubble = document.createElement("div");
  bubble.className = "message-bubble";

  if (msg.sender_id !== currentUserID) {
    const sender = document.createElement("div");
    sender.className = "message-sender";
    sender.textContent = `User ${msg.sender_id}`;
    bubble.appendChild(sender);
  }

  const content = document.createElement("div");
  content.className = "message-content";
  content.textContent = msg.content;

  const time = document.createElement("div");
  time.className = "message-time";
  time.textContent = new Date(msg.created_at).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  bubble.appendChild(content);
  bubble.appendChild(time);
  messageDiv.appendChild(bubble);

  const emptyState = chatMessages.querySelector(".empty-state");
  if (emptyState) {
    emptyState.remove();
  }

  chatMessages.appendChild(messageDiv);
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

function handleHistory(payload) {
  chatMessages.innerHTML = "";

  if (payload.messages.length === 0) {
    const systemMsg = document.createElement("div");
    systemMsg.className = "system-message";
    systemMsg.textContent = "No messages yet. Start the conversation!";
    chatMessages.appendChild(systemMsg);
    return;
  }

  payload.messages.forEach((msg) => handleNewMessage(msg));
}

function handleError(payload) {
  const errorDiv = document.createElement("div");
  errorDiv.className = "error-message";
  errorDiv.textContent = `Error: ${payload.message}`;
  chatMessages.appendChild(errorDiv);
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

function handleTyping(payload) {
  if (!payload.user_ids || payload.user_ids.length === 0) {
    typingIndicator.classList.remove("active");
    return;
  }

  const otherUsers = payload.user_ids.filter((id) => id !== currentUserID);

  if (otherUsers.length > 0) {
    typingIndicator.textContent = `User(s) ${otherUsers.join(", ")} typing...`;
    typingIndicator.classList.add("active");
  } else {
    typingIndicator.classList.remove("active");
  }
}

function sendEvent(type, payload) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type, payload }));
  } else {
    console.error("WebSocket is not connected");
  }
}

async function loadRooms() {
  try {
    const response = await fetch(`/users/rooms?user_id=${currentUserID}`);
    if (!response.ok) {
      console.error("Failed to load rooms");
      return;
    }

    const userRooms = await response.json();
    
    userRooms.forEach((room) => {
      const roomData = {
        id: room.id,
        name: room.name || `Room ${room.id}`,
        lastMessage: "No messages yet",
        time: new Date(room.created_at),
      };
      rooms.set(room.id, roomData);
      renderRoomItem(roomData);
    });
  } catch (error) {
    console.error("Error loading rooms:", error);
  }
}

function renderRoomItem(room) {
  let roomItem = document.getElementById(`room-${room.id}`);

  if (!roomItem) {
    roomItem = document.createElement("div");
    roomItem.className = "room-item";
    roomItem.id = `room-${room.id}`;
    roomItem.addEventListener("click", () => selectRoom(room.id));
    roomsList.appendChild(roomItem);
  }

  roomItem.innerHTML = `
    <div class="room-item-header">
      <div class="room-name">${room.name}</div>
      <div class="room-time">${formatTime(room.time)}</div>
    </div>
    <div class="room-preview">${room.lastMessage}</div>
  `;
}

function selectRoom(roomId) {
  if (currentRoomID === roomId) return;

  document
    .querySelectorAll(".room-item")
    .forEach((item) => item.classList.remove("active"));
  document.getElementById(`room-${roomId}`).classList.add("active");

  currentRoomID = roomId;
  const room = rooms.get(roomId);

  chatHeader.style.display = "flex";
  chatInput.style.display = "flex";
  chatRoomName.textContent = room.name;
  chatRoomAvatar.textContent = `R${roomId}`;

  messageInput.disabled = false;
  sendButton.disabled = false;

  sendEvent("join_room", { room_id: roomId });
}

function updateRoomPreview(roomId, message, time) {
  const room = rooms.get(roomId);
  if (!room) return;

  room.lastMessage = message;
  room.time = new Date(time);
  renderRoomItem(room);
}

function formatTime(date) {
  const now = new Date();
  const messageDate = new Date(date);
  const diffDays = Math.floor(
    (now - messageDate) / (1000 * 60 * 60 * 24)
  );

  if (diffDays === 0) {
    return messageDate.toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    });
  } else if (diffDays === 1) {
    return "Yesterday";
  } else if (diffDays < 7) {
    return messageDate.toLocaleDateString([], { weekday: "short" });
  } else {
    return messageDate.toLocaleDateString([], {
      month: "short",
      day: "numeric",
    });
  }
}

sendButton.addEventListener("click", (e) => {
  e.preventDefault();

  if (!currentRoomID) {
    alert("Please select a room first");
    return;
  }

  const content = messageInput.value.trim();
  if (!content) return;

  sendEvent("send_message", { content });
  messageInput.value = "";
});

messageInput.addEventListener("keypress", (e) => {
  if (e.key === "Enter") {
    sendButton.click();
  }
});

let typingTimeout;
messageInput.addEventListener("input", () => {
  if (!currentRoomID) return;

  clearTimeout(typingTimeout);
  sendEvent("typing", {});

  typingTimeout = setTimeout(() => {}, 3000);
});
