package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *MyServer) UploadGroupImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifiez la méthode HTTP
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Parsez le formulaire
		err := r.ParseMultipartForm(10 << 20) // Limite de 10 MB
		if err != nil {
			http.Error(w, `{"error": "Failed to parse form"}`, http.StatusBadRequest)
			return
		}

		// Récupérez le fichier
		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, `{"error": "Failed to read file"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Sauvegardez le fichier sur le serveur (ou téléchargez-le sur un service cloud)
		filePath := "./uploads/" + handler.Filename
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, `{"error": "Failed to save file"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, `{"error": "Failed to copy file"}`, http.StatusInternalServerError)
			return
		}

		// Retournez l'URL de l'image
		imageURL := "http://127.0.0.1:8079/image_path/" + handler.Filename
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": imageURL})
	}
}

func UploadImages(w http.ResponseWriter, r *http.Request, filePath string) (string, error) {
	err := r.ParseMultipartForm(20 << 20) // Limite de 20MB
	if err != nil {
		return "", fmt.Errorf("error parsing form: %v", err)
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", fmt.Errorf("error retrieving file from form: %v", err)
	}
	defer file.Close()

	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}

	fullFilePath := filepath.Join(filePath, handler.Filename)
	f, err := os.OpenFile(fullFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("error creating file on server: %v", err)
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return "", fmt.Errorf("error writing file: %v", err)
	}

	// l'URL image
	imageURL := fmt.Sprintf("http://127.0.0.1:8079/image_path/%s", handler.Filename)
	return imageURL, nil
}

func IsValidImageExtension(filename string) bool {
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif"}
	fileExt := strings.ToLower(filepath.Ext(filename))
	for _, ext := range validExtensions {
		if ext == fileExt {
			return true
		}
	}
	return false
}
