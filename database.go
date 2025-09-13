package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func getDb() *sql.DB {

	dsn := "postgres://mydb:mydb@localhost:5432/mydb?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("failed to connect db", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Postgres not avaliable")
	}

	fmt.Println("Postgres connect sussesfully")

	return db
}
