package main

import (
	"log"
	"net/http"
	"web-socket/handler"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WebSocket server is running!"))
	})

	http.HandleFunc("/ws", handler.WebSocketHandler)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}