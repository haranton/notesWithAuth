package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"notesauth/internal/models"
	"notesauth/internal/repository"
	"notesauth/internal/utils"
)

type Handlers struct {
	repo      *repository.Repository
	jwtSecret string
}

func New(repo *repository.Repository, jwtSecret string) *Handlers {
	return &Handlers{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (h *Handlers) SetupRoutes() {
	// Публичные маршруты
	http.HandleFunc("/register", h.Register)
	http.HandleFunc("/login", h.Login)

	// Защищенные маршруты
	http.HandleFunc("/notes", h.NotesHandler)
}

func (h *Handlers) Start(port string) error {
	return http.ListenAndServe(":"+port, nil)
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем существование пользователя
	_, err := h.repo.GetUserByLogin(req.Login)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user := &models.User{
		Login:    req.Login,
		Password: req.Password,
	}

	if err := h.repo.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
		"login":   req.Login,
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	user, err := h.repo.GetUserByLogin(req.Login)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создаем токен
	token, err := utils.CreateToken(user.ID, user.Login, h.jwtSecret)
	if err != nil {
		http.Error(w, "Token creation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   token,
		"message": "Login successful",
	})
}

func (h *Handlers) NotesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.GetNotes(w, r)
	case "POST":
		h.CreateNote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handlers) GetNotes(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*utils.Claims)

	notes, err := h.repo.GetNotesByUserID(claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*utils.Claims)

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	note := &models.Note{
		Name:   req.Name,
		UserID: claims.UserID,
	}

	if err := h.repo.CreateNote(note); err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}
