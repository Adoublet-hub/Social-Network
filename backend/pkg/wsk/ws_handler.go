package wsk

import (
	"backend/pkg/models"
	"backend/pkg/zwt"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (w *WebsocketChat) HanderUsersConnection(wr http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(wr, r, nil)
	if err != nil {
		log.Printf("Échec de l'upgrade WebSocket : %v", err)
		http.Error(wr, "Failed to establish WebSocket connection", http.StatusInternalServerError)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Println("Token manquant dans la query string")
		http.Error(wr, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := zwt.VerifyJWT(token)
	if err != nil || claims.Username == "" {
		log.Printf("Token invalide ou utilisateur non défini : %v", err)
		http.Error(wr, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := claims.Username
	log.Printf("Connexion WebSocket pour l'utilisateur : %s", username)

	w.Mu.Lock()
	if existingUser, exists := w.Users[username]; exists {
		if existingUser.Connection == conn {
			log.Printf("Connexion déjà établie pour l'utilisateur : %s", username)
			w.Mu.Unlock()
			return
		}
	}
	w.Mu.Unlock()

	userChat := NewUserChat(&Channel{
		messageChannel: w.MessageChannel,
		leaveChannel:   w.LeaveChannel,
	}, username, conn)

	w.JoinChannel <- userChat

	go userChat.listenForMessages()
}

func (u *UserChat) handleOutgoingMessages() {
	for {
		select {
		case msg := <-u.channels.messageChannel:
			if msg.TargetUsername == u.Username {
				err := u.Connection.WriteJSON(msg)
				if err != nil {
					log.Printf("Erreur lors de l'envoi du message à %s : %v", u.Username, err)
				}
			}
		}
	}
}

func (w *WebsocketChat) ReplaceUser(user *UserChat) {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	if existingUser, exists := w.Users[user.Username]; exists {
		existingUser.Connection.Close()
	}
	w.Users[user.Username] = user
}

func (u *UserChat) listenForMessages() {
	defer func() {
		u.channels.leaveChannel <- u
		u.Connection.Close()
		log.Printf("Closing connection for user: %s", u.Username)
	}()

	for {
		var msg models.Message
		_, data, err := u.Connection.ReadMessage()
		if err != nil {
			log.Printf("Erreur lors de la lecture du message pour %s: %v", u.Username, err)
			return
		}

		log.Printf("Message reçu de %s : %s", u.Username, string(data))

		err = json.Unmarshal(data, &msg)
		if err != nil {
			log.Printf("Format de message invalide de %s : %v", u.Username, err)
			continue
		}

		msg.SenderUsername = u.Username
		msg.Timestamp = time.Now()

		switch msg.Type {
		case "newMessage":
			log.Printf("Message texte à envoyer : %+v", msg)
			u.channels.messageChannel <- &msg
		case "newImage":
			log.Printf("Image à envoyer : %+v", msg)
			u.channels.messageChannel <- &msg
		default:
			log.Printf("Type de message inconnu : %s", msg.Type)
		}
	}
}
