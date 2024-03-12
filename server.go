package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type User struct {
	Username string
	Password string
}

type ViewData struct {
	Msg string
}

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	fsTraversal := http.FileServer(http.Dir("directorytraversal"))
	http.Handle("/traversal/", http.StripPrefix("/traversal/", fsTraversal))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/crash", crashHandler)
	log.Fatal(http.ListenAndServe(":3100", nil))
}

func crashHandler(w http.ResponseWriter, r *http.Request) {
	log.Fatal(fmt.Errorf("crasheo forzado del servidor"))
}

func OpenDatabase(databaseName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databaseName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		INSERT OR IGNORE INTO Users (username, password) VALUES ('root', 'vulnerable');
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func authenticateUser(w http.ResponseWriter, username, password string) bool {
	db, err := OpenDatabase("database.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	defer db.Close()

	// Consulta hecha a propósito para ser vulnerable a SQLi
	var count int
	query := fmt.Sprintf("SELECT count(*) FROM Users WHERE username = '%s' AND password = '%s'", username, password)
	err = db.QueryRow(query).Scan(&count)

	return err == nil && count > 0
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	data := ViewData{Msg: ""}
	tmpl, err := template.ParseFiles("templates/index.html")

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		if authenticateUser(w, username, password) {
			tmpl, _ = template.ParseFiles("templates/success.html")
		} else {
			data.Msg = "*Usuario o contraseña incorrectos"
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
