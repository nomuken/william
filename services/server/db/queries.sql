-- name: GetPeerByEmail :one
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
WHERE email = $1
LIMIT 1;

-- name: GetPeerByID :one
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
WHERE peer_id = $1
LIMIT 1;

-- name: GetPeerByEmailAndInterface :one
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
WHERE email = $1 AND interface_id = $2
LIMIT 1;

-- name: CreatePeer :exec
INSERT INTO peers (email, peer_id, interface_id, allowed_ip, config)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePeerConfig :exec
UPDATE peers
SET config = $1
WHERE peer_id = $2;

-- name: DeletePeerByID :exec
DELETE FROM peers
WHERE peer_id = $1;

-- name: DeletePeersByInterface :exec
DELETE FROM peers
WHERE interface_id = $1;

-- name: ListPeers :many
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
ORDER BY created_at DESC;

-- name: ListPeersByEmail :many
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
WHERE email = $1
ORDER BY created_at DESC;

-- name: ListPeersByInterface :many
SELECT email, peer_id, interface_id, allowed_ip, config, created_at
FROM peers
WHERE interface_id = $1
ORDER BY created_at DESC;

-- name: CreateInterface :exec
INSERT INTO interfaces (id, name, address, listen_port, mtu, endpoint)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateInterface :exec
UPDATE interfaces
SET name = $1, address = $2, listen_port = $3, mtu = $4, endpoint = $5
WHERE id = $6;

-- name: DeleteInterface :exec
DELETE FROM interfaces
WHERE id = $1;

-- name: GetInterface :one
SELECT id, name, address, listen_port, mtu, endpoint, created_at
FROM interfaces
WHERE id = $1
LIMIT 1;

-- name: ListInterfaces :many
SELECT id, name, address, listen_port, mtu, endpoint, created_at
FROM interfaces
ORDER BY id;

-- name: CreateAllowedEmail :exec
INSERT INTO allowed_emails (interface_id, email)
VALUES ($1, $2);

-- name: DeleteAllowedEmail :exec
DELETE FROM allowed_emails
WHERE interface_id = $1 AND email = $2;

-- name: DeleteAllowedEmailsByInterface :exec
DELETE FROM allowed_emails
WHERE interface_id = $1;

-- name: ListAllowedEmails :many
SELECT interface_id, email, created_at
FROM allowed_emails
WHERE interface_id = $1
ORDER BY email;

-- name: ListAllowedInterfacesByEmail :many
SELECT interface_id
FROM allowed_emails
WHERE email = $1
ORDER BY interface_id;

-- name: AllowedEmailExists :one
SELECT COUNT(1)
FROM allowed_emails
WHERE interface_id = $1 AND email = $2;
