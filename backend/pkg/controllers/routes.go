package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

func (s *MyServer) routes() {

	s.Router.HandleFunc("/protected", Chain(s.ProtectedHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/

	s.Router.Handle("/verify_token", Chain(s.VerifyTokenHandler(), enableCORS, LogRequestMiddleware))
	s.Router.Handle("/register", Chain(s.RegisterHandler(), enableCORS, LogRequestMiddleware))
	s.Router.Handle("/login", Chain(s.LoginHandler(), enableCORS, LogRequestMiddleware))
	s.Router.HandleFunc("/logout", Chain(s.LogoutHandler(), enableCORS, LogRequestMiddleware))

	s.Router.Handle("/create_post", Chain(s.CreatePostHandlers(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/recent_posts", Chain(s.ListPostHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/like_post", Chain(s.LikePost(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/unlike_post", Chain(s.UnlikePost(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/

	s.Router.Handle("/create_comment", Chain(s.CreateCommentHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_comment", Chain(s.ListCommentHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/like_comment", Chain(s.LikeComment(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/unlike_comment", Chain(s.UnlikeComment(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/
	s.Router.Handle("/list_users", Chain(s.ListUsers(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_amis", Chain(s.ListAmis(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/viewprofil/{userId}", Chain(s.GetUserProfilHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/myprofil", Chain(s.MyProfil(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/update_profile", Chain(s.UpdateProfileHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/

	s.Router.HandleFunc("/notifications", Chain(s.GetNotificationsHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/mark_as_read", Chain(
		s.MarkNotificationAsRead(),
		enableCORS,
		LogRequestMiddleware,
		s.Authenticate,
	))
	s.Router.HandleFunc("/follow_request", Chain(s.FollowUserHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/accept_follower", Chain(s.AcceptFollowerHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/decline_follower", Chain(s.DeclineFollowerHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/unfollow", Chain(s.UnfollowUserHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	s.Router.HandleFunc("/search_users", Chain(s.SearchUsersHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/get_follow_requests", Chain(s.GetFollowRequestsHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/

	s.Router.HandleFunc("/online", Chain(s.OnlineUsersHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.HandleFunc("/message", Chain(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.GetMessagesHandler()(w, r)
		case http.MethodPost:
			s.PostMessageHandler()(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}, enableCORS, LogRequestMiddleware, s.Authenticate))

	s.Router.HandleFunc("/ws", Chain(
		s.WebSocketChat.HanderUsersConnection,
		enableCORS,
		LogRequestMiddleware,
	))

	s.Router.HandleFunc("/messagegroup", Chain(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.GetMessagesGroupsHandler()(w, r)
		case http.MethodPost:
			s.PostMessageGroupHandler()(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}, enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/
	s.Router.Handle("/users", Chain(s.SearchUsersHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/group/{id}", Chain(s.GetGroupDataHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_group", Chain(s.ListGroupsHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/create_group", Chain(s.CreateGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/groups/{groupId}/invit_group", Chain(s.InviteToGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/create_post_group", Chain(s.CreatePostGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_post_group", Chain(s.ListPostGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/join_group_request", Chain(s.RequestToJoinGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/create_comment_group", Chain(s.CreateCommentPostsGroup(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_comments_group", Chain(s.ListCommentsByPostGroupHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/

	s.Router.Handle("/group/{id}/create_event", Chain(s.CreateEventHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/list_event", Chain(s.ListEvent(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/respond_to_event", Chain(s.RespondToEventHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/invite_to_event", Chain(s.InviteToEventHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/accept_group_invite", Chain(s.AcceptGroupInviteHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))
	s.Router.Handle("/get_user_votes", Chain(s.GetUserVotesHandler(), enableCORS, LogRequestMiddleware, s.Authenticate))

	/*-------------------------------------------------------------------------------*/
}

// ProtectedHandler affiche un message spécifique à l'utilisateur
func (s *MyServer) ProtectedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(uuid.UUID)
		w.Write([]byte(fmt.Sprintf("Hello, user %s", userID.String())))
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			log.Println("CORS preflight request received")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
