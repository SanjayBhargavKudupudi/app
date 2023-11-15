package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

var (
	oauth2Config = &oauth2.Config{
		ClientID:     "Expressnest",                      // Replace with your client ID from Keycloak
		ClientSecret: "JkU6MpovsEFBNZZVTwYtr0sl4tC8milc", // Replace with your client secret from Keycloak
		RedirectURL:  "http://localhost:3000/",           // Replace with your redirect URI
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://keycloak-server/auth/realms/Nest/protocol/openid-connect/auth",
			TokenURL: "http://keycloak-server/auth/realms/Nest/protocol/openid-connect/token",
		},
	}
	// Placeholder for the database connection
	db *sql.DB
)

func main() {
	var err error
	db, err = setupDB()
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/signin", signInHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/register", registerUserHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func setupDB() (*sql.DB, error) {
	connStr := "postgres://username:0116@localhost/dbname?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Exchange the received code for a token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Token contains the access and refresh tokens
	// Handle the tokens appropriately (e.g., create session, store tokens)
	fmt.Fprintf(w, "Access Token: %s", token.AccessToken) // Placeholder
}

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Add validation for username and password as per your requirements

	if err := registerUser(db, username, password); err != nil {
		http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User registered successfully")
}

func registerUser(db *sql.DB, username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user into the database
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, string(hashedPassword))
	return err
}
