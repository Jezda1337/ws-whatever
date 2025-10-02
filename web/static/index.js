const messageForm = document.getElementById("message-form");
const roomForm = document.getElementById("room-form");
const chatList = document.getElementById("chat");
const statusDiv = document.getElementById("status");
const currentRoomSpan = document.getElementById("current-room");
const messageInput = messageForm.message;
const messageButton = messageForm.querySelector("button");

let ws = null;
let currentRoomID = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const userID =
  new URLSearchParams(window.location.search).get("user_id") || "1";

function connect() {
  ws = new WebSocket(`ws://${document.location.host}/ws?user_id=${userID}`);

  ws.onopen = () => {
    console.log("WebSocket connected");
    statusDiv.textContent = "Connected";
    statusDiv.className = "status connected";
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
    statusDiv.className = "status disconnected";

    messageInput.disabled = true;
    messageButton.disabled = true;

    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
      console.log(
        `Reconnecting in ${delay}ms... (attempt ${reconnectAttempts})`,
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
  const li = document.createElement("li");

  const meta = document.createElement("div");
  meta.className = "meta";
  meta.textContent = `User ${msg.sender_id} • ${new Date(msg.created_at).toLocaleTimeString()}`;

  const content = document.createElement("div");
  content.textContent = msg.content;

  li.appendChild(meta);
  li.appendChild(content);
  chatList.appendChild(li);
  chatList.scrollTop = chatList.scrollHeight;
}

function handleHistory(payload) {
  chatList.innerHTML = "";

  const systemMsg = document.createElement("li");
  systemMsg.className = "system";
  systemMsg.textContent = `Loaded ${payload.messages.length} messages`;
  chatList.appendChild(systemMsg);

  payload.messages.forEach((msg) => handleNewMessage(msg));
}

function handleError(payload) {
  const li = document.createElement("li");
  li.className = "error";
  li.textContent = `Error: ${payload.message}`;
  chatList.appendChild(li);
  chatList.scrollTop = chatList.scrollHeight;
}

function handleTyping(payload) {
  const typingDiv = document.getElementById("typing-indicator");
  if (!typingDiv) return;

  const otherUsers = payload.user_ids.filter((id) => id !== parseInt(userID));

  if (otherUsers.length > 0) {
    typingDiv.textContent = `User(s) ${otherUsers.join(", ")} typing...`;
    typingDiv.style.display = "block";
  } else {
    typingDiv.style.display = "none";
  }
}

function sendEvent(type, payload) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type, payload }));
  } else {
    console.error("WebSocket is not connected");
  }
}

roomForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const roomID = parseInt(roomForm.roomId.value, 10);

  if (isNaN(roomID)) {
    alert("Please enter a valid room ID");
    return;
  }

  sendEvent("join_room", { room_id: roomID });
  currentRoomID = roomID;
  currentRoomSpan.textContent = roomID;

  messageInput.disabled = false;
  messageButton.disabled = false;
  messageInput.focus();
});

messageForm.addEventListener("submit", (e) => {
  e.preventDefault();

  if (!currentRoomID) {
    alert("Please join a room first");
    return;
  }

  const content = messageInput.value.trim();
  if (!content) return;

  sendEvent("send_message", { content });
  messageInput.value = "";
});

let typingTimeout;
messageInput.addEventListener("input", () => {
  if (!currentRoomID) return;

  clearTimeout(typingTimeout);
  sendEvent("typing", {});

  typingTimeout = setTimeout(() => {}, 3000);
});

connect();
