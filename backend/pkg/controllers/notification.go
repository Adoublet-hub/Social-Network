package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

func (s *MyServer) AddNotification(userID, senderID, content, notificationType string) error {
	DB, err := s.Store.OpenDatabase()
	if err != nil {
		log.Println("⚠️ Erreur ouverture DB", err)
		return err
	}
	defer DB.Close()

	notificationID := uuid.Must(uuid.NewV4()).String()
	log.Println(" Ajout d'une notification :", notificationID, "| Destinataire:", userID, "| Type:", notificationType)

	_, err = DB.Exec(
		"INSERT INTO notifications (id, user_id, sender_id, content, type) VALUES (?, ?, ?, ?, ?)",
		notificationID, userID, senderID, content, notificationType,
	)
	if err != nil {
		log.Println(" Erreur insertion notification", err)
		return err
	}

	log.Println("Notification ajoutée avec succès:", notificationID)
	return nil
}

func (s *MyServer) GetNotificationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("⚠️ userID manquant du contexte")
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		log.Println("🔍 Récupération des notifications pour l'utilisateur :", userID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("⚠️ Erreur ouverture DB:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		rows, err := DB.Query(`
			SELECT n.id, n.content, n.created_at, n.read, n.type, u.username, u.avatar
			FROM notifications n
			LEFT JOIN users u ON n.sender_id = u.id
			WHERE n.user_id = ? AND n.read = 0
			ORDER BY n.created_at DESC
		`, userID.String())

		if err != nil {
			log.Println("⚠️ Erreur lors de la récupération des notifications :", err)
			http.Error(w, "Failed to retrieve notifications", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var notifications []map[string]interface{}
		for rows.Next() {
			var id, content, notificationType, username string
			var avatar sql.NullString
			var createdAt string
			var read bool

			err = rows.Scan(&id, &content, &createdAt, &read, &notificationType, &username, &avatar)
			if err != nil {
				log.Println("⚠️ Erreur scan notification:", err)
				continue
			}
			notifications = append(notifications, map[string]interface{}{
				"id":          id,
				"content":     content,
				"created_at":  createdAt,
				"read":        read,
				"type":        notificationType,
				"sender_name": username,
				"avatar":      avatar.String,
			})
		}

		if notifications == nil {
			notifications = []map[string]interface{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notifications)
	}
}

func (s *MyServer) MarkNotificationAsRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérification de la méthode HTTP
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Décoder le corps de la requête
		var request struct {
			NotificationID string `json:"notification_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		if request.NotificationID == "" {
			http.Error(w, `{"error": "Notification ID is required"}`, http.StatusBadRequest)
			return
		}

		// Connexion à la base de données
		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		// Mise à jour de la notification comme "lue"
		_, err = DB.Exec(`UPDATE notifications SET read = 1 WHERE id = ?`, request.NotificationID)
		if err != nil {
			log.Println("Failed to update notification:", err)
			http.Error(w, `{"error": "Failed to mark notification as read"}`, http.StatusInternalServerError)
			return
		}

		// Réponse de succès
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "Notification marked as read"}`))
	}
}
