CREATE TABLE IF NOT EXISTS users (
  id int PRIMARY KEY auto_increment,
  created_at DATETIME default CURRENT_TIMESTAMP,
  name text NOT NULL,
  email varchar(255) UNIQUE NOT NULL,
  password_hash text NOT NULL,
  activated tinyint NOT NULL,
  version int NOT NULL DEFAULT 1
);

