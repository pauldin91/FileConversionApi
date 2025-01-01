-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2025-01-01T13:09:29.647Z

CREATE TABLE "users" (
  "username" varchar NOT NULL,
  "id" uuid PRIMARY KEY,
  "role" varchar NOT NULL DEFAULT 'converter',
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "is_email_verified" bool NOT NULL DEFAULT false,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "entries" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "documents" (
  "id" uuid PRIMARY KEY,
  "entry_id" uuid NOT NULL,
  "filename" varchar NOT NULL
);

CREATE UNIQUE INDEX ON "users" ("id");

CREATE UNIQUE INDEX ON "users" ("username");

CREATE UNIQUE INDEX ON "users" ("full_name");

CREATE UNIQUE INDEX ON "entries" ("id");

CREATE UNIQUE INDEX ON "documents" ("id");

ALTER TABLE "entries" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "documents" ADD FOREIGN KEY ("entry_id") REFERENCES "entries" ("id");
