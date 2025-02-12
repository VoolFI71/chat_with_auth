package db

import (
    "database/sql"
    "log"
	"fmt"
)

func ConnectAuth() (*sql.DB, error) {
	database, err := sql.Open("postgres", "postgresql://postgres:1234@db:5432/go?sslmode=disable") 
    if err != nil {
		//panic(err)
        return nil, err
    }

    if err := database.Ping(); err != nil {
        log.Fatal("Ошибка при проверке соединения:", err)
		defer database.Close()
		return nil, fmt.Errorf("Ошибка при проверке соединения: %w", err)
    }

    //_, err = database.Exec(`DROP TABLE g`)

    _, err = database.Exec(`CREATE TABLE IF NOT EXISTS g (
        username VARCHAR(50) UNIQUE,
        password VARCHAR(100),
        balance DECIMAL(10, 2),
        email VARCHAR(50) UNIQUE
    )`)
    if err != nil {
        defer database.Close()
        return nil, err
    }

    return database, nil
}

func ConnectChat() (*sql.DB, error) {
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

    //_, err = database.Exec(`DROP TABLE chat`)

    _, err = database.Exec(`CREATE TABLE IF NOT EXISTS chat (
        username VARCHAR(50),
        message VARCHAR(100),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        defer database.Close()
        return nil, err
    }

    _, err = database.Exec("CREATE INDEX idx_created_at ON chat (created_at)")
    if err != nil {
        fmt.Println("Error creating index:", err)
        return nil, err
    }
    return database, nil
}