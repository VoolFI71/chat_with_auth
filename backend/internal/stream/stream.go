package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	peers      = make(map[*websocket.Conn]*webrtc.PeerConnection)
	peersMutex sync.Mutex
)

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	pc, err := createPeerConnection(conn)
	if err != nil {
		fmt.Println("Error creating PeerConnection:", err)
		return
	}

	peersMutex.Lock()
	peers[conn] = pc
	peersMutex.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		// Обработка сообщений от клиента (например, ICE кандидаты и SDP)
		handleMessage(msg, conn)
	}
}

func createPeerConnection(conn *websocket.Conn) (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Отправьте кандидата обратно клиенту
			conn.WriteJSON(candidate.ToJSON())
		}
	})

	return pc, nil
}

func handleMessage(msg []byte, conn *websocket.Conn) {
	var message map[string]interface{}
	if err := json.Unmarshal(msg, &message); err != nil {
		fmt.Println("Error unmarshalling message:", err)
		return
	}

	pc, ok := peers[conn]
	if !ok {
		fmt.Println("PeerConnection not found for the connection")
		return
	}

	switch message["type"] {
	case "offer":
		// Обработка SDP-предложения
		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  message["sdp"].(string),
		}
		pc.SetRemoteDescription(offer)
		answer, err := pc.CreateAnswer(nil) // Передаем nil в качестве аргумента
		if err == nil {
			pc.SetLocalDescription(answer)
			conn.WriteJSON(answer)
		}
	case "candidate":
		// Обработка ICE кандидатов
		candidate := webrtc.ICECandidateInit{
			SDPMid:        stringPtr(message["mid"].(string)), // Создаем указатель на строку
			SDPMLineIndex: uint16Ptr(uint16(message["index"].(float64))), // Преобразуем в uint16 и создаем указатель
			Candidate:     message["candidate"].(string),
		}
		pc.AddICECandidate(candidate)
	}
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}

// Вспомогательная функция для создания указателя на uint16
func uint16Ptr(i uint16) *uint16 {
	return &i
}
