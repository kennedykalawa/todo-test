const express = require('express');
const Database = require('better-sqlite3');
const path = require('path');

const app = express();
const db = new Database('todos.db');
const PORT = 3000;

// Init DB
db.exec(`
  CREATE TABLE IF NOT EXISTS todos (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT    NOT NULL,
    done  INTEGER NOT NULL DEFAULT 0
  )
`);

app.use(express.json());
app.use(express.static(path.join(__dirname, 'static')));

// GET all todos
app.get('/api/todos', (req, res) => {
  try {
    const todos = db.prepare('SELECT * FROM todos ORDER BY id DESC').all();
    res.json(todos.map(t => ({ ...t, done: t.done === 1 })));
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// POST create todo
app.post('/api/todos', (req, res) => {
  const { title } = req.body;
  if (!title?.trim()) return res.status(400).json({ error: 'title required' });
  try {
    const result = db.prepare('INSERT INTO todos (title) VALUES (?)').run(title.trim());
    res.status(201).json({ id: result.lastInsertRowid, title: title.trim(), done: false });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// PUT update todo
app.put('/api/todos/:id', (req, res) => {
  const id = parseInt(req.params.id);
  const { title, done } = req.body;
  
  try {
    const todo = db.prepare('SELECT * FROM todos WHERE id=?').get(id);
    if (!todo) return res.status(404).json({ error: 'not found' });

    if (title !== undefined) {
      db.prepare('UPDATE todos SET title=? WHERE id=?').run(title, id);
    }
    if (done !== undefined) {
      db.prepare('UPDATE todos SET done=? WHERE id=?').run(done ? 1 : 0, id);
    }
    
    const updatedTodo = db.prepare('SELECT * FROM todos WHERE id=?').get(id);
    res.json({ ...updatedTodo, done: updatedTodo.done === 1 });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// DELETE todo
app.delete('/api/todos/:id', (req, res) => {
  const id = parseInt(req.params.id);
  try {
    const result = db.prepare('DELETE FROM todos WHERE id=?').run(id);
    if (result.changes === 0) return res.status(404).json({ error: 'not found' });
    res.status(204).send();
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.listen(PORT, () => {
  console.log(`Node todo server running on http://localhost:${PORT}`);
});
