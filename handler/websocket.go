package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"web-socket/request"

	"github.com/gorilla/websocket"
)

// Map untuk menyimpan koneksi WebSocket per pengguna.
var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

// Upgrader digunakan untuk meng-upgrade koneksi HTTP ke WebSocket.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Izinkan semua asal (origin) untuk koneksi WebSocket.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler memeriksa apakah permintaan adalah permintaan WebSocket dan mengizinkan upgrade.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	handleConnections(conn)
}

func handleConnections(conn *websocket.Conn) {
	var username string

	defer func() {
		mu.Lock()
		// Hapus koneksi dari map saat pengguna terputus.
		delete(clients, username)
		mu.Unlock()
		 // Tutup koneksi WebSocket.
		conn.Close()
		log.Printf("%s disconnected", username)
	}()

	// Baca pesan pertama untuk mendapatkan username pengguna.
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Read:", err)
		return
	}

	var input request.Input
	if err := json.Unmarshal(message, &input); err != nil {
		log.Println("Unmarshal:", err)
		return
	}

	username = input.Username
	log.Printf("%s connected", username)

	// Simpan koneksi dalam map dengan username sebagai kunci.
	mu.Lock()
	clients[username] = conn
	mu.Unlock()

	for {
		// Baca pesan dari koneksi WebSocket.
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read:", err)
			break
		}

		var input request.Input
		if err := json.Unmarshal(message, &input); err != nil {
			log.Println("Unmarshal:", err)
			continue
		}

		log.Printf("Received message from %s: %s", input.Username, input.Message)

		// Kirim pesan ke pengguna lain yang dituju.
		mu.Lock()
		// Periksa apakah pengguna tujuan terhubung.
		if recipientConn, ok := clients[input.Username]; ok {
			// Kirim pesan ke pengguna tujuan.
			if err := recipientConn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Write:", err)
			}
		} else {
			log.Printf("User %s not connected", input.Username)
		}
		mu.Unlock()
	}
}