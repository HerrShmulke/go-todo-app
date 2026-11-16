DELETE FROM todos;
ALTER TABLE todos
ADD COLUMN user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX todos_user_id_idx ON todos(user_id);