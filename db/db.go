package db

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
)

func NewMySQLStorage(cfg mysql.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal()
	}

	return db, nil
}

func InitDatabase(db *sql.DB) {
	if err := db.Ping(); err != nil {
		log.Fatal()
	}
	log.Println("Database has connected successfully")
}
