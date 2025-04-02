package wsk

import (
	"backend/pkg/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketChat struct {
	Users          map[string]*UserChat
	JoinChannel    userChannel
	LeaveChannel   userChannel
	MessageChannel messageChannel
	MessageHistory map[string][]*models.Message
	Mu             sync.Mutex
}

func NewWebsocketChat() *WebsocketChat {
	w := &WebsocketChat{
		Users:          make(map[string]*UserChat),
		JoinChannel:    make(userChannel),
		LeaveChannel:   make(userChannel),
		MessageChannel: make(messageChannel),
		MessageHistory: make(map[string][]*models.Message),
	}
	go w.UsersChatManager()
	return w
}

func (w *WebsocketChat) UsersChatManager() {
	for {
		select {
		case user := <-w.JoinChannel:
			w.Mu.Lock()
			w.Users[user.Username] = user
			w.sendHistory(user)
			w.Mu.Unlock()
			log.Printf("User %s joined the chat", user.Username)

		case user := <-w.LeaveChannel:
			w.Mu.Lock()
			delete(w.Users, user.Username)
			w.Mu.Unlock()
			log.Printf("User %s left the chat", user.Username)

		case msg := <-w.MessageChannel:
			w.Mu.Lock()

			if msg.Type == "ping" {
				w.Mu.Unlock()
				continue
			}
			switch msg.Type {

			case "typing":

				if targetUser, ok := w.Users[msg.TargetUsername]; ok {
					if err := targetUser.Connection.WriteJSON(msg); err != nil {
						log.Printf("Error sending typing status to %s: %v", targetUser.Username, err)
					}
				}

			case "newMessage":
				log.Printf("Received message: %+v", msg)

				// envois le message au destinataire
				if targetUser, ok := w.Users[msg.TargetUsername]; ok {
					if err := targetUser.Connection.WriteJSON(msg); err != nil {
						log.Printf("Erreur lors de l'envoi du message à %s : %v", targetUser.Username, err)
						targetUser.Connection.Close()
						delete(w.Users, targetUser.Username)
					}
				}

				// ajoute le message dans l'historique
				w.MessageHistory[msg.SenderUsername] = append(w.MessageHistory[msg.SenderUsername], msg)
				w.MessageHistory[msg.TargetUsername] = append(w.MessageHistory[msg.TargetUsername], msg)

			case "newImage":
				log.Printf("Image à envoyer : %+v", msg)

				for _, user := range w.Users {
					if user.Connection != nil && (user.Username == msg.TargetUsername || user.Username == msg.SenderUsername) {
						err := user.Connection.WriteJSON(msg)
						if err != nil {
							log.Printf("Erreur lors de l'envoi de l'image à %s : %v", user.Username, err)
							user.Connection.Close()
							delete(w.Users, user.Username)
						}
					}
				}

				w.MessageHistory[msg.SenderUsername] = append(w.MessageHistory[msg.SenderUsername], msg)
				w.MessageHistory[msg.TargetUsername] = append(w.MessageHistory[msg.TargetUsername], msg)

			default:
				log.Printf("Unknown message type: %s", msg.Type)
			}

			w.Mu.Unlock()
		}
	}
}

func (w *WebsocketChat) sendHistory(user *UserChat) {
	if messages, ok := w.MessageHistory[user.Username]; ok {
		for _, msg := range messages {
			user.Connection.WriteJSON(msg)
		}
	}
}
