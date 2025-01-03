CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "users" (
  "id" UUID PRIMARY KEY DEFAULT(uuid_generate_v4()),
  "username" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT(now()),  
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "entries" (
  "id" UUID PRIMARY KEY DEFAULT(uuid_generate_v4()),
  "user_id" UUID NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "documents" (
  "id" UUID PRIMARY KEY DEFAULT(uuid_generate_v4()),
  "entry_id" UUID NOT NULL,
  "filename" varchar NOT NULL
);

ALTER TABLE "entries" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "documents" ADD FOREIGN KEY ("entry_id") REFERENCES "entries" ("id");

CREATE INDEX ON "users" ("username");
CREATE INDEX ON "users" ("email");

CREATE INDEX ON "users" ("full_name");

CREATE INDEX ON "users" ("id");

CREATE INDEX ON "entries" ("id");

CREATE INDEX ON "documents" ("id");

CREATE INDEX ON "documents" ("filename");
