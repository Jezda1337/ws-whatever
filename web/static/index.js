const log = console.log;
const messageForm = document.getElementById("message-form");
const chatroomForm = document.getElementById("chatroom-form");
const ws = new WebSocket(`ws://${document.location.host}/ws`);
let currentchatroom = "";
ws.onopen = () => {
  // message form handler
  messageForm.message.focus();
  messageForm.addEventListener("submit", (e) => {
    e.preventDefault();
    if (currentchatroom !== "") {
      const messageField = messageForm.message;
      if (!messageField.value) return;

      ws.send(
        JSON.stringify({ type: "send_message", payload: messageField.value }),
      );
    }
  });

  // chat-room handler
  chatroomForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const chatroomField = chatroomForm.chatroom;

    log(chatroomField.value);
    currentchatroom = chatroomField.value;
    ws.send(
      JSON.stringify({ type: "change_room", payload: chatroomField.value }),
    );
  });
};

ws.onmessage = (e) => {
  const li = document.createElement("li");
  const data = JSON.parse(e.data);
  li.textContent = `${data.from}: ${data.payload}`;
  document.getElementById("chat").appendChild(li);
};
