package infra

import (
	"context"
	"database/sql"

	"github.com/nomuken/william/services/server/internal/domain"
)

type SQLInterfaceRouteStore struct {
	db *sql.DB
}

type SQLPeerRouteStore struct {
	db *sql.DB
}

func NewSQLInterfaceRouteStore(db *sql.DB) *SQLInterfaceRouteStore {
	return &SQLInterfaceRouteStore{db: db}
}

func NewSQLPeerRouteStore(db *sql.DB) *SQLPeerRouteStore {
	return &SQLPeerRouteStore{db: db}
}

func (store *SQLInterfaceRouteStore) ListByInterface(ctx context.Context, interfaceID string) ([]domain.InterfaceRoute, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT interface_id, cidr, created_at
		FROM interface_allowed_routes
		WHERE interface_id = $1
		ORDER BY cidr
	`, interfaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.InterfaceRoute
	for rows.Next() {
		var route domain.InterfaceRoute
		if err := rows.Scan(&route.InterfaceID, &route.CIDR, &route.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

func (store *SQLInterfaceRouteStore) Create(ctx context.Context, interfaceID string, cidr string) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO interface_allowed_routes (interface_id, cidr)
		VALUES ($1, $2)
		ON CONFLICT (interface_id, cidr) DO NOTHING
	`, interfaceID, cidr)
	return err
}

func (store *SQLInterfaceRouteStore) Delete(ctx context.Context, interfaceID string, cidr string) error {
	_, err := store.db.ExecContext(ctx, `
		DELETE FROM interface_allowed_routes
		WHERE interface_id = $1 AND cidr = $2
	`, interfaceID, cidr)
	return err
}

func (store *SQLInterfaceRouteStore) DeleteByInterface(ctx context.Context, interfaceID string) error {
	_, err := store.db.ExecContext(ctx, `
		DELETE FROM interface_allowed_routes
		WHERE interface_id = $1
	`, interfaceID)
	return err
}

func (store *SQLPeerRouteStore) ListByPeer(ctx context.Context, peerID string) ([]domain.PeerRoute, error) {
	rows, err := store.db.QueryContext(ctx, `
		SELECT peer_id, cidr, created_at
		FROM peer_allowed_routes
		WHERE peer_id = $1
		ORDER BY cidr
	`, peerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []domain.PeerRoute
	for rows.Next() {
		var route domain.PeerRoute
		if err := rows.Scan(&route.PeerID, &route.CIDR, &route.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

func (store *SQLPeerRouteStore) Create(ctx context.Context, peerID string, cidr string) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO peer_allowed_routes (peer_id, cidr)
		VALUES ($1, $2)
		ON CONFLICT (peer_id, cidr) DO NOTHING
	`, peerID, cidr)
	return err
}

func (store *SQLPeerRouteStore) Delete(ctx context.Context, peerID string, cidr string) error {
	_, err := store.db.ExecContext(ctx, `
		DELETE FROM peer_allowed_routes
		WHERE peer_id = $1 AND cidr = $2
	`, peerID, cidr)
	return err
}

func (store *SQLPeerRouteStore) DeleteByPeer(ctx context.Context, peerID string) error {
	_, err := store.db.ExecContext(ctx, `
		DELETE FROM peer_allowed_routes
		WHERE peer_id = $1
	`, peerID)
	return err
}
