package main

import (
	"backend/pkg/controllers"
	"backend/pkg/db"
	"backend/pkg/wsk"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Error to run : %v\n", err)
		os.Exit(1)
	}
}

// Fonction principale pour exécuter le serveur
func run() error {
	store := &db.DBStore{}
	wsChat := wsk.NewWebsocketChat()
	srv := controllers.NewServer(store, wsChat)

	db, err := store.OpenDatabase()
	if err != nil {
		return fmt.Errorf("failed to open database : %w", err)
	}

	// Fonction pour fermer la base de données à la fin
	defer func() {
		if err := srv.Store.CloseDatabase(db); err != nil {
			log.Printf("error when closing database : %v\n", err)
		}
	}()

	// Configuration pour écouter les signaux d'arrêt
	signalChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine pour gérer l'arrêt du serveur
	go func() {
		<-signalChan
		log.Println("stopping the server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("server shutdown error: %w", err)
		}
		close(done)
	}()

	// Goroutine pour gérer le chat WebSocket
	go func() {
		wsChat.UsersChatManager()
	}()

	if err := srv.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error when starting server : %w", err)
	}

	<-done
	log.Println("Server gracefully stopped")
	return nil
}
