package main

import (
	"fmt"
	"log"
	"net/http"
)

const(
	PORT = 80
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r,"./index.html")
	})

	// Serve static files like index.js from /static/ URL path
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("server run at port ",PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%v",PORT),mux)
	if err != nil {
		log.Panic(err)
	}
}