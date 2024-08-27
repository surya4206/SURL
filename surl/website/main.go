package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to the database
	var err error
	db, err = sql.Open("postgres", "postgresql://surl:surl@localhost/urlshortener?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS shortened_urls (
			short_code TEXT PRIMARY KEY,
			original_url TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/r/", handleRedirect)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>URL Shortener</title>
</head>
<body>
    <h1>URL Shortener</h1>
    <form action="/shorten" method="post">
        <input type="url" name="url" placeholder="Enter URL" required>
        <input type="submit" value="Shorten">
    </form>
</body>
</html>
`
	t, _ := template.New("index").Parse(tmpl)
	t.Execute(w, nil)
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode()

	_, err := db.Exec("INSERT INTO shortened_urls (short_code, original_url) VALUES ($1, $2)", shortCode, url)
	if err != nil {
		log.Printf("Error inserting URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("http://%s/r/%s", r.Host, shortCode)
	fmt.Fprintf(w, "Shortened URL: <a href=\"%s\">%s</a>", shortURL, shortURL)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[len("/r/"):]

	var url string
	err := db.QueryRow("SELECT original_url FROM shortened_urls WHERE short_code = $1", shortCode).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("Error querying database: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func generateShortCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:6]
}
