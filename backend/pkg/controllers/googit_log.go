package controllers

/*
import (
	"backend/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func (s *MyServer) GoogleRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := s.GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *MyServer) GoogleCallbackRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := s.GoogleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := s.GoogleOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email     string `json:"email"`
			FirstName string `json:"given_name"`
			LastName  string `json:"family_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user := models.User{
			Username:  userInfo.Email,
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			Email:     userInfo.Email,
			Password:  "", //le mot de passe peut être vide
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		err = RegisterUser(w, r, DB, user)
		if err != nil {
			http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *MyServer) GitHubRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := s.GitHubOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *MyServer) GitHubCallbackRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := s.GitHubOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := s.GitHubOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email     string `json:"email"`
			Login     string `json:"login"`
			FirstName string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user := models.User{
			Username:  userInfo.Login,
			FirstName: userInfo.FirstName,
			LastName:  "",
			Email:     userInfo.Email,
			Password:  "",
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		err = RegisterUser(w, r, DB, user)
		if err != nil {
			http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
*/
/*
func (s *MyServer) GoogleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GoogleLogin handler called")
		url := s.GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
		log.Println("Redirecting to:", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *MyServer) GoogleCallBackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GoogleCallBackHandler called")
		code := r.URL.Query().Get("code")
		log.Println("Authorization code:", code)
		if code == "" {
			http.Error(w, "No code in URL", http.StatusBadRequest)
			return
		}
		token, err := s.GoogleOAuthConfig.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Println("Failed to exchange token:", err)
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		client := s.GoogleOAuthConfig.Client(oauth2.NoContext, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			log.Println("Failed to get user info:", err)
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			log.Println("Failed to decode user info:", err)
			http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("User authenticated: %s", userInfo.Email)
		fmt.Fprintf(w, "User authenticated: %s", userInfo.Email)
	}
}

func (s *MyServer) GitHubLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := s.GitHubOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *MyServer) GitHubCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := s.GitHubOAuthConfig.Exchange(oauth2.NoContext, code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		client := s.GitHubOAuthConfig.Client(oauth2.NoContext, token)
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
			http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var primaryEmail string
		for _, email := range emails {
			if email.Primary {
				primaryEmail = email.Email
				break
			}
		}
		fmt.Printf("User authenticated %s \n", primaryEmail)
		fmt.Fprintf(w, "User authenticated: %s", primaryEmail)
	}
}
*/
