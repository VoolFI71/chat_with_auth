package db

import (
    "database/sql"
    "fmt"
    //"log"
)

var database *sql.DB

func Connect() error {
    var err error
    database, err = sql.Open("postgres", "postgresql://postgres:1234@db:5432/go?sslmode=disable")
    if err != nil {
        return err
    }

    if err := database.Ping(); err != nil {
        return fmt.Errorf("Ошибка при проверке соединения: %w", err)
    }

    _, err = database.Exec(`
        CREATE TABLE IF NOT EXISTS g (
            username VARCHAR(50) UNIQUE,
            password VARCHAR(100),
            balance DECIMAL(10, 2),
            email VARCHAR(50) UNIQUE
        );
        
        CREATE TABLE IF NOT EXISTS chat (
            username VARCHAR(50),
            message VARCHAR(100),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            image BYTEA
        );

        CREATE INDEX IF NOT EXISTS idx_created_at ON chat (created_at);
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
