-- Drop the existing primary key constraint
ALTER TABLE peers DROP CONSTRAINT peers_pkey;

-- Add a composite primary key on (email, interface_id)
ALTER TABLE peers ADD PRIMARY KEY (email, interface_id);
