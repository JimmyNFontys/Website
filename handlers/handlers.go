package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Globale variabele voor de databaseverbinding
var db *sql.DB
var bundle *i18n.Bundle

// Initialiseer de vertalingen
func InitI18n() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Laad vertalingen in alle ondersteunde talen
	bundle.MustLoadMessageFile("locales/en.json")
	bundle.MustLoadMessageFile("locales/nl.json")
	bundle.MustLoadMessageFile("locales/de.json")
}

// SetDB instellen van de globale databaseverbinding
func SetDB(database *sql.DB) {
	db = database
}

// HomeHandler rendert de homepagina
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Kies de taal op basis van de voorkeur van de gebruiker
	langs := []language.Tag{
		language.English, // Fallback als de voorkeurstaal niet wordt ondersteund
		language.Dutch,
		language.German,
	}
	matcher := language.NewMatcher(langs)
	accept := r.Header.Get("Accept-Language")
	tag, _, _ := language.ParseAcceptLanguage(accept)
	found, _, _ := matcher.Match(tag...)

	// Maak een vertaler met de gekozen taal
	translator := i18n.NewLocalizer(bundle, found.String())

	// Vertaal de tekst voor de homepage
	welcome, _ := translator.Localize(&i18n.LocalizeConfig{MessageID: "welcome"})
	description, _ := translator.Localize(&i18n.LocalizeConfig{MessageID: "description"})
	button, _ := translator.Localize(&i18n.LocalizeConfig{MessageID: "button"})

	// Lees het inhoud van home.html
	tmpl, err := os.ReadFile("templates/home.html")
	if err != nil {
		log.Fatal(err)
	}

	// Parseer de HTML-template
	t, err := template.New("home").Parse(string(tmpl))
	if err != nil {
		log.Fatal(err)
	}

	// Definieer een struct voor het invoegen van vertaalde tekst in de template
	type TemplateData struct {
		Lang        string
		Welcome     string
		Description string
		Button      string
	}

	// Maak een struct met vertaalde tekst
	data := TemplateData{
		Lang:        found.String(),
		Welcome:     welcome,
		Description: description,
		Button:      button,
	}

	// Render de template met de vertaalde tekst
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		log.Fatal(err)
	}

	// Stuur de gerenderde HTML-respons terug naar de client
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// ReservationPageHandler rendert de reservatiepagina.
func ReservationPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/reservation.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// ReserveHandler behandelt reserveringen
func ReserveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Eerst de formdata parsen
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verzamel form data, inclusief activiteiten als boolean waarde
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	checkInDate := r.FormValue("check_in_date")
	checkOutDate := r.FormValue("check_out_date")
	accommodation := r.FormValue("accommodation")

	activityRestaurant := r.FormValue("activity_restaurant") == "1"
	dateActivityRestaurant := sql.NullString{String: r.FormValue("date_activity_restaurant"), Valid: activityRestaurant && r.FormValue("date_activity_restaurant") != ""}
	activityRestaurantTime := r.FormValue("activity_restaurant_time")

	activityBowling := r.FormValue("activity_bowling") == "1"
	dateActivityBowling := sql.NullString{String: r.FormValue("date_activity_bowling"), Valid: activityBowling && r.FormValue("date_activity_bowling") != ""}
	activityBowlingTime := r.FormValue("activity_bowling_time")

	activityBicycle := r.FormValue("activity_bicycle") == "1"
	dateActivityBicycle := sql.NullString{String: r.FormValue("date_activity_bicycle"), Valid: activityBicycle && r.FormValue("date_activity_bicycle") != ""}
	activityBicycleTime := r.FormValue("activity_bicycle_time")

	location := r.FormValue("location")
	licensePlate := r.FormValue("license_plate")

	// Voer de insert uit met de nieuwe velden
	_, err := db.Exec("INSERT INTO reservations (name, surname, email, phone, check_in_date, check_out_date, accommodation, activity_restaurant, date_activity_restaurant, activity_restaurant_time, activity_bowling, date_activity_bowling, activity_bowling_time, activity_bicycle, date_activity_bicycle, activity_bicycle_time, location, license_plate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		name, surname, email, phone, checkInDate, checkOutDate, accommodation, activityRestaurant, dateActivityRestaurant, activityRestaurantTime, activityBowling, dateActivityBowling, activityBowlingTime, activityBicycle, dateActivityBicycle, activityBicycleTime, location, licensePlate)

	if err != nil {
		log.Printf("Fout bij het opslaan van de reservering: %v", err)
		http.Error(w, "Er is een fout opgetreden bij het opslaan van de reservering", http.StatusInternalServerError)
		return
	}

	// Stuur een bevestigingsemail naar de klant
	// (Aangenomen dat de implementatie van sendConfirmationEmail hetzelfde blijft)
	if err := sendConfirmationEmail(email, name, checkInDate, checkOutDate, accommodation); err != nil {
		http.Error(w, "Er is een fout opgetreden bij het versturen van de bevestigingsemail: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

// Behandeld confirmation pagina met de gegeven reserveringsdetails.
func ConfirmationPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/confirmation.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// loginAuth is een type dat de smtp.Auth interface implementeert
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
		}
	}
	return nil, nil
}

func sendConfirmationEmail(recipient, name, checkInDate, checkOutDate, accommodation string) error {
	from := "holiday_parks@hotmail.com"
	password := "h0lidayp@rks!"

	// Definieer de SMTP server details
	smtpHost := "smtp-mail.outlook.com"
	smtpPort := "587"

	// Stel het bericht op dat je wilt versturen
	message := []byte("To: " + recipient + "\r\n" +
		"Subject: Reservation Confirmation\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"Dear " + name + ",\n\n" +
		"Thank you for your reservation" + "\n" +
		"From: " + checkInDate + " to " + checkOutDate + "\n" +
		"For: " + accommodation + "\n\n" +
		"See you soon!\n\n" +
		"Team Holidayparks")

	// Instellen van authenticatie
	auth := LoginAuth(from, password)

	// Verstuur de email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort, // SMTP server adres
		auth,                  // Authenticatie
		from,                  // Afzender adres
		[]string{recipient},   // Ontvanger adres
		message,               // Bericht body
	)

	if err != nil {
		return err
	}
	return nil
}
