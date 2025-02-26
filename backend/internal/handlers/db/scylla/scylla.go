package scylla

import (
    "github.com/gocql/gocql"
    "log"
    "time"

)

type DB struct {
    Session *gocql.Session
}

func NewDB(host string, keyspace string) (*DB, error) {
    // Создаем кластер и подключаемся к ScyllaDB
    time.Sleep(11 * time.Second) // Задержка перед подключением

    cluster := gocql.NewCluster(host)
    cluster.Keyspace = "system" // Подключаемся к системному ключевому пространству для выполнения DDL-запросов
    session, err := cluster.CreateSession()
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // Создаем ключевое пространство
    query := `CREATE KEYSPACE IF NOT EXISTS ` + keyspace + ` WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};`
    
    if err := session.Query(query).Exec(); err != nil {
        log.Fatalf("Ошибка при создании ключевого пространства: %v", err)
    }

    log.Printf("Ключевое пространство '%s' успешно создано.", keyspace)

    // Создаем новую сессию с новым ключевым пространством
    cluster.Keyspace = keyspace
    session, err = cluster.CreateSession()
    if err != nil {
        log.Fatal(err)
    }

    return &DB{Session: session}, nil
}

func (db *DB) Close() {
    db.Session.Close()
}
