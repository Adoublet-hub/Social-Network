package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

func (s *MyServer) GetMessagesGroupsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received GET request to fetch messages")

		username := r.URL.Query().Get("user")
		offsetStr := r.URL.Query().Get("offset")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			offset = 0
		}

		log.Println("Fetching messages for user:", username, "with offset:", offset)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		query := `
			SELECT id, sender_username, target_username, content, timestamp, type, emoji 
			FROM chatGroup 
			WHERE target_username = ? 
			ORDER BY timestamp ASC 
			LIMIT 10 OFFSET ?
		`

		rows, err := DB.Query(query, username, offset)

		if err != nil {
			log.Println("Failed to fetch chatGroup:", err)
			http.Error(w, "Failed to fetch chatGroup", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []map[string]interface{}

		for rows.Next() {
			var id, senderUsername, targetUsername, content, messageType string
			var timestamp time.Time
			var emoji sql.NullString

			if err := rows.Scan(&id, &senderUsername, &targetUsername, &content, &timestamp, &messageType, &emoji); err != nil {
				log.Println("Failed to scan message:", err)
				http.Error(w, "Failed to scan message", http.StatusInternalServerError)
				return
			}

			msg := map[string]interface{}{
				"id":              id,
				"sender_username": senderUsername,
				"target_username": targetUsername,
				"content":         content,
				"timestamp":       timestamp,
				"type":            messageType,
				"emoji":           "",
			}
			if emoji.Valid {
				msg["emoji"] = emoji.String
			}
			messages = append(messages, msg)
		}

		log.Println("Fetched", len(messages), "messages for user:", username)

		if len(messages) == 0 {
			messages = []map[string]interface{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			log.Println("Failed to encode messages:", err)
			http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
		}
	}
}

func (s *MyServer) PostMessageGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.WebSocketChat.Mu.Lock()
		defer s.WebSocketChat.Mu.Unlock()

		log.Println("Received POST request to send a message")

		var msg models.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			log.Println("Invalid message format:", err)
			http.Error(w, "Invalid message format", http.StatusBadRequest)
			return
		}

		log.Printf("Message reçu : %+v", msg)

		if msg.Content == "" || msg.TargetUsername == "" {
			http.Error(w, "Message content or target username is missing", http.StatusBadRequest)
			return
		}

		msg.ID = uuid.Must(uuid.NewV4())
		msg.Timestamp = time.Now()
		msg.Type = "newMessage"

		sender, ok := r.Context().Value(usernameIDKey).(string)
		if !ok {
			log.Println("Username not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		msg.SenderUsername = sender

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		query := `INSERT INTO chatGroup (id, sender_username, target_username, content, timestamp, type, emoji) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = DB.Exec(query, msg.ID.String(), msg.SenderUsername, msg.TargetUsername, msg.Content, msg.Timestamp, msg.Type, msg.Emoji)
		if err != nil {
			log.Printf("Failed to save message: %v", err)
			http.Error(w, "Failed to save message", http.StatusInternalServerError)
			return
		}

		message := map[string]interface{}{
			"id":              msg.ID.String(),
			"sender_username": msg.SenderUsername,
			"target_username": msg.TargetUsername,
			"content":         msg.Content,
			"timestamp":       msg.Timestamp,
			"type":            msg.Type,
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"status":  "success",
			"message": "Message enregistré avec succès",
		}
		json.NewEncoder(w).Encode(response)

		if receiverConn, ok := s.WebSocketChat.Users[msg.TargetUsername]; ok && receiverConn.Connection != nil {
			if err := receiverConn.Connection.WriteJSON(message); err != nil {
				log.Printf("Failed to send message to receiver: %v", err)
				delete(s.WebSocketChat.Users, msg.TargetUsername)
			}
		}

		if senderConn, ok := s.WebSocketChat.Users[msg.SenderUsername]; ok && senderConn.Connection != nil {
			if err := senderConn.Connection.WriteJSON(message); err != nil {
				log.Printf("Failed to send message to sender: %v", err)
				delete(s.WebSocketChat.Users, msg.SenderUsername)
			}
		}

	}
}
