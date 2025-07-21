package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	ws "github.com/gorilla/websocket"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

const (
	START category = "start"
	TURN category = "turn"
	ROLE category = "role"
	PLAY category = "play"
	END category = "end"
	PREPAREDISCONNECT category = "preparedisconnect"
	DISCONNECT = "disconnect"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type category string

type msg struct{
	Category category `json:"type"`
	Data any  `json:"data"`
}

type payload struct {
	Player string `json:"name"`
	Msg msg `json:"data"`
}
//chanel with websocket connect 
type connection struct {
	// The websocket connection.
	ws *ws.Conn
	// important ให้ hub ส่งกับไปยังลูกค้ายถูก
	send chan msg
	confirmSend chan struct{}
}
func (c *connection) write(mt int, payload msg) error {
	data,err := json.Marshal(payload)
	if err != nil {
		log.Panic(err)
	}
	return c.ws.WriteMessage(mt,data)
}


type subscription struct {
	conn *connection
	unregister chan *subscription
	gameInput chan payload
	name string // identify
}
//read = server read
//loop for read data for client 
//and make message struct and send to Hub
func (s *subscription) readPump() {
	conn := s.conn.ws
	log.Println("create readPump",s.conn)
	defer func() {
		s.unregister <- s
		conn.WriteMessage(ws.CloseNormalClosure,[]byte("end"))
		log.Println("close readPump",s.conn)
	}()

	conn.SetReadLimit(maxMessageSize)

	var msg msg
	var syntaxErr *json.SyntaxError
	var unmarshalErr *json.UnmarshalTypeError
	for {
		err := conn.ReadJSON(&msg)
		if err != nil { 
			//if error just log 
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
	
			// Check if it's a JSON decode error (bad format)
			if errors.As(err, &syntaxErr) || errors.As(err, &unmarshalErr) {
				continue // skip this iteration, don't close connection
			}

			break //do not ues return 
		}
		s.gameInput <- payload{Player: s.name, Msg: msg}
	}
}
//write = server write
//loop for Hub send data 
// and send to client
func (s *subscription) writePump() {
	conn := s.conn
	log.Println("create writePump",s.conn)
	defer func() {
		conn.ws.WriteMessage(ws.CloseNormalClosure,[]byte("end"))
		log.Println("close writePump",s.conn)
	}()



	for msg := range conn.send {
		log.Printf("Sending to %s: %v", s.name, msg)
		//if client close error
		if err := conn.write(ws.TextMessage, msg); err != nil {
			log.Printf("write error: %v", err)
			break  // แก้เป็น break แทน return เพื่อให้ defer ทำงานชัวร์
		}

		if msg.Category == END{
			s.conn.confirmSend <- struct{}{}
		}

	}

}