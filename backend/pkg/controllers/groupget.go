package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"
)

func (s *MyServer) GetGroupDataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		re := regexp.MustCompile(`^/group/([a-zA-Z0-9-]+)$`)
		matches := re.FindStringSubmatch(r.URL.Path)
		if len(matches) < 2 {
			http.Error(w, `{"error": "Group ID is required"}`, http.StatusBadRequest)
			return
		}
		groupID := matches[1]
		log.Println("ID du groupe :", groupID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		var group models.Group
		query := `SELECT id, name, description, creator_id, created_at FROM groups WHERE id = ?`
		err = tx.QueryRow(query, groupID).Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID, &group.CreatedAt)
		if err != nil {
			http.Error(w, `{"error": "Group not found"}`, http.StatusNotFound)
			return
		}

		members, err := s.fetchGroupMembers(tx, groupID)
		if err != nil {
			http.Error(w, `{"error": "Failed to load members"}`, http.StatusInternalServerError)
			return
		}
		log.Println("Membres du groupe :", members)

		var posts []models.PostGroup
		postQuery := `SELECT id, group_id, user_id, title, content, created_at, updated_at FROM group_posts WHERE group_id = ?`
		postRows, err := tx.Query(postQuery, groupID)
		if err != nil {
			http.Error(w, `{"error": "Failed to load posts"}`, http.StatusInternalServerError)
			return
		}
		defer postRows.Close()

		for postRows.Next() {
			var post models.PostGroup
			if err := postRows.Scan(&post.ID, &post.GroupID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt); err != nil {
				http.Error(w, `{"error": "Failed to scan post"}`, http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}

		group.Members = members
		response := map[string]interface{}{
			"group":   group,
			"members": members,
			//"messages": messages,
			"posts": posts,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

func (s *MyServer) fetchGroupMessages(DB *sql.DB, groupID, username string, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, sender_username, target_username, content, timestamp, type, emoji 
		FROM chatGroup 
		WHERE group_id = ? AND (sender_username = ? OR target_username = ?) 
		ORDER BY timestamp DESC 
		LIMIT 10 OFFSET ?`

	rows, err := DB.Query(query, groupID, username, username, offset)
	if err != nil {
		log.Println("Failed to fetch chatGroup:", err)
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, senderUsername, targetUsername, content, messageType string
		var timestamp time.Time
		var emoji sql.NullString

		if err := rows.Scan(&id, &senderUsername, &targetUsername, &content, &timestamp, &messageType, &emoji); err != nil {
			log.Println("Failed to scan message:", err)
			return nil, err
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
	return messages, nil
}

func (s *MyServer) fetchGroupMembers(tx *sql.Tx, groupID string) ([]models.GroupMember, error) {
	query := `
		SELECT gm.user_id, u.username, u.avatar, gm.role, gm.status 
		FROM group_members gm 
		INNER JOIN users u ON gm.user_id = u.id 
		WHERE gm.group_id = ?`

	rows, err := tx.Query(query, groupID)
	if err != nil {
		log.Println("Failed to load members:", err)
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var member models.GroupMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.Avatar, &member.Role, &member.Status); err != nil {
			log.Println("Error scanning member row:", err)
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}
