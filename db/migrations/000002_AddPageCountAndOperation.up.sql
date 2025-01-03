ALTER TABLE entries ADD COLUMN status varchar NOT NULL DEFAULT 'processing';
ALTER TABLE entries ADD COLUMN operation varchar NOT NULL DEFAULT 'merge';
ALTER TABLE documents ADD COLUMN page_count integer NOT NULL DEFAULT 0;