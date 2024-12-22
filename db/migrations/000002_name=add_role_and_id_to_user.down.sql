ALTER TABLE entries DROP CONSTRAINT entries_user_id_fkey;

ALTER TABLE entries ADD COLUMN user_username varchar;


ALTER TABLE users DROP CONSTRAINT users_pkey;

ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (username);

ALTER TABLE entries
ADD CONSTRAINT entries_user_username_fkey FOREIGN KEY (user_username) REFERENCES users (username);


UPDATE entries
SET user_username = (SELECT username FROM users WHERE users.id = entries.user_id);


ALTER TABLE entries DROP COLUMN user_id;

ALTER TABLE users DROP COLUMN id;

ALTER TABLE users DROP COLUMN role;