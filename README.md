# todo-fullstack

Same Todo CRUD app in 4 stacks. SQLite + plain HTML/JS frontend (served by each backend).

## API shape (identical across all)

| Method | Path            | Body                     | Description  |
|--------|-----------------|--------------------------|--------------|
| GET    | /api/todos      | —                        | List all     |
| POST   | /api/todos      | `{ "title": "..." }`     | Create       |
| PUT    | /api/todos/:id  | `{ "title"?, "done"? }`  | Update       |
| DELETE | /api/todos/:id  | —                        | Delete       |

---

## Go  →  port 8080

**Stack:** `net/http` (stdlib) + `modernc.org/sqlite` (pure Go, no CGO)

```bash
cd todo-go
go mod tidy
go run main.go
# open http://localhost:8080
```

---

## Node.js  →  port 3000

**Stack:** Express + `better-sqlite3`

```bash
cd todo-node
npm install
npm start
# open http://localhost:3000

# or with auto-reload:
npm run dev
```

---

## Python  →  port 5000

**Stack:** Flask + `sqlite3` (stdlib)

```bash
cd todo-python
pip install -r requirements.txt
python app.py
# open http://localhost:5000
```

With a venv:
```bash
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
python app.py
```

---

## Rust  →  port 8081

**Stack:** Actix-web + `rusqlite` (bundled SQLite, no system dep needed)

```bash
cd todo-rust
cargo run
# open http://localhost:8081
# first run compiles ~30s
```

---

## Project layout

```
todo-go/
  main.go
  go.mod
  static/index.html
  todos.db          ← created at runtime

todo-node/
  server.js
  package.json
  static/index.html
  todos.db

todo-python/
  app.py
  requirements.txt
  static/index.html
  todos.db

todo-rust/
  Cargo.toml
  src/main.rs
  static/index.html
  todos.db
```

Each project is fully self-contained. The `todos.db` file is created automatically on first run.
