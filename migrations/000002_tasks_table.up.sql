CREATE TABLE IF NOT EXISTS tasks (
  id int PRIMARY KEY auto_increment,
  created_at DATETIME default CURRENT_TIMESTAMP,
  title text NOT NULL,
  created_by text NOT NULL,
  version int NOT NULL DEFAULT 1
)
