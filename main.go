package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-this-in-production")
var db *sql.DB

type Claims struct {
	UserID   int
	Username string
	jwt.RegisteredClaims
}

type Note struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}

type createdNote struct {
	Name string `json:"name"`
}

type User struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type createdUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func createJWTToken(userID int, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := validateJWTToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		next(w, r.WithContext(ctx))

	}
}

func createTables() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login VARCHAR(100) NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		user_id INT References users(id)
	)
	`)
	if err != nil {
		log.Fatal("failed create table")
	}
}

func getNotes(w http.ResponseWriter, r *http.Request) {

	userClaims := r.Context().Value("user").(*Claims)
	userID := userClaims.UserID

	rows, err := db.Query("SELECT id, name from notes where user_id = $1", userID)
	if err != nil {
		http.Error(w, "failed request to db", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	arrNotes := []Note{}

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.Id, &note.Name); err != nil {
			http.Error(w, "failed reading rows", http.StatusInternalServerError)
			return
		}
		arrNotes = append(arrNotes, note)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arrNotes)
}

func createNote(w http.ResponseWriter, r *http.Request) {

	userClaims := r.Context().Value("user").(*Claims)
	userID := userClaims.UserID

	var note createdNote
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, "Error read body of request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	_, err = db.Exec("INSERT INTO notes (name, user_id) VALUES ($1, $2)", note.Name, userID)
	if err != nil {
		http.Error(w, "Error created note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)

}

func notesRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getNotes(w, r)
	case "POST":
		createNote(w, r)
	}

}

func registerUser(w http.ResponseWriter, r *http.Request) {

	var user createdUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error read body of request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var existUser string
	err = db.QueryRow("select login from users where login =$1", user.Login).Scan(&existUser)
	if err == nil {
		http.Error(w, "User already exists", http.StatusInternalServerError)
		return
	}

	if err != sql.ErrNoRows {
		// Ошибка базы данных
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (login, password) VALUES ($1, $2)", user.Login, user.Password)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error read body of request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var existUser User
	err = db.QueryRow("SELECT id, login, password FROM users WHERE login = $1 AND password = $2", user.Login, user.Password).Scan(
		&existUser.Id,       // Сканируем id в поле ID
		&existUser.Login,    // Сканируем login в поле Login
		&existUser.Password, // Сканируем password в поле Password
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if existUser.Login != "" {

		token, err := createJWTToken(existUser.Id, existUser.Login)

		if err != nil {
			http.Error(w, "Token creation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]string{
			"token":   token,
			"message": "login succesful",
		})
	}
}

func main() {

	db = getDb()
	defer db.Close()

	createTables()

	http.HandleFunc("/notes", authMiddleware(notesRouter))
	http.HandleFunc("/login", loginUser)
	http.HandleFunc("/register", registerUser)
	port := ":8080"
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
