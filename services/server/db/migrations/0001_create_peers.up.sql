CREATE TABLE peers (
  email TEXT PRIMARY KEY,
  peer_id TEXT NOT NULL,
  interface_id TEXT NOT NULL,
  allowed_ip TEXT NOT NULL,
  config TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX peers_peer_id_idx ON peers(peer_id);
