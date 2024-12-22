-- Step 1: Add a new `id` column to `users` as a BIGSERIAL
ALTER TABLE users ADD COLUMN id UUID DEFAULT uuid_generate_v4();

UPDATE users SET id = uuid_generate_v4();

-- Step 2: Drop the foreign key constraint in `entries` that references `users.username`
ALTER TABLE entries DROP CONSTRAINT entries_user_username_fkey;

-- Step 3: Add a new `user_id` column to `entries`
ALTER TABLE entries ADD COLUMN user_id UUID;

-- Step 4: Populate the `user_id` column in `entries` based on the `username` in `users`
UPDATE entries
SET user_id = (SELECT id FROM users WHERE users.username = entries.user_username);

-- Step 5: Drop the old primary key constraint on `users.username`
ALTER TABLE users DROP CONSTRAINT users_pkey;

-- Step 6: Set `id` as the new primary key for `users`
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Step 7: Add a new foreign key constraint in `entries` referencing `users.id`
ALTER TABLE entries
ADD CONSTRAINT entries_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

-- Step 8: Drop the old `user_username` column in `entries`
ALTER TABLE entries DROP COLUMN user_username;

ALTER TABLE users ADD COLUMN role varchar;


