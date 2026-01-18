package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nomuken/william/services/server/internal/domain"
)

type WireguardUsecase interface {
	ListInterfaces(ctx context.Context, email string) ([]domain.WireguardInterface, error)
	CreatePeer(ctx context.Context, email string, interfaceID string) (domain.WireguardPeer, error)
	GetPeerByEmail(ctx context.Context, email string) (domain.PeerRecord, error)
	GetPeerByEmailAndInterface(ctx context.Context, email string, interfaceID string) (domain.PeerRecord, error)
	DeletePeer(ctx context.Context, email string, peerID string) error
	ListPeerStatuses(ctx context.Context, email string) ([]domain.PeerStatus, error)
}

type WireguardService struct {
	repository          domain.WireguardRepository
	store               domain.PeerStore
	interfaceStore      domain.InterfaceStore
	allowedEmailStore   domain.AllowedEmailStore
	interfaceRouteStore domain.InterfaceRouteStore
}

var ErrPeerAlreadyExists = errors.New("peer already exists")
var ErrPeerNotFound = errors.New("peer not found")
var ErrPeerForbidden = errors.New("peer access forbidden")
var ErrEmailNotAllowed = errors.New("email is not allowed")
var ErrInterfaceNotFound = errors.New("interface not found")

func NewWireguardService(repository domain.WireguardRepository, store domain.PeerStore, interfaceStore domain.InterfaceStore, allowedEmailStore domain.AllowedEmailStore, interfaceRouteStore domain.InterfaceRouteStore) *WireguardService {
	return &WireguardService{
		repository:          repository,
		store:               store,
		interfaceStore:      interfaceStore,
		allowedEmailStore:   allowedEmailStore,
		interfaceRouteStore: interfaceRouteStore,
	}
}

func (service *WireguardService) ListInterfaces(ctx context.Context, email string) ([]domain.WireguardInterface, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	allowedInterfaces, err := service.allowedEmailStore.ListInterfaceIDsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	allInterfaces, err := service.repository.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	storedConfigs, err := service.interfaceStore.List(ctx)
	if err != nil {
		return nil, err
	}
	configByID := make(map[string]domain.InterfaceConfig, len(storedConfigs))
	for _, config := range storedConfigs {
		configByID[config.ID] = config
	}

	allowed := make(map[string]struct{}, len(allowedInterfaces))
	for _, id := range allowedInterfaces {
		allowed[id] = struct{}{}
	}

	interfaces := make([]domain.WireguardInterface, 0, len(allowedInterfaces))
	for _, item := range allInterfaces {
		if _, ok := allowed[item.ID]; ok {
			if config, ok := configByID[item.ID]; ok && config.Name != "" {
				item.Name = config.Name
			}
			interfaces = append(interfaces, item)
		}
	}

	return interfaces, nil
}

func (service *WireguardService) ListPeerStatuses(ctx context.Context, email string) ([]domain.PeerStatus, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	peers, err := service.store.ListByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if len(peers) == 0 {
		return []domain.PeerStatus{}, nil
	}

	allowedPeerIDs := make(map[string]domain.PeerRecord, len(peers))
	for _, peer := range peers {
		allowedPeerIDs[peer.PeerID] = peer
	}

	stats, err := service.repository.ListPeerStats(ctx)
	if err != nil {
		return nil, err
	}

	configs, err := service.interfaceStore.List(ctx)
	if err != nil {
		return nil, err
	}
	nameByID := make(map[string]string, len(configs))
	for _, config := range configs {
		nameByID[config.ID] = config.Name
	}

	items := make([]domain.PeerStatus, 0, len(stats))
	for _, stat := range stats {
		record, ok := allowedPeerIDs[stat.PeerID]
		if !ok {
			continue
		}
		items = append(items, domain.PeerStatus{
			PeerID:          stat.PeerID,
			InterfaceID:     record.InterfaceID,
			InterfaceName:   nameByID[record.InterfaceID],
			RxBytes:         stat.RxBytes,
			TxBytes:         stat.TxBytes,
			LastHandshakeAt: stat.LastHandshakeAt,
		})
	}

	return items, nil
}

