package infra

import (
	"context"
	"database/sql"

	"github.com/nomuken/william/services/server/internal/db"
	"github.com/nomuken/william/services/server/internal/domain"
)

// SQLInterfaceStore persists wireguard interface metadata.
type SQLInterfaceStore struct {
	queries *db.Queries
}

// SQLAllowedEmailStore tracks allowed emails per interface.
type SQLAllowedEmailStore struct {
	queries *db.Queries
}

// NewSQLInterfaceStore wires interface storage using sqlc queries.
func NewSQLInterfaceStore(database *sql.DB) *SQLInterfaceStore {
	return &SQLInterfaceStore{queries: db.New(database)}
}

// NewSQLAllowedEmailStore wires allowed email storage using sqlc queries.
func NewSQLAllowedEmailStore(database *sql.DB) *SQLAllowedEmailStore {
	return &SQLAllowedEmailStore{queries: db.New(database)}
}

func (store *SQLInterfaceStore) Get(ctx context.Context, id string) (domain.InterfaceConfig, error) {
	row, err := store.queries.GetInterface(ctx, id)
	if err != nil {
		return domain.InterfaceConfig{}, err
	}

	return domain.InterfaceConfig{
		ID:         row.ID,
		Name:       row.Name,
		Address:    row.Address,
		ListenPort: uint32(row.ListenPort),
		MTU:        uint32(row.Mtu),
		Endpoint:   row.Endpoint,
	}, nil
}

func (store *SQLInterfaceStore) List(ctx context.Context) ([]domain.InterfaceConfig, error) {
	rows, err := store.queries.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.InterfaceConfig, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.InterfaceConfig{
			ID:         row.ID,
			Name:       row.Name,
			Address:    row.Address,
			ListenPort: uint32(row.ListenPort),
			MTU:        uint32(row.Mtu),
			Endpoint:   row.Endpoint,
		})
	}

	return items, nil
}

func (store *SQLInterfaceStore) Create(ctx context.Context, config domain.InterfaceConfig) error {
	params := db.CreateInterfaceParams{
		ID:         config.ID,
		Name:       config.Name,
		Address:    config.Address,
		ListenPort: int64(config.ListenPort),
		Mtu:        int64(config.MTU),
		Endpoint:   config.Endpoint,
	}

	return store.queries.CreateInterface(ctx, params)
}

func (store *SQLInterfaceStore) Update(ctx context.Context, config domain.InterfaceConfig) error {
	params := db.UpdateInterfaceParams{
		Name:       config.Name,
		Address:    config.Address,
		ListenPort: int64(config.ListenPort),
		Mtu:        int64(config.MTU),
		Endpoint:   config.Endpoint,
		ID:         config.ID,
	}

	return store.queries.UpdateInterface(ctx, params)
}

func (store *SQLInterfaceStore) Delete(ctx context.Context, id string) error {
	return store.queries.DeleteInterface(ctx, id)
}

func (store *SQLAllowedEmailStore) ListByInterface(ctx context.Context, interfaceID string) ([]domain.AllowedEmail, error) {
	emails, err := store.queries.ListAllowedEmails(ctx, interfaceID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.AllowedEmail, 0, len(emails))
	for _, email := range emails {
		items = append(items, domain.AllowedEmail{
			InterfaceID: email.InterfaceID,
			Email:       email.Email,
			CreatedAt:   email.CreatedAt,
		})
	}

	return items, nil
}

func (store *SQLAllowedEmailStore) ListInterfaceIDsByEmail(ctx context.Context, email string) ([]string, error) {
	return store.queries.ListAllowedInterfacesByEmail(ctx, email)
}

func (store *SQLAllowedEmailStore) Exists(ctx context.Context, interfaceID string, email string) (bool, error) {
	count, err := store.queries.AllowedEmailExists(ctx, db.AllowedEmailExistsParams{
		InterfaceID: interfaceID,
		Email:       email,
	})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (store *SQLAllowedEmailStore) Create(ctx context.Context, interfaceID string, email string) error {
	params := db.CreateAllowedEmailParams{
		InterfaceID: interfaceID,
		Email:       email,
	}
	return store.queries.CreateAllowedEmail(ctx, params)
}

func (store *SQLAllowedEmailStore) Delete(ctx context.Context, interfaceID string, email string) error {
	params := db.DeleteAllowedEmailParams{
		InterfaceID: interfaceID,
		Email:       email,
	}
	return store.queries.DeleteAllowedEmail(ctx, params)
}

func (store *SQLAllowedEmailStore) DeleteByInterface(ctx context.Context, interfaceID string) error {
	return store.queries.DeleteAllowedEmailsByInterface(ctx, interfaceID)
}
