package service

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"sync"
)

// tip
// 1. uses rooms map[string]map[uuid]*connection for not anonymous or map[string]map[*connection]bool for anonymous
// 2. send websocket.conn and uuid to chanel for identify in websocket
// 3. Hubเก็บ channel ของทุกลูกค่ายสำคัญมาก เพื่อให้ hub ส่งและ close channel ไปยังลูกค้ายถูก
// 4. in handler create subscription , send register<- and create readpub() writepub()
type playerData struct {
	*subscription
	Role string
}

type gameResource struct {
	Users map[string]*playerData
	Board [3][3]string
	Turn  string

	gameInput  chan payload
	unregister chan *subscription
	confirmSend chan struct{}
}

func (g *gameResource) runGame() {
	log.Println("game start")
	player1 := g.Users["player1"]
	player2 := g.Users["player2"]

	temp1 := rand.Intn(2)
	temp2 := rand.Intn(2)
	if temp1 == 0 {
		g.Turn = "o"
	} else {
		g.Turn = "x"
	}
	if temp2 == 0 {
		player1.Role = "o"
		player2.Role = "x"
	} else {
		player1.Role = "x"
		player2.Role = "o"
	}

	sendPlayer1 := player1.subscription.conn.send
	sendplayer2 := player2.subscription.conn.send
	connPlayer1 := player1.subscription.conn.ws
	connPlayer2 := player2.subscription.conn.ws

	sendPlayer1 <- msg{Category: ROLE, Data: player1.Role}
	sendplayer2 <- msg{Category: ROLE, Data: player2.Role}

	sendPlayer1 <- msg{Category: TURN, Data: g.Turn}
	sendplayer2 <- msg{Category: TURN, Data: g.Turn}

	// add read Deadline to start player
	if player1.Role == g.Turn{
		connPlayer1.SetReadDeadline(time.Now().Add(pongWait))
	}else{
		connPlayer2.SetReadDeadline(time.Now().Add(pongWait))
	}

	for {
	select {
		case i := <-g.gameInput:

			//check turn
			playerRole :=  g.Users[i.Player].Role
			log.Println("play by:",i.Player,playerRole,i.Msg)
			if !(playerRole == g.Turn) {
				continue
			}

			//run board
			if i.Msg.Category == PLAY {
				//fmt.Println(reflect.TypeOf(i.Msg.Data))
				temp, ok := i.Msg.Data.(float64)
				if !ok {
					continue
				}
				data := int(temp)
				row := data / 3
				col := data % 3

				if !(g.Board[row][col] == "") {
					continue
				}

				if ok := try(func() {
					g.Board[row][col] = playerRole
				});!ok {continue}

				sendPlayer1 <- msg{Category: PLAY, Data: map[string]any{"value": playerRole , "index": row*3+col}}
				sendplayer2 <- msg{Category: PLAY, Data: map[string]any{"value": playerRole , "index": row*3+col}}
				
				result := checkWin(g.Board)
				// if game end
				if result != "" {
					//create WaitGroup for confirm message already sended before end thread
					var wg sync.WaitGroup
					for _,v := range g.Users{
						wg.Add(1)
						go func (confirmSend chan struct{}){
							for range  confirmSend{
								//gone message from writePump() and Done thread
								wg.Done()
								return
							}
						}(v.subscription.conn.confirmSend)
					}
					//send message 
					sendPlayer1 <- msg{Category: END, Data: result}
					sendplayer2 <- msg{Category: END, Data: result}
					//wait for confirm
					wg.Wait()

					for _, v := range g.Users {
						close(v.subscription.conn.send)
					}
					
					return
				}
				// stop readline current user and add deadline for another user
				test := g.Users[i.Player].conn.ws
				test.SetReadDeadline(time.Time{})
				g.Turn= ChangeTurn(playerRole)

				if player1.Role == g.Turn{
					connPlayer1.SetReadDeadline(time.Now().Add(pongWait))
				}else{
					connPlayer2.SetReadDeadline(time.Now().Add(pongWait))
				}

				sendPlayer1 <- msg{Category: TURN, Data: g.Turn}
				sendplayer2 <- msg{Category: TURN, Data: g.Turn}

			}

		case s := <-g.unregister:
			role :=  g.Users[s.name].Role 
			sendPlayer1 <- msg{Category: DISCONNECT, Data: role}
			sendplayer2 <- msg{Category: DISCONNECT, Data: role}
			for _, v := range g.Users {
				try(func() {
					close(v.subscription.conn.send)
				});

			}
			return

	}
	}

}

func ServeWs(w http.ResponseWriter, r *http.Request, queue *queue) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	queue.addQueue(conn)
}

func gameslayer(subs []*subscription, roomId string) {
	log.Printf("game room:%v created\n", roomId)
	users := make(map[string]*playerData)
	//ทุก goroutine มี same channel แต่ มีตัวรับแค่ server (no boardcast on problem)
	//แต่ตัวไหน ตัวส่งไม่รู้ สุดท้ายpumpก็ต้องมี identifer อยู่ดี (userid,pointer websocket.Conn,unique name, etc)

	//for same channel hub and pump
	gameInput := make(chan payload)
	unregister := make(chan *subscription)
	confirmSend := make(chan struct{})

	for i, s := range subs {
		playername := fmt.Sprintf("player%v", i+1)
		//add data for game hub
		s.conn.confirmSend = confirmSend
		s.unregister = unregister
		s.gameInput = gameInput
		s.name = playername
		users[playername] = &playerData{subscription: s}
	}

	g := gameResource{Users: users, gameInput: gameInput, unregister: unregister,confirmSend:confirmSend}
	go g.runGame()
}

func try(callback func()) (ok bool) {
    defer func() {
        if r := recover(); r != nil {
            ok = false
        }
    }()

	callback()
	return true 
}
// game := gameResource{
// 	Users:      users,
// 	gameInput:  gameInput,
// 	unregister: unregister,
// 	InHub:      H.InHub,
// 	OutHub:     H.OutHub,
// 	Moniter:    H.Moniter,
// }

// // can create hub for moniter game
// type  regisHub struct {
// 	RoomName string //identify
// 	Send     chan msg //channel ลูกค่าย and websocket.conn(optional)
// }
// type message struct {
// 	RoomName string//identify
// 	Data     any
// }

// var H = Hub{
// 	Moniter: make(chan message),
// 	InHub:   make(chan regisHub),
// 	OutHub:  make(chan regisHub),
// 	Rooms:   make(map[string]chan msg),// server เก็บ channel ลูกค่าย
// }


