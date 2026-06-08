use actix_files::Files;
use actix_web::{
    delete, get, post, put,
    web::{Data, Json, Path},
    App, HttpResponse, HttpServer, Responder,
};
use rusqlite::{params, Connection};
use serde::{Deserialize, Serialize};
use std::sync::Mutex;

// ── models ────────────────────────────────────────────────────────────────────

#[derive(Serialize)]
struct Todo {
    id: i64,
    title: String,
    done: bool,
}

#[derive(Deserialize)]
struct CreateTodo {
    title: String,
}

#[derive(Deserialize)]
struct UpdateTodo {
    title: Option<String>,
    done: Option<bool>,
}

// ── state ─────────────────────────────────────────────────────────────────────

struct AppState {
    db: Mutex<Connection>,
}

// ── handlers ──────────────────────────────────────────────────────────────────

#[get("/api/todos")]
async fn list(state: Data<AppState>) -> impl Responder {
    let db = state.db.lock().unwrap();
    let mut stmt = db
        .prepare("SELECT id, title, done FROM todos ORDER BY id DESC")
        .unwrap();
    let todos: Vec<Todo> = stmt
        .query_map([], |row| {
            Ok(Todo {
                id: row.get(0)?,
                title: row.get(1)?,
                done: row.get::<_, i64>(2)? == 1,
            })
        })
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();
    HttpResponse::Ok().json(todos)
}

#[post("/api/todos")]
async fn create(state: Data<AppState>, body: Json<CreateTodo>) -> impl Responder {
    let title = body.title.trim().to_string();
    if title.is_empty() {
        return HttpResponse::BadRequest().json(serde_json::json!({"error": "title required"}));
    }
    let db = state.db.lock().unwrap();
    db.execute("INSERT INTO todos (title) VALUES (?1)", params![title])
        .unwrap();
    let id = db.last_insert_rowid();
    HttpResponse::Created().json(Todo { id, title, done: false })
}

#[put("/api/todos/{id}")]
async fn update(state: Data<AppState>, id: Path<i64>, body: Json<UpdateTodo>) -> impl Responder {
    let id = id.into_inner();
    let db = state.db.lock().unwrap();
    if let Some(ref title) = body.title {
        db.execute("UPDATE todos SET title=?1 WHERE id=?2", params![title, id])
            .unwrap();
    }
    if let Some(done) = body.done {
        db.execute(
            "UPDATE todos SET done=?1 WHERE id=?2",
            params![done as i64, id],
        )
        .unwrap();
    }
    let todo = db
        .query_row(
            "SELECT id, title, done FROM todos WHERE id=?1",
            params![id],
            |row| {
                Ok(Todo {
                    id: row.get(0)?,
                    title: row.get(1)?,
                    done: row.get::<_, i64>(2)? == 1,
                })
            },
        )
        .ok();
    match todo {
        Some(t) => HttpResponse::Ok().json(t),
        None => HttpResponse::NotFound().json(serde_json::json!({"error": "not found"})),
    }
}

#[delete("/api/todos/{id}")]
async fn remove(state: Data<AppState>, id: Path<i64>) -> impl Responder {
    let db = state.db.lock().unwrap();
    db.execute("DELETE FROM todos WHERE id=?1", params![id.into_inner()])
        .unwrap();
    HttpResponse::NoContent().finish()
}

// ── main ──────────────────────────────────────────────────────────────────────

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let conn = Connection::open("todos.db").expect("failed to open db");
    conn.execute_batch(
        "CREATE TABLE IF NOT EXISTS todos (
            id    INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT    NOT NULL,
            done  INTEGER NOT NULL DEFAULT 0
        )",
    )
    .unwrap();

    let state = Data::new(AppState {
        db: Mutex::new(conn),
    });

    println!("Rust todo server running on http://localhost:8081");

    HttpServer::new(move || {
        App::new()
            .app_data(state.clone())
            .service(list)
            .service(create)
            .service(update)
            .service(remove)
            .service(Files::new("/", "static").index_file("index.html"))
    })
    .bind("127.0.0.1:8081")?
    .run()
    .await
}
