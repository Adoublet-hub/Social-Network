package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

func (s *MyServer) CreateEventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		var event models.GroupEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Println("Invalid request payload", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		log.Printf("Event Title: %s, Description: %s, EventDate: %s, GroupID: %s, UserID: %s", event.Title, event.Description, event.EventDate.Format(time.RFC3339), event.GroupID, event.UserID)

		event.ID = uuid.Must(uuid.NewV4())
		event.UserID = userID

		if event.Title == "" || event.Description == "" || event.EventDate.IsZero() || event.GroupID == uuid.Nil {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		log.Printf("Event Title: %s, Description: %s, EventDate: %s, GroupID: %s, UserID: %s", event.Title, event.Description, event.EventDate.Format(time.RFC3339), event.GroupID, event.UserID)

		query := `INSERT INTO group_events (id, group_id, user_id, title, description, event_date) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = DB.Exec(query, event.ID, event.GroupID, event.UserID, event.Title, event.Description, event.EventDate)
		if err != nil {
			log.Println("Failed to create event", err)
			http.Error(w, "Failed to create event", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":          event.ID,
			"title":       event.Title,
			"description": event.Description,
			"event_date":  event.EventDate.Format(time.RFC3339),
			"group_id":    event.GroupID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	}
}

func (s *MyServer) ListEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		log.Println("userID :", userID)

		query := r.URL.Query()
		GroupIDStr := query.Get("group_id")
		if GroupIDStr == "" {
			log.Println("Group ID not provided")
			http.Error(w, "Group ID not provided", http.StatusBadRequest)
			return
		}
		groupID, err := uuid.FromString(GroupIDStr)
		if err != nil {
			http.Error(w, "Invalid Group ID", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println()
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		page := 1
		limit := 10

		queryParams := r.URL.Query()
		if p := queryParams.Get("page"); p != "" {
			page, err = strconv.Atoi(p)
			if err != nil || page < 1 {
				page = 1
			}
		}

		if l := queryParams.Get("limit"); l != "" {
			limit, err = strconv.Atoi(l)
			if err != nil || limit < 1 {
				limit = 10
			}
		}

		offset := (page - 1) * limit
		log.Printf("Fetching event from database (page: %d, limit: %d)\n", page, limit)

		events, err := GetEventByGroup(DB, groupID, offset, limit)
		if err != nil {
			http.Error(w, "Failed to retrDBieve event", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "Failed to encode events", http.StatusInternalServerError)
		}

		log.Printf("Événements récupérés pour le groupe %s : %+v", groupID, events)

	}
}

func GetEventByGroup(DB *sql.DB, groupID uuid.UUID, offset, limit int) ([]models.GroupEvent, error) {

	if DB == nil {
		return nil, errors.New("database connection is nil")
	}
	rows, err := DB.Query("SELECT id, group_id, user_id, title, description, event_date, created_at FROM group_events WHERE group_id = ? LIMIT ? OFFSET ?", groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.GroupEvent
	for rows.Next() {
		var event models.GroupEvent
		err := rows.Scan(&event.ID, &event.GroupID, &event.UserID, &event.Title, &event.Description, &event.EventDate, &event.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Par défaut, event.Options est à zéro, ou récupère des données  si nécessaire
		event.Options = models.EventOptions{
			Going:    0,
			NotGoing: 0,
		}

		events = append(events, event)
	}

	return events, nil
}

func (s *MyServer) RespondToEventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		var response models.EventResponse
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
			log.Println("❌ JSON Decode Error:", err)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if response.Response != "Going" && response.Response != "Not going" {
			log.Println("❌ Invalid Response Value:", response.Response)
			http.Error(w, "Invalid response", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("❌ Database Connection Error:", err)
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		var existingResponse string
		err = DB.QueryRow("SELECT response FROM event_responses WHERE event_id = ? AND user_id = ?", response.EventID, userID).Scan(&existingResponse)

		if err == nil {
			if existingResponse == response.Response {
				// Annulation du vote
				_, err = DB.Exec("DELETE FROM event_responses WHERE event_id = ? AND user_id = ?", response.EventID, userID)
				if err != nil {
					log.Println("❌ Failed to Remove Vote:", err)
					http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
					return
				}
			} else {
				// Mise à jour du vote
				_, err = DB.Exec("UPDATE event_responses SET response = ? WHERE event_id = ? AND user_id = ?", response.Response, response.EventID, userID)
				if err != nil {
					log.Println("❌ Failed to Update Vote:", err)
					http.Error(w, "Failed to update vote", http.StatusInternalServerError)
					return
				}
			}
		} else {
			// Premier vote
			_, err = DB.Exec("INSERT INTO event_responses (event_id, user_id, response) VALUES (?, ?, ?)", response.EventID, userID, response.Response)
			if err != nil {
				log.Println("❌ Failed to Insert Vote:", err)
				http.Error(w, "Failed to insert vote", http.StatusInternalServerError)
				return
			}
		}

		// Mise à jour des compteurs de votes
		_, err = DB.Exec(`
			UPDATE group_events
			SET options = JSON_SET(
				IFNULL(options, '{}'),
				'$."Going"', (SELECT COUNT(*) FROM event_responses WHERE event_id = ? AND response = 'Going'),
				'$."NotGoing"', (SELECT COUNT(*) FROM event_responses WHERE event_id = ? AND response = 'NotGoing')
			)
			WHERE id = ?`, response.EventID, response.EventID, response.EventID)

		if err != nil {
			log.Println("❌ Failed to Update Vote Counts:", err)
			http.Error(w, "Failed to update vote counts", http.StatusInternalServerError)
			return
		}

		log.Println("✅ Vote Successfully Recorded for Event:", response.EventID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Vote successfully recorded"})
	}
}

func (s *MyServer) InviteToEventHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		log.Println("User ID:", userID)

		var inviteRequest struct {
			EventID  uuid.UUID `json:"event_id"`
			Username string    `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&inviteRequest); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		//  si l'événement existe
		var exists bool
		err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_events WHERE id = ?)", inviteRequest.EventID).Scan(&exists)
		if err != nil || !exists {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}

		//  si l'utilisateur existe
		var invitedUserID uuid.UUID
		err = DB.QueryRow("SELECT id FROM users WHERE username = ?", inviteRequest.Username).Scan(&invitedUserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// ajouter l'invitation
		query := "INSERT INTO event_invitations (event_id, user_id) VALUES (?, ?)"
		_, err = DB.Exec(query, inviteRequest.EventID, invitedUserID)
		if err != nil {
			http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invitation sent successfully"})
	}
}

func (s *MyServer) GetUserVotesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		query := r.URL.Query()
		groupIDStr := query.Get("group_id")
		if groupIDStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		groupID, err := uuid.FromString(groupIDStr)
		if err != nil {
			http.Error(w, "Invalid Group ID", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		rows, err := DB.Query(`
			SELECT event_id, response FROM event_responses 
			WHERE user_id = ? AND event_id IN (
				SELECT id FROM group_events WHERE group_id = ?
			)
		`, userID, groupID)

		if err != nil {
			http.Error(w, "Failed to fetch votes", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var votes []map[string]interface{}
		for rows.Next() {
			var eventID uuid.UUID
			var response string
			if err := rows.Scan(&eventID, &response); err != nil {
				http.Error(w, "Failed to scan votes", http.StatusInternalServerError)
				return
			}

			votes = append(votes, map[string]interface{}{
				"event_id": eventID,
				"response": response,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(votes)
	}
}
