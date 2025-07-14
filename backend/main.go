package main

import (
	"fmt"
	"github/ukilolll/revenge/internal/service"
	"log"
	"net/http"
)

const(
	PORT = 8080
)

func main() {
	//can loop new queue for multiqueue system (like new queue() in java)
	queue := service.NewQueue()
	go queue.RunQueue()

	r := http.NewServeMux()
	
	r.HandleFunc("/apiv1", func(w http.ResponseWriter, r *http.Request) {
		service.ServeWs(w, r, queue)
	})

	log.Printf("server run at port:%v\n",PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%v",PORT),r)
	if err != nil {
		log.Panic(err)
	}
}
