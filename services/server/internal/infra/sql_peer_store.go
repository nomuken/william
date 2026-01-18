package infra

import (
	"context"
	"database/sql"

	"github.com/nomuken/william/services/server/internal/db"
	"github.com/nomuken/william/services/server/internal/domain"
)

// SQLPeerStore provides peer persistence backed by SQL.
type SQLPeerStore struct {
	queries *db.Queries
}

// NewSQLPeerStore wires a peer store using sqlc queries.
func NewSQLPeerStore(database *sql.DB) *SQLPeerStore {
	return &SQLPeerStore{queries: db.New(database)}
}

func (store *SQLPeerStore) GetByEmail(ctx context.Context, email string) (domain.PeerRecord, error) {
	peer, err := store.queries.GetPeerByEmail(ctx, email)
	if err != nil {
		return domain.PeerRecord{}, err
	}

	return domain.PeerRecord{
		Email:       peer.Email,
		PeerID:      peer.PeerID,
		InterfaceID: peer.InterfaceID,
		AllowedIP:   peer.AllowedIp,
		Config:      peer.Config,
		CreatedAt:   peer.CreatedAt,
	}, nil
}

func (store *SQLPeerStore) ListByEmail(ctx context.Context, email string) ([]domain.PeerRecord, error) {
	peers, err := store.queries.ListPeersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PeerRecord, 0, len(peers))
	for _, peer := range peers {
		items = append(items, domain.PeerRecord{
			Email:       peer.Email,
			PeerID:      peer.PeerID,
			InterfaceID: peer.InterfaceID,
			AllowedIP:   peer.AllowedIp,
			Config:      peer.Config,
			CreatedAt:   peer.CreatedAt,
		})
	}

	return items, nil
}

func (store *SQLPeerStore) GetByPeerID(ctx context.Context, peerID string) (domain.PeerRecord, error) {
	peer, err := store.queries.GetPeerByID(ctx, peerID)
	if err != nil {
		return domain.PeerRecord{}, err
	}

	return domain.PeerRecord{
		Email:       peer.Email,
		PeerID:      peer.PeerID,
		InterfaceID: peer.InterfaceID,
		AllowedIP:   peer.AllowedIp,
		Config:      peer.Config,
		CreatedAt:   peer.CreatedAt,
	}, nil
}

func (store *SQLPeerStore) GetByEmailAndInterface(ctx context.Context, email string, interfaceID string) (domain.PeerRecord, error) {
	peer, err := store.queries.GetPeerByEmailAndInterface(ctx, db.GetPeerByEmailAndInterfaceParams{
		Email:       email,
		InterfaceID: interfaceID,
	})
	if err != nil {
		return domain.PeerRecord{}, err
	}

	return domain.PeerRecord{
		Email:       peer.Email,
		PeerID:      peer.PeerID,
		InterfaceID: peer.InterfaceID,
		AllowedIP:   peer.AllowedIp,
		Config:      peer.Config,
		CreatedAt:   peer.CreatedAt,
	}, nil
}

func (store *SQLPeerStore) List(ctx context.Context) ([]domain.PeerRecord, error) {
	peers, err := store.queries.ListPeers(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PeerRecord, 0, len(peers))
	for _, peer := range peers {
		items = append(items, domain.PeerRecord{
			Email:       peer.Email,
			PeerID:      peer.PeerID,
			InterfaceID: peer.InterfaceID,
			AllowedIP:   peer.AllowedIp,
			Config:      peer.Config,
			CreatedAt:   peer.CreatedAt,
		})
	}

	return items, nil
}

func (store *SQLPeerStore) ListByInterface(ctx context.Context, interfaceID string) ([]domain.PeerRecord, error) {
	peers, err := store.queries.ListPeersByInterface(ctx, interfaceID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.PeerRecord, 0, len(peers))
	for _, peer := range peers {
		items = append(items, domain.PeerRecord{
			Email:       peer.Email,
			PeerID:      peer.PeerID,
			InterfaceID: peer.InterfaceID,
			AllowedIP:   peer.AllowedIp,
			Config:      peer.Config,
			CreatedAt:   peer.CreatedAt,
		})
	}

	return items, nil
}

func (store *SQLPeerStore) Create(ctx context.Context, record domain.PeerRecord) error {
	params := db.CreatePeerParams{
		Email:       record.Email,
		PeerID:      record.PeerID,
		InterfaceID: record.InterfaceID,
		AllowedIp:   record.AllowedIP,
		Config:      record.Config,
	}

	return store.queries.CreatePeer(ctx, params)
}

func (store *SQLPeerStore) UpdateConfig(ctx context.Context, peerID string, config string) error {
	params := db.UpdatePeerConfigParams{
		Config: config,
		PeerID: peerID,
	}

	return store.queries.UpdatePeerConfig(ctx, params)
}

func (store *SQLPeerStore) DeleteByPeerID(ctx context.Context, peerID string) error {
	if _, err := store.queries.GetPeerByID(ctx, peerID); err != nil {
		return err
	}

	return store.queries.DeletePeerByID(ctx, peerID)
}

func (store *SQLPeerStore) DeleteByInterface(ctx context.Context, interfaceID string) error {
	return store.queries.DeletePeersByInterface(ctx, interfaceID)
}
