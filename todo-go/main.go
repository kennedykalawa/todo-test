package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "todos.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id    INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT    NOT NULL,
		done  INTEGER NOT NULL DEFAULT 0
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/todos")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query("SELECT id, title, done FROM todos ORDER BY id DESC")
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer rows.Close()
			var todos []Todo
			for rows.Next() {
				var t Todo
				var done int
				rows.Scan(&t.ID, &t.Title, &done)
				t.Done = done == 1
				todos = append(todos, t)
			}
			if todos == nil {
				todos = []Todo{}
			}
			json.NewEncoder(w).Encode(todos)

		case http.MethodPost:
			var body struct{ Title string `json:"title"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, `{"error":"invalid json"}`, 400)
				return
			}
			if body.Title == "" {
				http.Error(w, `{"error":"title required"}`, 400)
				return
			}
			res, err := db.Exec("INSERT INTO todos (title) VALUES (?)", body.Title)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			id, _ := res.LastInsertId()
			json.NewEncoder(w).Encode(Todo{ID: int(id), Title: body.Title, Done: false})

		default:
			http.Error(w, "method not allowed", 405)
		}
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var body struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid json"}`, 400)
			return
		}

		// Check if exists
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM todos WHERE id=?", id).Scan(&exists)
		if exists == 0 {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}

		if body.Title != nil {
			db.Exec("UPDATE todos SET title=? WHERE id=?", *body.Title, id)
		}
		if body.Done != nil {
			val := 0
			if *body.Done {
				val = 1
			}
			db.Exec("UPDATE todos SET done=? WHERE id=?", val, id)
		}

		var t Todo
		var done int
		err := db.QueryRow("SELECT id, title, done FROM todos WHERE id=?", id).Scan(&t.ID, &t.Title, &done)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		t.Done = done == 1
		json.NewEncoder(w).Encode(t)

	case http.MethodDelete:
		res, err := db.Exec("DELETE FROM todos WHERE id=?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", 405)
	}
}

func main() {
	initDB()
	defer db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/todos", todosHandler)
	http.HandleFunc("/api/todos/", todosHandler)

	log.Println("Go todo server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
