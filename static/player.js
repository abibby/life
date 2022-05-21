//@ts-check
import { makeid } from "./random.js";
import Socket from "./socket.js";

const nameInput = document.getElementById("name");
const changeButtons = document.querySelectorAll(".btn-change[data-change]");

let playerID = localStorage.getItem("player-id");
if (playerID === null) {
  playerID = makeid(25);
  localStorage.setItem("player-id", playerID);
}

let name = "";

const socket = new Socket("room", playerID);
socket.onUpdate((data) => {
  const player = data.players.find((p) => p.id === playerID);
  if (player !== undefined) {
    name = player.name;
    nameInput.value = name;
  }
  document.getElementById("content").innerText = JSON.stringify(
    data,
    undefined,
    "    "
  );
});

for (const changeButton of changeButtons) {
  changeButton.addEventListener("click", (e) => {
    socket.updateLife(Number(changeButton.dataset.change));
  });
}

nameInput.addEventListener("input", () => {
  name = nameInput.value;
  socket.setName(name);
});
