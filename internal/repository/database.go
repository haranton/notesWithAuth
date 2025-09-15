package repository

import (
	"database/sql"
	"fmt"
	"log"
	"notesauth/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository struct {
	db *sql.DB
}

func New(databaseURL string) *Repository {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database not available", err)
	}

	fmt.Println("Database connected successfully")

	return &Repository{db: db}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) createTables() error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        login VARCHAR(100) NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    CREATE TABLE IF NOT EXISTS notes (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        created_at TIMESTAMP DEFAULT NOW()
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRow(query, user.Login, user.Password).Scan(&user.ID)
}

func (r *Repository) GetUserByLogin(login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password FROM users WHERE login = $1`
	err := r.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.Password)
	return user, err
}

func (r *Repository) CreateNote(note *models.Note) error {
	query := `INSERT INTO notes (name, user_id) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRow(query, note.Name, note.UserID).Scan(&note.ID)
}

func (r *Repository) GetNotesByUserID(userID int) ([]*models.Note, error) {
	query := `SELECT id, name, user_id FROM notes WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(&note.ID, &note.Name, &note.UserID)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}
