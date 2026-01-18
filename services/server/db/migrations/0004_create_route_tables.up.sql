CREATE TABLE interface_allowed_routes (
  interface_id TEXT NOT NULL REFERENCES interfaces(id) ON DELETE CASCADE,
  cidr TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (interface_id, cidr)
);

CREATE UNIQUE INDEX IF NOT EXISTS peers_peer_id_unique ON peers(peer_id);

CREATE TABLE peer_allowed_routes (
  peer_id TEXT NOT NULL REFERENCES peers(peer_id) ON DELETE CASCADE,
  cidr TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (peer_id, cidr)
);
