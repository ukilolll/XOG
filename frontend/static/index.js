const socket = new WebSocket("ws://localhost:80/apiv1");
const board = document.getElementById("board");
const statusEl = document.getElementById("status");
const role = document.getElementById("role");
const resultId = document.getElementById("result");
const divbutton = document.getElementById("divbutton")

let myRole = ""; // "x" or "o"
let myTurn = false;

const cells = [];

// Build board
for (let i = 0; i < 9; i++) {
const cell = document.createElement("div");
cell.classList.add("cell");
cell.dataset.index = i;

cell.addEventListener("click", () => {
    if (myTurn && cell.textContent === "") {
    const index = parseInt(cell.dataset.index);
    const data = JSON.stringify({
        type: "play",
        data: index
    })
    socket.send(data);
    console.log("send to server:",data)
    myTurn = false;
    }
});

board.appendChild(cell);
cells.push(cell);
}

socket.onopen = () => {
    console.log("server connect!!!")
    statusEl.textContent = "Waiting for another player...";
};

socket.onmessage = (event) => {
    console.log("Server:", event.data);
    let msg = {}
    try{
        msg = JSON.parse(event.data);
    }catch(err){
        console.log("server ping")
    }

    switch (msg.type) {
    case "wait":
        statusEl.textContent = "Waiting for another player...";
        break;

    case "start":
        statusEl.textContent = "Game started!";
        break;

    case "role":
        myRole = msg.data; // "x" or "o"
        role.textContent = `You are ${myRole.toUpperCase()}`;
        break;

    case "turn":
        const turn = msg.data; // "x" or "o"
        myTurn = (turn === myRole);
        statusEl.textContent = myTurn ? "Your turn!" : "Opponent's turn";
        break;

    case "play":
        const index = msg.data.index;  // 0-8
        const value = msg.data.value;  // "x" or "o"
        cells[index].textContent = value;
        break;

    case "disconnect":
        let data = msg.data
        if (data === myRole){
            resultId.textContent = "Time Out disconnected."
        }else{
            resultId.textContent = "You win!!! Opponent disconnect."
        }
        divbutton.innerHTML =  `<button onclick="location.reload()">play again</button>`
        break;

    case "end":
        let result = msg.data;
        console.log(result,"draw")
        if (result === "draw") {
            resultId.textContent = "It's a draw!";
        }
        else {
            resultId.textContent = result === myRole ? "You win!" : "You lose!";
        }
        myTurn = false;
        divbutton.innerHTML =  `<button onclick="location.reload()">play again</button>`
        break;
    }
};

socket.onclose = (event) => {
    console.log("Code:", event.code);     // e.g., 1000 = normal, 1006 = abnormal
    console.log("Reason:", event.reason); 
    statusEl.textContent = "Connection closed";
};

socket.onerror = (err) => {
    statusEl.textContent = "WebSocket error.";
    console.error("WebSocket error:", err);
};