package controllers

import (
	"backend/pkg/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
)

type GroupRequest struct {
	GroupID    string `json:"group_id"`
	ReceiverID string `json:"receiver_id"`
}

func (s *MyServer) ListGroupsHandler() http.HandlerFunc {
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

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		var count int
		query := `SELECT COUNT(*) FROM group_members WHERE user_id = ? AND status = 'accepted'`
		err = DB.QueryRow(query, userID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "User is not a member of any group", http.StatusForbidden)
			return
		}

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

		query = `
			SELECT g.id, g.name, g.description
			FROM groups g
			JOIN group_members gm ON g.id = gm.group_id
			WHERE gm.user_id = ? AND gm.status = 'accepted'
			LIMIT ? OFFSET ?
			`
		rows, err := DB.Query(query, userID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var groups []models.Group

		for rows.Next() {
			var group models.Group
			if err := rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
				http.Error(w, "Failed to scan group", http.StatusInternalServerError)
				return
			}
			groups = append(groups, group)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups)
	}
}

func (s *MyServer) CreateGroupHandler() http.HandlerFunc {
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

		var group models.Group
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		group.ID = uuid.Must(uuid.NewV4())
		group.CreatorID = userID

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		query := `INSERT INTO groups (id, name, description, creator_id) VALUES (?, ?, ?, ?)`
		_, err = tx.Exec(query, group.ID, group.Name, group.Description, group.CreatorID)
		if err != nil {
			http.Error(w, "Failed to create group", http.StatusInternalServerError)
			return
		}

		// ajouter le créateur comme membre du groupe avec le rôle de "creator"
		query = `INSERT INTO group_members (id, group_id, user_id, status, role) VALUES (?, ?, ?, 'accepted', 'creator')`
		_, err = tx.Exec(query, uuid.Must(uuid.NewV4()), group.ID, group.CreatorID)
		if err != nil {
			http.Error(w, "Failed to add group creator as member", http.StatusInternalServerError)
			return
		}
		log.Printf("Creating group: %v", group.Name)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Group created successfully"))
	}
}
func (s *MyServer) InviteToGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req GroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
			return
		}

		log.Printf(" demande d'invitation reçue : GroupID=%s, InviteeID=%s\n", req.GroupID, req.ReceiverID)

		if req.ReceiverID == "" || req.GroupID == "" {
			http.Error(w, `{"error": "Invalid Receiver ID or Group ID"}`, http.StatusBadRequest)
			return
		}

		inviterID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, `{"error": "User not logged in"}`, http.StatusUnauthorized)
			return
		}

		receiverID, err := uuid.FromString(req.ReceiverID)
		if err != nil || inviterID == receiverID {
			http.Error(w, `{"error": "Invalid Receiver ID"}`, http.StatusBadRequest)
			return
		}

		groupID, err := uuid.FromString(req.GroupID)
		if err != nil {
			http.Error(w, `{"error": "Invalid Group ID"}`, http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("echec d'ouverture de la base de données", err)
			http.Error(w, `{"error": "Failed to open database"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		var inviterRole string
		query := `SELECT role FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted'`
		err = DB.QueryRow(query, groupID, inviterID).Scan(&inviterRole)
		if err != nil {
			http.Error(w, `{"error": "User not authorized to invite to group"}`, http.StatusUnauthorized)
			return
		}

		var status string
		query = `SELECT status FROM group_members WHERE group_id = ? AND user_id = ?`
		err = DB.QueryRow(query, groupID, receiverID).Scan(&status)
		if err == nil && status == "pending" {
			http.Error(w, `{"error": "User already invited to the group"}`, http.StatusConflict)
			return
		}

		insertQuery := `INSERT INTO group_members (group_id, user_id, role, status) VALUES (?, ?, 'member', 'pending')`
		_, err = DB.Exec(insertQuery, groupID, receiverID)
		if err != nil {
			http.Error(w, `{"error": "Failed to invite user"}`, http.StatusInternalServerError)
			return
		}

		err = s.AddNotification(receiverID.String(), inviterID.String(), "Un utilisateur vous a invité à rejoindre un groupe", "group_invite")
		if err != nil {
			http.Error(w, `{"error": "Failed to add notification"}`, http.StatusInternalServerError)
			return
		}

		log.Println("✅ Invitation envoyée avec succès")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User invited successfully"})
	}
}

func (s *MyServer) AcceptGroupInviteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("📩 Demande d'acceptation de l'invitation reçue")

		// Vérifier la méthode HTTP
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Décoder le corps de la requête
		var request struct {
			GroupID        string `json:"group_id"`
			NotificationID string `json:"notification_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Println("❌ Requête invalide", err)
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Vérifier que les données ne sont pas vides
		if request.GroupID == "" || request.NotificationID == "" {
			http.Error(w, `{"error": "GroupID and NotificationID are required"}`, http.StatusBadRequest)
			return
		}

		// Récupérer l'utilisateur authentifié
		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, `{"error": "User not logged in"}`, http.StatusUnauthorized)
			return
		}

		// Ouvrir la base de données
		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("❌ Erreur de connexion à la base de données", err)
			http.Error(w, `{"error": "Failed to open database"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		// Vérifier que l'invitation existe et est en attente
		var status string
		err = DB.QueryRow("SELECT status FROM group_members WHERE group_id = ? AND user_id = ?", request.GroupID, userID).Scan(&status)
		if err != nil {
			log.Println("❌ L'invitation n'existe pas", err)
			http.Error(w, `{"error": "Invitation not found"}`, http.StatusNotFound)
			return
		}

		if status != "pending" {
			http.Error(w, `{"error": "Invitation is not pending"}`, http.StatusBadRequest)
			return
		}

		// Accepter l'invitation (passer "pending" → "accepted")
		_, err = DB.Exec("UPDATE group_members SET status = 'accepted' WHERE group_id = ? AND user_id = ?", request.GroupID, userID)
		if err != nil {
			log.Println("❌ Échec de l'acceptation de l'invitation", err)
			http.Error(w, `{"error": "Failed to accept invitation"}`, http.StatusInternalServerError)
			return
		}

		// Marquer la notification comme lue
		_, err = DB.Exec("UPDATE notifications SET read = 1 WHERE id = ?", request.NotificationID)
		if err != nil {
			log.Println("❌ Échec de mise à jour de la notification", err)
			http.Error(w, `{"error": "Failed to update notification"}`, http.StatusInternalServerError)
			return
		}

		log.Println("✅ Invitation acceptée avec succès")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invitation accepted"})
	}
}

func (s *MyServer) RequestToJoinGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			GroupID uuid.UUID `json:"group_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("userID").(uuid.UUID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}

		query := `INSERT INTO group_members (group_id, user_id, role, status) VALUES (?, ?, 'member', 'pending')`
		_, err = DB.Exec(query, request.GroupID, userID)
		if err != nil {
			http.Error(w, "Failed to request to join group", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Request sent"})
	}
}

/*-------------------------------------------------------------------------------*/
