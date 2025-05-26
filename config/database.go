package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	connString := "sqlserver://admin:Celcius@1980@103.102.153.62:1433?database=codetech"

	DB, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error membuka koneksi:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Tidak bisa connect:", err)
	}

	fmt.Println("Connected to SQL Server! ✅")
}
