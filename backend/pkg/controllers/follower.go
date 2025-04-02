package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type FollowRequest struct {
	ReceiverID string `json:"friend_id"`
}

const (
	ErrInvalidUUID       = "Invalid UUID format"
	ErrInternalServer    = "Internal server error"
	MsgFollowRequestSent = "Follow request sent"
	MsgFollowAccepted    = "Follow request accepted"
	MsgFollowDeclined    = "Follow request declined"
	MsgInvalidJSONBody   = "Invalid JSON body"
)

func writeJSONResponse(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (s *MyServer) FollowUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		var req FollowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: MsgInvalidJSONBody,
			})
			return
		}

		senderID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Unauthorized",
			})
			return
		}

		receiverID, err := uuid.FromString(req.ReceiverID)
		if err != nil || senderID == receiverID {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: ErrInvalidUUID,
			})
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: ErrInternalServer,
			})
			return
		}
		defer DB.Close()

		var exists int
		err = DB.QueryRow(`
			SELECT COUNT(*) FROM follow_requests WHERE sender_id = ? AND receiver_id = ?
		`, senderID, receiverID).Scan(&exists)

		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: ErrInternalServer,
			})
			return
		}

		if exists > 0 {
			writeJSONResponse(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "Follow request already exists or user is already followed",
			})
			return
		}

		var isPrivate bool
		err = DB.QueryRow("SELECT is_private FROM users WHERE id = ?", receiverID).Scan(&isPrivate)
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}

		if isPrivate {
			// 🔹 Si compte privé → Ajouter à `follow_requests`
			_, err = DB.Exec("INSERT INTO follow_requests (sender_id, receiver_id) VALUES (?, ?)", senderID, receiverID)
			if err != nil {
				writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Failed to create follow request",
				})
				return
			}

			err = s.AddNotification(receiverID.String(), senderID.String(), "Nouvelle demande de suivi", "follow_request")
			if err != nil {
				log.Println("⚠️ Erreur lors de l'ajout de la notification :", err)
			}

			writeJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "Follow request sent",
			})
		} else {
			// 🔹 Si compte public → Ajouter directement à `followers`
			_, err = DB.Exec("INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, 'accepted')", senderID, receiverID)
			if err != nil {
				writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Failed to follow user",
				})
				return
			}

			err = s.AddNotification(receiverID.String(), senderID.String(), "Un utilisateur a commencé à vous suivre", "follow")
			if err != nil {
				log.Println("⚠️ Erreur lors de l'ajout de la notification :", err)
			}

			writeJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "You are now following this user",
			})
		}
	}
}

