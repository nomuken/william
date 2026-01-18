package infra

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/nomuken/william/services/server/internal/domain"
)

// MockWireguardRepository simulates wireguard operations for dev environments.
type MockWireguardRepository struct {
	interfaceStore domain.InterfaceStore
	peerStore      domain.PeerStore
}

func NewMockWireguardRepository(interfaceStore domain.InterfaceStore, peerStore domain.PeerStore) *MockWireguardRepository {
	return &MockWireguardRepository{
		interfaceStore: interfaceStore,
		peerStore:      peerStore,
	}
}

func (repo *MockWireguardRepository) ListInterfaces(ctx context.Context) ([]domain.WireguardInterface, error) {
	configs, err := repo.interfaceStore.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.WireguardInterface, 0, len(configs))
	for _, config := range configs {
		items = append(items, interfaceFromConfig(config))
	}
	return items, nil
}

func (repo *MockWireguardRepository) GetInterface(ctx context.Context, interfaceID string) (domain.WireguardInterface, error) {
	config, err := repo.interfaceStore.Get(ctx, interfaceID)
	if err != nil {
		return domain.WireguardInterface{}, err
	}
	return interfaceFromConfig(config), nil
}

func (repo *MockWireguardRepository) CreateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	return interfaceFromConfig(config), nil
}

func (repo *MockWireguardRepository) UpdateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	return interfaceFromConfig(config), nil
}

func (repo *MockWireguardRepository) DeleteInterface(ctx context.Context, interfaceID string) error {
	return nil
}

func (repo *MockWireguardRepository) CreatePeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (domain.WireguardPeer, error) {
	config, err := repo.interfaceStore.Get(ctx, interfaceID)
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	if endpoint == "" {
		endpoint = config.Endpoint
	}
	if endpoint == "" {
		return domain.WireguardPeer{}, errors.New("endpoint is required")
	}

	allowedIP, err := repo.nextAllowedIP(ctx, config)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	privateKey, err := randomKey()
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	publicKey, err := randomKey()
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	allowedIPs = normalizeAllowedIPs(allowedIP, allowedIPs)
	configText := fmt.Sprintf("[Interface]\nPrivateKey = %s\nAddress = %s\n\n[Peer]\nPublicKey = %s\nAllowedIPs = %s\nEndpoint = %s\n", privateKey, allowedIP, publicKey, strings.Join(allowedIPs, ", "), endpoint)

	return domain.WireguardPeer{
		ID:          publicKey,
		InterfaceID: interfaceID,
		AllowedIP:   allowedIP,
		Config:      configText,
	}, nil
}

func (repo *MockWireguardRepository) UpdatePeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error {
	return nil
}

func (repo *MockWireguardRepository) DeletePeer(ctx context.Context, peerID string) error {
	return nil
}

func (repo *MockWireguardRepository) ListPeerStats(ctx context.Context) ([]domain.PeerStat, error) {
	peers, err := repo.peerStore.List(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]domain.PeerStat, 0, len(peers))
	for _, peer := range peers {
		stats = append(stats, domain.PeerStat{
			PeerID:      peer.PeerID,
			InterfaceID: peer.InterfaceID,
		})
	}

	return stats, nil
}

func (repo *MockWireguardRepository) ListFirewallRules(ctx context.Context) (string, error) {
	return "", nil
}

func (repo *MockWireguardRepository) EnsureFirewallChain(ctx context.Context) error {
	return nil
}

func (repo *MockWireguardRepository) SyncPeerFirewallRules(ctx context.Context, interfaceID string, peerAllowedIP string, allowedIPs []string) error {
	return nil
}

func (repo *MockWireguardRepository) RemovePeerFirewallRules(ctx context.Context, peerAllowedIP string) error {
	return nil
}

func (repo *MockWireguardRepository) ListConfigs(ctx context.Context, interfaceID string) ([]domain.WireguardConfig, error) {
	configs, err := repo.interfaceStore.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.WireguardConfig, 0, len(configs))
	for _, config := range configs {
		if interfaceID != "" && config.ID != interfaceID {
			continue
		}
		items = append(items, domain.WireguardConfig{
			InterfaceID: config.ID,
			Config:      "",
		})
	}

	return items, nil
}

func (repo *MockWireguardRepository) nextAllowedIP(ctx context.Context, config domain.InterfaceConfig) (string, error) {
	prefix, err := netip.ParsePrefix(config.Address)
	if err != nil {
		return "", fmt.Errorf("parse interface address: %w", err)
	}
	interfaceAddr := prefix.Addr()

	used := make(map[netip.Addr]struct{})
	used[interfaceAddr] = struct{}{}
	used[prefix.Masked().Addr()] = struct{}{}

	peers, err := repo.peerStore.ListByInterface(ctx, config.ID)
	if err != nil {
		return "", err
	}
	for _, peer := range peers {
		peerPrefix, err := netip.ParsePrefix(peer.AllowedIP)
		if err != nil {
			continue
		}
		if peerPrefix.Addr().Is4() {
			used[peerPrefix.Addr()] = struct{}{}
		}
	}

	candidate, err := nextAvailableIPv4Mock(prefix, interfaceAddr, used)
	if err != nil {
		return "", err
	}

	return candidate.String() + "/32", nil
}

func interfaceFromConfig(config domain.InterfaceConfig) domain.WireguardInterface {
	return domain.WireguardInterface{
		ID:         config.ID,
		Name:       config.Name,
		Address:    config.Address,
		ListenPort: config.ListenPort,
		PublicKey:  deterministicKey(config.ID),
		MTU:        config.MTU,
	}
}

func deterministicKey(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func randomKey() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer), nil
}

func nextAvailableIPv4Mock(prefix netip.Prefix, interfaceAddr netip.Addr, used map[netip.Addr]struct{}) (netip.Addr, error) {
	if !prefix.Addr().Is4() {
		return netip.Addr{}, errors.New("interface address is not IPv4")
	}
	if prefix.Bits() >= 31 {
		return netip.Addr{}, errors.New("prefix too small to allocate address")
	}

	networkAddr := prefix.Masked().Addr()
	for candidate := networkAddr.Next(); prefix.Contains(candidate); candidate = candidate.Next() {
		if candidate == interfaceAddr {
			continue
		}
		if _, exists := used[candidate]; exists {
			continue
		}
		return candidate, nil
	}

	return netip.Addr{}, errors.New("no available address in prefix")
}
