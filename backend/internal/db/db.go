package db

import (
    "database/sql"
    "log"
	"fmt"
)

func Connect() (*sql.DB, error) {
	database, err := sql.Open("postgres", "postgresql://postgres:1234@db:5432/go?sslmode=disable") 
    if err != nil {
		//panic(err)
        return nil, err
    }
    //defer database.Close()

    if err := database.Ping(); err != nil {
        log.Fatal("Ошибка при проверке соединения:", err)
		database.Close()
		return nil, fmt.Errorf("Ошибка при проверке соединения: %w", err)
    }

    _, err = database.Exec(`CREATE TABLE IF NOT EXISTS g (
        username VARCHAR(50) UNIQUE,
        password VARCHAR(100),
        balance DECIMAL(10, 2)
    )`)
    if err != nil {
        database.Close()
        return nil, fmt.Errorf("Ошибка при создании таблицы: %w", err)
    }

    return database, nil
}