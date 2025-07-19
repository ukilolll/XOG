package service

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type queue struct {
	Queue      map[*subscription]bool
	Mu         sync.Mutex
	Register   chan *subscription
	unregister chan *subscription
}

func NewQueue() *queue {
	return &queue{
		Queue:      make(map[*subscription]bool),
		Mu:         sync.Mutex{},
		Register:   make(chan *subscription),
		unregister: make(chan *subscription),
	}
}


func (q *queue) RunQueue() {
	log.Println("queue system running")
	log.Println(q.unregister)
	for {
		select {
		case sub := <-q.Register:
			conn := sub.conn.ws
			q.Mu.Lock()
			q.Queue[sub] = true
			queueLen := len(q.Queue)
			log.Printf("client in queue %p\n", conn)
			log.Printf("queue now: %v\n", queueLen)
			q.Mu.Unlock()

			if queueLen >= 2 {
				// Get all connections and channels
				var subs []*subscription
				for conn := range q.Queue {
					subs = append(subs, conn)
				}
				q.Mu.Lock()
				// Clear the queue BEFORE closing contxt
				q.Queue = make(map[*subscription]bool)
				q.Mu.Unlock()

				// Send START message to all clients
				for _, s := range subs {
					err := s.conn.ws.WriteJSON(map[string]any{
						"type": START,
						"data": "",
					})
					if err != nil {
						log.Printf("Error sending START to %p: %v", conn, err)
					}
				}

				// Start the game
				id, _ := uuid.NewUUID()
				log.Printf("Starting game with %d players\n", len(subs))
				gameslayer(subs, id.String())

			} else {
				// Send WAIT message
				err := conn.WriteJSON(map[string]any{
					"type": "WAIT",
					"data": "",
				})
				if err != nil {
					log.Printf("Error sending WAIT to %p: %v", conn, err)
					// If we can't send WAIT, remove from queue
					delete(q.Queue, sub)
				}
			}

		case conn := <-q.unregister:
			q.Mu.Lock()
			delete(q.Queue, conn)
			q.Mu.Unlock()

		}
	}
}

// addQueue is the entry point for a new connection.
func (q *queue) addQueue(conn *ws.Conn) {
	log.Printf("client in ready %p\n", conn)
	//for handle use reload page
	time.Sleep(100 * time.Millisecond)
	//send = ตัวรับจาก hub ที่ส่งมาหา pump
	c := &connection{send: make(chan msg), ws: conn}
	//unregister and gameInput pump ตัวส่งส่งไปหา hub
	s := &subscription{conn: c, unregister: q.unregister}
	go s.writePump() //server read from client
	go s.readPump()  //server write from clinet
	q.Register <- s
}
