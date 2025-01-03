ALTER TABLE documents DROP CONSTRAINT documents_entry_id_fkey;

ALTER TABLE documents
ADD CONSTRAINT documents_entry_id_fkey 
FOREIGN KEY (entry_id) REFERENCES entries (id) ON DELETE CASCADE;
