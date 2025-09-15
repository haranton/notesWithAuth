package models

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"-"`
}

// Note представляет заметку
type Note struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}

// CreateUserRequest запрос на создание пользователя
type CreateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginRequest запрос на авторизацию
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CreateNoteRequest запрос на создание заметки
type CreateNoteRequest struct {
	Name string `json:"name"`
}

// JWT Claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}
