# Utilisation de l'image officielle Golang comme base
FROM golang:1.18-alpine 

# Définir le répertoire de travail à l'intérieur du conteneur
WORKDIR /app

# Copier le fichier `go.mod` et `go.sum` pour installer les dépendances
COPY go.mod go.sum ./

# Télécharger toutes les dépendances dans le module Go
RUN go mod download

# Copier le reste des fichiers de ton backend
COPY . .

# Construire l'application Go (produit un fichier exécutable nommé 'main')
RUN go build -o main ./main.go

# Exposer le port sur lequel ton backend sera accessible
EXPOSE 8080

# Commande pour exécuter l'application
CMD ["/app/main"]
