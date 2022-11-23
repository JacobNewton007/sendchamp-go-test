CREATE TABLE IF NOT EXISTS tokens (
  id int PRIMARY KEY auto_increment,
  hash BLOB,
  user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
  expiry timestamp(0) NOT NULL,
  scope text NOT NULL
)