package main

import (
	"Webapp1/handlers" // Importeer je handlers package
	"bufio"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Importeer de MySQL driver
)

var db *sql.DB

// Lees configuratie uit een bestand
func readConfig(filename string) (map[string]string, error) {
	config := make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error when opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error when scanning file: %v", err)
		return nil, err
	}

	return config, nil
}

// Initialiseer de database
func initDatabase() {
	var err error
	config, err := readConfig("config.txt")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	dsn := config["username"] + ":" + config["password"] + "@tcp(" + config["ip"] + ":" + config["port"] + ")/" + config["databaseName"]
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Error when trying to ping database: %v", err)
	}
}

func main() {
	// Initialiseer de database
	initDatabase()
	defer db.Close()

	// Initialiseer i18n
	handlers.InitI18n()

	// Stel de globale databaseverbinding in de handlers in
	handlers.SetDB(db)

	fs := http.FileServer(http.Dir("./image"))
	http.Handle("/image/", http.StripPrefix("/image/", fs))

	// Setup de HTTP-routers en handlers
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/reservation", handlers.ReservationPageHandler)
	http.HandleFunc("/reserve", handlers.ReserveHandler)
	http.HandleFunc("/confirmation", handlers.ConfirmationPageHandler)
	http.Handle("/locales/", http.StripPrefix("/locales/", http.FileServer(http.Dir("./locales"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("./scripts"))))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
