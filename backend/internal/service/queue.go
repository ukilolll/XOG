package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type login struct{
	Conn *ws.Conn
	Done chan struct{}
}

func NewQueue() *queue{
	return &queue{
		Queue: make(map[*ws.Conn]chan struct{}) ,
		Mu: sync.Mutex{} , 

		Register: make(chan *login),
		unregister: make(chan *login),
	}
}

type queue struct {
	Queue map[*ws.Conn]chan struct{}
	Mu sync.Mutex

	Register chan *login
	unregister chan *login
}

func(q *queue)ready(conn *ws.Conn){
	done := make(chan struct{})
	q.Register <- &login{Conn: conn,Done: done}
	fmt.Printf("client in ready %p\n",conn)

	pingTicker := time.NewTicker(1 * time.Second)
	defer func(){
		fmt.Printf("end queueing  %p\n",conn)
		pingTicker.Stop()
	}()
	// Start goroutine to send pings
	for {
	select {
		
	case <-pingTicker.C:
		if err := conn.WriteMessage(ws.TextMessage, []byte("ping")); err != nil {
			q.unregister <- &login{Conn: conn,Done: done}
			return
		}
	case _,ok := <-done:
		if !ok {return}
				
	}
	}
}

func(q *queue) RunQueue(){
	log.Println("queue system running")
	for{
	select{
	case l := <-q.Register:
		q.Queue[l.Conn] = l.Done
		fmt.Printf("client in queue %p\n",l.Conn)
		fmt.Printf("queue now:%v\n",len(q.Queue))
		
		var conns []*ws.Conn
		if len(q.Queue) >= 2 {
		// go runGame
		for conn,ch := range q.Queue{
			conn.WriteJSON(map[string]any{
				"type":START,
				"data":"",
			})
			conns = append(conns, conn)
			close(ch)
		}	

		id,_ := uuid.NewUUID()
		gameslayer(conns,id.String())
			
		q.Queue = make(map[*ws.Conn]chan struct{})

		}else{
			l.Conn.WriteJSON(map[string]any{
					"type":Wait,
					"data":"",
			})
		}
		
	case l := <- q.unregister:
		delete(q.Queue,l.Conn)
		close(l.Done)
	}
	}
}

func(q *queue)addQueue(conn *ws.Conn){
	go q.ready(conn)
}

// //empty map
// q.Queue = make(map[*ws.Conn]bool)