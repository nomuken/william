CREATE TABLE interfaces (
  id TEXT PRIMARY KEY,
  address TEXT NOT NULL,
  listen_port INTEGER NOT NULL,
  mtu INTEGER NOT NULL,
  endpoint TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE allowed_emails (
  interface_id TEXT NOT NULL,
  email TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (interface_id, email)
);

CREATE INDEX allowed_emails_email_idx ON allowed_emails(email);
