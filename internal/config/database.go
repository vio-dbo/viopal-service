package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQLConnection() *sql.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("failed open db 1:", err)
	}

	// pool config
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// verify connection
	if err := db.Ping(); err != nil {
		log.Println("connecting to:", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
		log.Fatal("failed ping db 2:", err)
	}

	log.Println("MySQL connected successfully")

	return db
}
