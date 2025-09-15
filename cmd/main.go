package main

import (
	"log"
	"notesauth/internal/config"
	"notesauth/internal/handlers"
	"notesauth/internal/repository"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация репозитория
	repo := repository.New(cfg.DatabaseURL)

	// Инициализация обработчиков
	h := handlers.New(repo, cfg.JWTSecret)

	// Настройка маршрутов
	h.SetupRoutes()

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(h.Start(cfg.Port))
}
