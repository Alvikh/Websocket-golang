package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsChan adalah channel yang digunakan untuk mengirimkan payload Ws antara klien dan server WebSocket
var wsChan = make(chan WsPayload)

// clients menyimpan semua klien WebSocket yang terhubung beserta username mereka
var clients = make(map[WebSocketConnection]string)

// upgradeConnection adalah websocket upgrader dari package gorilla/websocket
// Ini digunakan untuk meng-upgrade koneksi HTTP menjadi koneksi WebSocket
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,                                       // Buffer baca WebSocket
	WriteBufferSize: 1024,                                       // Buffer tulis WebSocket
	CheckOrigin:     func(r *http.Request) bool { return true }, // Mengizinkan koneksi dari semua origin
}

// WebSocketConnection adalah struct pembungkus untuk koneksi WebSocket dari gorilla/websocket
type WebSocketConnection struct {
	*websocket.Conn
}

// WsJsonResponse mendefinisikan struktur respons yang dikirim kembali ke klien WebSocket
type WsJsonResponse struct {
	Action         string   `json:"action"`          // Tindakan yang dilakukan
	Message        string   `json:"message"`         // Konten pesan
	MessageType    string   `json:"message_type"`    // Jenis pesan (informasi, error, dll.)
	ConnectedUsers []string `json:"connected_users"` // Daftar pengguna yang terhubung saat ini
}

// WsPayload mendefinisikan struktur pesan yang diterima dari klien WebSocket
type WsPayload struct {
	Action   string              `json:"action"`   // Tindakan yang harus dilakukan
	Username string              `json:"username"` // Username dari klien
	Message  string              `json:"message"`  // Konten pesan
	Conn     WebSocketConnection `json:"-"`        // Koneksi WebSocket (tidak diserialisasi dalam JSON)
}

// WsEndpoint meng-upgrade koneksi HTTP menjadi koneksi WebSocket
func WsEndpoint(c *gin.Context) {
	// Meng-upgrade koneksi HTTP menjadi WebSocket
	ws, err := upgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Klien terhubung ke endpoint")

	// Mengirim pesan koneksi awal ke klien
	var response WsJsonResponse
	response.Message = `<em><small>Terhubung ke server</small></em>`

	// Mendaftarkan koneksi klien baru
	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	// Mengirim respons awal ke klien
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	// Memulai goroutine untuk mendengarkan pesan dari klien
	go ListenForWs(&conn)
}

// ListenForWs menangani pesan WebSocket yang masuk dari klien
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		// Membaca data JSON dari koneksi WebSocket
		err := conn.ReadJSON(&payload)
		if err != nil {
			// Jika terjadi kesalahan, tidak melakukan apapun
		} else {
			// Mengirim pesan ke channel wsChan
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// ListenToWsChannel memproses pesan yang diterima di wsChan
func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		// Menunggu pesan masuk di channel wsChan
		e := <-wsChan

		switch e.Action {
		case "username":
			// Pengguna baru telah memberikan username, menyimpannya dan mengirimkan daftar pengguna yang terhubung
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "left":
			// Seorang pengguna keluar, menghapus mereka dari peta clients dan mengirimkan daftar pengguna yang diperbarui
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "broadcast":
			// Mengirimkan pesan dari seorang pengguna ke semua klien yang terhubung
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)
		}
	}
}

// getUserList mengembalikan daftar semua username yang terhubung saat ini
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList) // Mengurutkan username secara alfabet
	return userList
}

// broadcastToAll mengirimkan pesan ke semua klien WebSocket yang terhubung
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			// Jika klien terputus, menghapus mereka dari peta clients
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}
