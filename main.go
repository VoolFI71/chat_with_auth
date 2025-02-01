package main

import (
	"fmt"
    "net/http"
    "encoding/json"
    "database/sql"
    _ "github.com/jackc/pgx/v4/stdlib"
    "log"
)

type Subj struct {
    Product string `json:"name"`
    Price   int    `json:"price"`
 }
 
func JSONHandler(w http.ResponseWriter, req *http.Request) {
    // собираем данные
    subj := Subj{"Milk", 50}
    // кодируем в JSON
    resp, err := json.Marshal(subj)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    // устанавливаем заголовок Content-Type
    // для передачи клиенту информации, кодированной в JSON
    w.Header().Set("content-type", "application/json")
    // устанавливаем код 200
    w.WriteHeader(http.StatusOK)
    // пишем тело ответа
    w.Write(resp)
} 

func mainPage(res http.ResponseWriter, req *http.Request) {
    
    body := fmt.Sprintf("Method: %s\r\n", req.Method)
    body += "Header ===============\r\n"
    for k, v := range req.Header {
        body += fmt.Sprintf("%s: %v\r\n", k, v)
    }
    body += "Query parameters ===============\r\n"

    err := req.ParseForm()
    if err != nil {
        http.Error(res, "Unable to parse form", http.StatusBadRequest)
        return
    }

    for k, v := range req.Form {
        body += fmt.Sprintf("%s: %v\r\n", k, v)
    }
    res.Write([]byte(body))
} 

func main() {
    ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",`localhost`, `postgres`, `1234`, `go`)

    db, err := sql.Open("pgx", ps)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) NOT NULL UNIQUE
    );`

    // Выполняем запрос на создание таблицы
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal(err)
    }
    
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    mux := http.NewServeMux()
    mux.HandleFunc(`/`, mainPage)
    mux.HandleFunc(`/a`, JSONHandler)
    err = http.ListenAndServe(`:8080`, mux)
    if err != nil {
        panic(err)
    }
} 