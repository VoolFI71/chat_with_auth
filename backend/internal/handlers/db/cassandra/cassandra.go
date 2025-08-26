package cassandra

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
)

type DB struct {
	Session *gocql.Session
}

// NewDB создает новое подключение к Cassandra
func NewDB(host string, keyspace string) *DB {
	if envHost := os.Getenv("CASSANDRA_HOST"); envHost != "" {
		host = envHost
	}
	cluster := gocql.NewCluster(host) // Используем переданный хост
	cluster.Port = 9042               // Порт по умолчанию

	// Дадим Cassandra время подняться в docker-compose
	time.Sleep(3 * time.Second)

	// Сессия без keyspace для его создания при необходимости
	adminSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	createKeyspaceCQL := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`, keyspace)
	if err := adminSession.Query(createKeyspaceCQL).Exec(); err != nil {
		log.Println("failed to ensure keyspace:", err)
	}
	adminSession.Close()

	// Подключаемся к нужному keyspace
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы сообщений
	if err := session.Query(`CREATE TABLE IF NOT EXISTS messages (
        chat_id int,
        username text,
        message text,
        created_at bigint,
        image text,
        audio_data text,
        PRIMARY KEY ((chat_id), created_at)
    ) WITH CLUSTERING ORDER BY (created_at DESC)`).Exec(); err != nil {
		log.Println("failed to ensure messages table:", err)
	}

	return &DB{Session: session} // Возвращаем указатель на структуру DB
}

// Close закрывает сессию базы данных
func (db *DB) Close() {
	db.Session.Close()
}
