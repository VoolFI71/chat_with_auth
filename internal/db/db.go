package db

import (
    "context"
    "github.com/jackc/pgx/v4"
    "log"
)

var conn *pgx.Conn

func Connect() {
    var err error
    conn, err = pgx.Connect(context.Background(), "postgres://postgres:1234@localhost:5432/go")
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }
}

// GetConnection возвращает текущее соединение
func GetConnection() *pgx.Conn {
    return conn
}