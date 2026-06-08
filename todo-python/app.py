import sqlite3
from flask import Flask, request, jsonify, send_from_directory, g

app = Flask(__name__, static_folder='static')
DB = 'todos.db'


def get_db():
    if 'db' not in g:
        g.db = sqlite3.connect(DB)
        g.db.row_factory = sqlite3.Row
    return g.db


@app.teardown_appcontext
def close_db(error):
    db = g.pop('db', None)
    if db is not None:
        db.close()


def init_db():
    with sqlite3.connect(DB) as db:
        db.execute('''
            CREATE TABLE IF NOT EXISTS todos (
                id    INTEGER PRIMARY KEY AUTOINCREMENT,
                title TEXT    NOT NULL,
                done  INTEGER NOT NULL DEFAULT 0
            )
        ''')


@app.route('/')
def index():
    return send_from_directory('static', 'index.html')


@app.route('/api/todos', methods=['GET'])
def get_todos():
    db = get_db()
    rows = db.execute('SELECT * FROM todos ORDER BY id DESC').fetchall()
    return jsonify([{'id': r['id'], 'title': r['title'], 'done': bool(r['done'])} for r in rows])


@app.route('/api/todos', methods=['POST'])
def create_todo():
    data = request.get_json()
    title = (data or {}).get('title', '').strip()
    if not title:
        return jsonify({'error': 'title required'}), 400
    db = get_db()
    cur = db.execute('INSERT INTO todos (title) VALUES (?)', (title,))
    db.commit()
    todo_id = cur.lastrowid
    return jsonify({'id': todo_id, 'title': title, 'done': False}), 201


@app.route('/api/todos/<int:todo_id>', methods=['PUT'])
def update_todo(todo_id):
    data = request.get_json() or {}
    db = get_db()
    
    # Check if exists
    row = db.execute('SELECT * FROM todos WHERE id=?', (todo_id,)).fetchone()
    if row is None:
        return jsonify({'error': 'not found'}), 404

    if 'title' in data:
        db.execute('UPDATE todos SET title=? WHERE id=?', (data['title'], todo_id))
    if 'done' in data:
        db.execute('UPDATE todos SET done=? WHERE id=?', (int(data['done']), todo_id))
    db.commit()
    
    row = db.execute('SELECT * FROM todos WHERE id=?', (todo_id,)).fetchone()
    return jsonify({'id': row['id'], 'title': row['title'], 'done': bool(row['done'])})


@app.route('/api/todos/<int:todo_id>', methods=['DELETE'])
def delete_todo(todo_id):
    db = get_db()
    db.execute('DELETE FROM todos WHERE id=?', (todo_id,))
    db.commit()
    return '', 204


if __name__ == '__main__':
    init_db()
    print('Python todo server running on http://localhost:5000')
    app.run(port=5000, debug=True)
