package db

import (
	"database/sql"
	"fmt"
	"os"
	//"log"
)

var database *sql.DB

func Connect() error {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://postgres:1234@db:5432/go?sslmode=disable"
	}
	database, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err := database.Ping(); err != nil {
		return err
	}

	_, err = database.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            username VARCHAR(50) UNIQUE,
            password VARCHAR(100),
            email VARCHAR(50) UNIQUE
        );
        
        CREATE TABLE IF NOT EXISTS chat (
            chat_id INTEGER,  
            username VARCHAR(50),
            message VARCHAR(100),
            created_at INTEGER,
            image VARCHAR(100),
            audio_data BYTEA
        );

        CREATE INDEX IF NOT EXISTS idx_created_at_chat_id ON chat (created_at, chat_id); 
       `)

	if err != nil {
		return fmt.Errorf("ошибка при создании таблиц: %w", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return database
}

func Close() {
	if database != nil {
		database.Close()
	}
}
