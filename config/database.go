package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	connString := "user=postgres password=admin@2004 host=127.0.0.1 port=5432 dbname=codetech sslmode=disable"

	DB, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("Error membuka koneksi:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Tidak bisa connect:", err)
	}

	fmt.Println("Connected to SQL Server! ✅")
}
