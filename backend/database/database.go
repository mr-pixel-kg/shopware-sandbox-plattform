package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // PostgreSQL Treiber
	_ "github.com/mattn/go-sqlite3" // SQLite Treiber
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	"log"
)

var DB *sqlx.DB

func ConnectDB(config config.DatabaseConfig) {
	dsn, driver := parseDSN(config)

	// Datenbankverbindung öffnen
	var err error
	DB, err = sqlx.Open(driver, dsn)
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der Datenbank: %v", err)
	}

	// Verbindung testen
	if err = DB.Ping(); err != nil {
		log.Fatalf("Fehler beim Pingen der Datenbank: %v", err)
	}

	log.Println("Datenbank erfolgreich verbunden!")

	// Tabellen erstellen, falls sie noch nicht existieren
	createTables()
}

func parseDSN(config config.DatabaseConfig) (string, string) {
	dsn := ""
	driver := ""

	if config.Host == "" && config.Port == 0 && config.User == "" && config.Password == "" && config.Name == "" {
		// SQLite
		log.Println("Database configuration is empty, so we use a SQLite database")
		dsn = "file:database.db"
		driver = "sqlite3"
	} else {
		// PostgreSQL
		log.Println("Database configuration is loaded")
		dsn = fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
			config.User, config.Password, config.Name, config.Host, config.Port)
		driver = "postgres"
	}

	return dsn, driver
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS images (
		id VARCHAR(255) PRIMARY KEY,
		image_name VARCHAR(128) NOT NULL,
		image_tag VARCHAR(32) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS sandboxes (
		id VARCHAR(255) PRIMARY KEY,
		container_id VARCHAR(255) NOT NULL,
	    container_name VARCHAR(64) NOT NULL,
	    image_id VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    destroy_at TIMESTAMP DEFAULT NULL,
		FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ip_address VARCHAR(16) NOT NULL,
		user_agent VARCHAR(255) NOT NULL,
		username VARCHAR(64) DEFAULT NULL,
		action VARCHAR(16) NOT NULL,
		details JSON DEFAULT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    ip_address VARCHAR(16) NOT NULL,
	    user_agent VARCHAR(255) NOT NULL,
	    username VARCHAR(64) DEFAULT NULL,
		sandbox_id VARCHAR(255) NOT NULL,
		FOREIGN KEY(sandbox_id) REFERENCES sandboxes(id) ON DELETE CASCADE
	);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		log.Fatalf("Fehler beim Erstellen der Tabellen: %v", err)
	}
}