func (s *MyServer) UnfollowUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		var req struct {
			FollowedID string `json:"followed_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid JSON body",
			})
			return
		}

		log.Println("FollowedID reçu:", req.FollowedID)

		followerID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Unauthorized",
			})
			return
		}

		if req.FollowedID == "" {
			log.Println("❌ Erreur: FollowedID est vide")
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "FollowedID cannot be empty",
			})
			return
		}

		followedID, err := uuid.FromString(req.FollowedID)
		if err != nil {
			log.Println("❌ UUID invalide reçu:", req.FollowedID)
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid UUID format",
			})
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		defer DB.Close()

		// 🔹 Vérifier si l'utilisateur suit bien la personne avant de se désabonner
		var exists int
		err = DB.QueryRow(`
			SELECT COUNT(*) FROM followers 
			WHERE follower_id = ? AND followed_id = ?
		`, followerID, followedID).Scan(&exists)

		if err != nil {
			log.Println("❌ Erreur lors de la vérification du follow:", err)
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to verify follow status",
			})
			return
		}

		if exists == 0 {
			log.Println("❌ Erreur: L'utilisateur ne suit pas cette personne")
			writeJSONResponse(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "You are not following this user",
			})
			return
		}

		// 🔹 Supprimer l'entrée de la table `followers`
		_, err = DB.Exec("DELETE FROM followers WHERE follower_id = ? AND followed_id = ?", followerID, followedID)
		if err != nil {
			log.Println("❌ Erreur lors de la suppression du follow:", err)
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to unfollow user",
			})
			return
		}

		// 🔹 Supprimer une demande de suivi en attente (si elle existe)
		_, err = DB.Exec("DELETE FROM follow_requests WHERE sender_id = ? AND receiver_id = ?", followerID, followedID)
		if err != nil {
			log.Println("⚠️ Aucune demande de suivi trouvée ou erreur:", err)
		}

		// 🔹 Envoyer une notification si nécessaire
		err = s.AddNotification(followedID.String(), followerID.String(), "Un utilisateur s'est désabonné de vous", "unfollow")
		if err != nil {
			log.Println("⚠️ Erreur lors de l'ajout de la notification:", err)
		}

		writeJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Successfully unfollowed user",
		})
	}
}

func (s *MyServer) GetFollowRequestsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			writeJSONResponse(w, http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Unauthorized",
			})
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to open database",
			})
			return
		}
		defer DB.Close()

		rows, err := DB.Query(`
			SELECT fr.id, fr.sender_id, u.username, u.avatar
			FROM follow_requests fr
			JOIN users u ON fr.sender_id = u.id
			WHERE fr.receiver_id = ?`, userID)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to retrieve follow requests",
			})
			return
		}
		defer rows.Close()

		var requests []map[string]interface{}
		for rows.Next() {
			var requestID, senderID uuid.UUID
			var username, avatar string

			if err := rows.Scan(&requestID, &senderID, &username, &avatar); err != nil {
				writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
					Success: false,
					Message: "Failed to parse follow requests",
				})
				return
			}

			requests = append(requests, map[string]interface{}{
				"id":        requestID,
				"sender_id": senderID,
				"username":  username,
				"avatar":    avatar,
				"type":      "follow_request",
				"content":   "Vous avez une nouvelle demande d'ami.",
			})
		}

		writeJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Follow requests retrieved successfully",
			Data:    requests,
		})
	}
}

func (s *MyServer) AcceptFollowerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RequestID string `json:"request_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid JSON body",
			})
			return
		}

		requestID, err := uuid.FromString(req.RequestID)
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid UUID format",
			})
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		defer DB.Close()

		var senderID, receiverID uuid.UUID
		err = DB.QueryRow(`
			SELECT sender_id, receiver_id FROM follow_requests WHERE id = ?;
		`, requestID).Scan(&senderID, &receiverID)

		if err != nil {
			log.Println("❌ Erreur: Demande de suivi non trouvée pour ID", requestID)
			writeJSONResponse(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Follow request not found",
			})
			return
		}

		var alreadyFollowing int
		err = DB.QueryRow(`
			SELECT COUNT(*) FROM followers WHERE follower_id = ? AND followed_id = ?;
		`, senderID, receiverID).Scan(&alreadyFollowing)

		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Database error",
			})
			return
		}

		if alreadyFollowing > 0 {
			log.Println("⚠️ L'utilisateur suit déjà cette personne :", senderID, "->", receiverID)
			writeJSONResponse(w, http.StatusConflict, APIResponse{
				Success: false,
				Message: "Already following this user",
			})
			return
		}

		// 🔹 Ajouter le follower et supprimer la demande
		_, err = DB.Exec(`
			INSERT INTO followers (follower_id, followed_id) VALUES (?, ?);
			DELETE FROM follow_requests WHERE id = ?;
		`, senderID, receiverID, requestID)

		if err != nil {
			log.Println("❌ Erreur lors de l'ajout du follower :", err)
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to accept follow request",
			})
			return
		}

		err = s.AddNotification(senderID.String(), receiverID.String(), "Votre demande de suivi a été acceptée", "follow_accepted")
		if err != nil {
			log.Println("⚠️ Erreur lors de l'ajout de la notification :", err)
		}

		writeJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Follower request accepted",
		})
	}
}

func (s *MyServer) DeclineFollowerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RequestID string `json:"request_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid JSON body",
			})
			return
		}

		requestID, err := uuid.FromString(req.RequestID)
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid UUID format",
			})
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		defer DB.Close()

		var senderID, receiverID uuid.UUID
		err = DB.QueryRow(`
			SELECT sender_id, receiver_id FROM follow_requests WHERE id = ?;
		`, requestID).Scan(&senderID, &receiverID)

		if err != nil {
			log.Println("❌ Erreur: Demande de suivi non trouvée pour ID", requestID)
			writeJSONResponse(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Follow request not found",
			})
			return
		}

		// 🔹 Supprimer la demande de suivi
		_, err = DB.Exec("DELETE FROM follow_requests WHERE id = ?", requestID)
		if err != nil {
			log.Println("❌ Erreur lors de la suppression de la demande :", err)
			writeJSONResponse(w, http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to decline follow request",
			})
			return
		}

		// 🔹 Envoyer une notification
		err = s.AddNotification(senderID.String(), receiverID.String(), "Votre demande de suivi a été refusée", "follow_declined")
		if err != nil {
			log.Println("⚠️ Erreur lors de l'ajout de la notification :", err)
		}

		writeJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Follower request declined",
		})
	}
}
