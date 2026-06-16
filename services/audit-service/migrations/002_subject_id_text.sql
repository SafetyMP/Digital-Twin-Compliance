-- Subject identifiers may be twin IDs, alert UUIDs, or other opaque strings.
ALTER TABLE audit_entry_index
  ALTER COLUMN subject_id TYPE TEXT USING subject_id::text;
