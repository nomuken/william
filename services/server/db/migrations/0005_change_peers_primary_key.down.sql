-- Drop the composite primary key
ALTER TABLE peers DROP CONSTRAINT peers_pkey;

-- Restore the original primary key on email only
ALTER TABLE peers ADD PRIMARY KEY (email);
