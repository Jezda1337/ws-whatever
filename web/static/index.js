const log = console.log;
const form = document.forms[0];
const ws = new WebSocket(`ws://${document.location.host}/ws`);
ws.onopen = (e) => {
  log(e);
  form.message.focus();
  form.addEventListener("submit", (e) => {
    e.preventDefault();
    const messageField = form.message;
    if (!messageField.value) return;

    ws.send(
      JSON.stringify({ type: "send_message", payload: messageField.value }),
    );
  });
};

ws.onmessage = (e) => {
  const li = document.createElement("li");
  const data = JSON.parse(e.data);
  li.textContent = `${data.from}: ${data.payload}`;
  document.getElementById("chat").appendChild(li);
};