func (service *WireguardService) CreatePeer(ctx context.Context, email string, interfaceID string) (domain.WireguardPeer, error) {
	if email == "" {
		return domain.WireguardPeer{}, errors.New("email is required")
	}

	allowed, err := service.allowedEmailStore.Exists(ctx, interfaceID, email)
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	if !allowed {
		return domain.WireguardPeer{}, ErrEmailNotAllowed
	}

	if _, err := service.store.GetByEmailAndInterface(ctx, email, interfaceID); err == nil {
		return domain.WireguardPeer{}, ErrPeerAlreadyExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.WireguardPeer{}, err
	}

	interfaceConfig, err := service.interfaceStore.Get(ctx, interfaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WireguardPeer{}, ErrInterfaceNotFound
		}
		return domain.WireguardPeer{}, err
	}
	if interfaceConfig.Endpoint == "" {
		return domain.WireguardPeer{}, errors.New("endpoint is required")
	}

	allowedIPs, err := service.interfaceRouteStore.ListByInterface(ctx, interfaceID)
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	peerAllowedIPs := extractInterfaceRouteCIDRs(allowedIPs)

	peer, err := service.repository.CreatePeer(ctx, interfaceID, interfaceConfig.Endpoint, peerAllowedIPs)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	// Sync iptables rules for the newly created peer
	if err := service.repository.SyncPeerFirewallRules(ctx, interfaceID, peer.AllowedIP, peerAllowedIPs); err != nil {
		return domain.WireguardPeer{}, err
	}

	record := domain.PeerRecord{
		Email:       email,
		PeerID:      peer.ID,
		InterfaceID: peer.InterfaceID,
		AllowedIP:   peer.AllowedIP,
		Config:      peer.Config,
	}
	if err := service.store.Create(ctx, record); err != nil {
		return domain.WireguardPeer{}, err
	}

	return peer, nil
}

func (service *WireguardService) GetPeerByEmail(ctx context.Context, email string) (domain.PeerRecord, error) {
	if email == "" {
		return domain.PeerRecord{}, errors.New("email is required")
	}

	record, err := service.store.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PeerRecord{}, ErrPeerNotFound
		}
		return domain.PeerRecord{}, err
	}

	allowed, err := service.allowedEmailStore.Exists(ctx, record.InterfaceID, email)
	if err != nil {
		return domain.PeerRecord{}, err
	}
	if !allowed {
		return domain.PeerRecord{}, ErrPeerForbidden
	}

	return record, nil
}

func (service *WireguardService) GetPeerByEmailAndInterface(ctx context.Context, email string, interfaceID string) (domain.PeerRecord, error) {
	if email == "" {
		return domain.PeerRecord{}, errors.New("email is required")
	}
	if interfaceID == "" {
		return domain.PeerRecord{}, errors.New("interface id is required")
	}

	record, err := service.store.GetByEmailAndInterface(ctx, email, interfaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PeerRecord{}, ErrPeerNotFound
		}
		return domain.PeerRecord{}, err
	}

	allowed, err := service.allowedEmailStore.Exists(ctx, record.InterfaceID, email)
	if err != nil {
		return domain.PeerRecord{}, err
	}
	if !allowed {
		return domain.PeerRecord{}, ErrPeerForbidden
	}

	return record, nil
}

func (service *WireguardService) DeletePeer(ctx context.Context, email string, peerID string) error {
	if peerID == "" {
		return errors.New("peer id is required")
	}
	if email == "" {
		return errors.New("email is required")
	}

	record, err := service.store.GetByPeerID(ctx, peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPeerNotFound
		}
		return err
	}
	if record.Email != email {
		return ErrPeerForbidden
	}

	// Remove iptables rules for this peer
	if err := service.repository.RemovePeerFirewallRules(ctx, record.AllowedIP); err != nil {
		return err
	}

	if err := service.repository.DeletePeer(ctx, record.PeerID); err != nil {
		return err
	}

	if err := service.store.DeleteByPeerID(ctx, record.PeerID); err != nil {
		return err
	}

	return nil
}

func extractInterfaceRouteCIDRs(routes []domain.InterfaceRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, route := range routes {
		cidrs = append(cidrs, route.CIDR)
	}
	return cidrs
}
