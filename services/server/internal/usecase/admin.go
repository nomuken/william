package usecase

import (
	"context"
	"database/sql"
	"errors"
	"net/netip"
	"sort"
	"strings"

	"github.com/nomuken/william/services/server/internal/domain"
)

type AdminUsecase interface {
	ListInterfaces(ctx context.Context) ([]domain.AdminInterface, error)
	GetInterface(ctx context.Context, interfaceID string) (domain.AdminInterface, error)
	CreateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.AdminInterface, error)
	UpdateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.AdminInterface, error)
	DeleteInterface(ctx context.Context, interfaceID string) error
	ListAllowedEmails(ctx context.Context, interfaceID string) ([]domain.AllowedEmail, error)
	CreateAllowedEmail(ctx context.Context, interfaceID string, email string) error
	DeleteAllowedEmail(ctx context.Context, interfaceID string, email string) error
	ListPeers(ctx context.Context, interfaceID string) ([]domain.PeerRecord, error)
	DeletePeer(ctx context.Context, peerID string) error
	CreateWireguardPeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (domain.WireguardPeer, error)
	DeleteWireguardPeer(ctx context.Context, peerID string) error
	UpdateWireguardPeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error
	ListInterfaceRoutes(ctx context.Context, interfaceID string) ([]domain.InterfaceRoute, error)
	CreateInterfaceRoute(ctx context.Context, interfaceID string, cidr string) error
	DeleteInterfaceRoute(ctx context.Context, interfaceID string, cidr string) error
	ListPeerRoutes(ctx context.Context, peerID string) ([]domain.PeerRoute, error)
	CreatePeerRoute(ctx context.Context, peerID string, cidr string) error
	DeletePeerRoute(ctx context.Context, peerID string, cidr string) error
	ListPeerStats(ctx context.Context) ([]domain.PeerStat, error)
	GetFirewallRules(ctx context.Context) (string, error)
	ListWireguardConfigs(ctx context.Context, interfaceID string) ([]domain.WireguardConfig, error)
}

type AdminService struct {
	repository          domain.WireguardRepository
	peerStore           domain.PeerStore
	interfaceStore      domain.InterfaceStore
	allowedEmailStore   domain.AllowedEmailStore
	interfaceRouteStore domain.InterfaceRouteStore
	peerRouteStore      domain.PeerRouteStore
}

func NewAdminService(repository domain.WireguardRepository, peerStore domain.PeerStore, interfaceStore domain.InterfaceStore, allowedEmailStore domain.AllowedEmailStore, interfaceRouteStore domain.InterfaceRouteStore, peerRouteStore domain.PeerRouteStore) *AdminService {
	return &AdminService{
		repository:          repository,
		peerStore:           peerStore,
		interfaceStore:      interfaceStore,
		allowedEmailStore:   allowedEmailStore,
		interfaceRouteStore: interfaceRouteStore,
		peerRouteStore:      peerRouteStore,
	}
}

func (service *AdminService) ListInterfaces(ctx context.Context) ([]domain.AdminInterface, error) {
	configs, err := service.interfaceStore.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]domain.AdminInterface, 0, len(configs))
	for _, config := range configs {
		iface, err := service.repository.GetInterface(ctx, config.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, domain.AdminInterface{
			ID:         iface.ID,
			Name:       config.Name,
			Address:    iface.Address,
			ListenPort: iface.ListenPort,
			PublicKey:  iface.PublicKey,
			MTU:        iface.MTU,
			Endpoint:   config.Endpoint,
		})
	}

	return items, nil
}

func (service *AdminService) GetInterface(ctx context.Context, interfaceID string) (domain.AdminInterface, error) {
	config, err := service.interfaceStore.Get(ctx, interfaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AdminInterface{}, ErrInterfaceNotFound
		}
		return domain.AdminInterface{}, err
	}

	iface, err := service.repository.GetInterface(ctx, interfaceID)
	if err != nil {
		return domain.AdminInterface{}, err
	}

	return domain.AdminInterface{
		ID:         iface.ID,
		Name:       config.Name,
		Address:    iface.Address,
		ListenPort: iface.ListenPort,
		PublicKey:  iface.PublicKey,
		MTU:        iface.MTU,
		Endpoint:   config.Endpoint,
	}, nil
}

func (service *AdminService) CreateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.AdminInterface, error) {
	if err := validateInterfaceConfig(config); err != nil {
		return domain.AdminInterface{}, err
	}

	iface, err := service.repository.CreateInterface(ctx, config)
	if err != nil {
		return domain.AdminInterface{}, err
	}

	if err := service.interfaceStore.Create(ctx, config); err != nil {
		return domain.AdminInterface{}, err
	}

	return domain.AdminInterface{
		ID:         iface.ID,
		Name:       config.Name,
		Address:    iface.Address,
		ListenPort: iface.ListenPort,
		PublicKey:  iface.PublicKey,
		MTU:        iface.MTU,
		Endpoint:   config.Endpoint,
	}, nil
}

func (service *AdminService) UpdateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.AdminInterface, error) {
	currentConfig, err := service.interfaceStore.Get(ctx, config.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AdminInterface{}, ErrInterfaceNotFound
		}
		return domain.AdminInterface{}, err
	}

	if config.Name == "" {
		config.Name = currentConfig.Name
	}
	if config.Address == "" {
		config.Address = currentConfig.Address
	}
	if config.ListenPort == 0 {
		config.ListenPort = currentConfig.ListenPort
	}
	if config.MTU == 0 {
		config.MTU = currentConfig.MTU
	}
	if config.Endpoint == "" {
		config.Endpoint = currentConfig.Endpoint
	}

	if err := validateInterfaceConfig(config); err != nil {
		return domain.AdminInterface{}, err
	}

	iface, err := service.repository.UpdateInterface(ctx, config)
	if err != nil {
		return domain.AdminInterface{}, err
	}

	if err := service.interfaceStore.Update(ctx, config); err != nil {
		return domain.AdminInterface{}, err
	}

	return domain.AdminInterface{
		ID:         iface.ID,
		Name:       config.Name,
		Address:    iface.Address,
		ListenPort: iface.ListenPort,
		PublicKey:  iface.PublicKey,
		MTU:        iface.MTU,
		Endpoint:   config.Endpoint,
	}, nil
}

func (service *AdminService) DeleteInterface(ctx context.Context, interfaceID string) error {
	if _, err := service.interfaceStore.Get(ctx, interfaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInterfaceNotFound
		}
		return err
	}

	if err := service.repository.DeleteInterface(ctx, interfaceID); err != nil {
		return err
	}

	peers, err := service.peerStore.ListByInterface(ctx, interfaceID)
	if err != nil {
		return err
	}
	for _, peer := range peers {
		if err := service.peerRouteStore.DeleteByPeer(ctx, peer.PeerID); err != nil {
			return err
		}
	}

	if err := service.interfaceRouteStore.DeleteByInterface(ctx, interfaceID); err != nil {
		return err
	}

	if err := service.peerStore.DeleteByInterface(ctx, interfaceID); err != nil {
		return err
	}

	if err := service.allowedEmailStore.DeleteByInterface(ctx, interfaceID); err != nil {
		return err
	}

	if err := service.interfaceStore.Delete(ctx, interfaceID); err != nil {
		return err
	}

	return nil
}

func (service *AdminService) ListAllowedEmails(ctx context.Context, interfaceID string) ([]domain.AllowedEmail, error) {
	if _, err := service.interfaceStore.Get(ctx, interfaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInterfaceNotFound
		}
		return nil, err
	}
	return service.allowedEmailStore.ListByInterface(ctx, interfaceID)
}

func (service *AdminService) CreateAllowedEmail(ctx context.Context, interfaceID string, email string) error {
	if interfaceID == "" || email == "" {
		return errors.New("interfaceID and email are required")
	}
	if _, err := service.interfaceStore.Get(ctx, interfaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInterfaceNotFound
		}
		return err
	}
	return service.allowedEmailStore.Create(ctx, interfaceID, email)
}

func (service *AdminService) DeleteAllowedEmail(ctx context.Context, interfaceID string, email string) error {
	if interfaceID == "" || email == "" {
		return errors.New("interfaceID and email are required")
	}
	if _, err := service.interfaceStore.Get(ctx, interfaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInterfaceNotFound
		}
		return err
	}
	return service.allowedEmailStore.Delete(ctx, interfaceID, email)
}

func (service *AdminService) ListPeers(ctx context.Context, interfaceID string) ([]domain.PeerRecord, error) {
	if interfaceID == "" {
		return service.peerStore.List(ctx)
	}
	return service.peerStore.ListByInterface(ctx, interfaceID)
}

func (service *AdminService) ListPeerStats(ctx context.Context) ([]domain.PeerStat, error) {
	return service.repository.ListPeerStats(ctx)
}

func (service *AdminService) GetFirewallRules(ctx context.Context) (string, error) {
	return service.repository.ListFirewallRules(ctx)
}

func (service *AdminService) ListWireguardConfigs(ctx context.Context, interfaceID string) ([]domain.WireguardConfig, error) {
	return service.repository.ListConfigs(ctx, interfaceID)
}

func (service *AdminService) DeletePeer(ctx context.Context, peerID string) error {
	record, err := service.peerStore.GetByPeerID(ctx, peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPeerNotFound
		}
		return err
	}

	// Remove iptables rules for this peer
	if err := service.repository.RemovePeerFirewallRules(ctx, record.AllowedIP); err != nil {
		return err
	}

	if err := service.repository.DeletePeer(ctx, record.PeerID); err != nil {
		return err
	}

	if err := service.peerRouteStore.DeleteByPeer(ctx, record.PeerID); err != nil {
		return err
	}

	if err := service.peerStore.DeleteByPeerID(ctx, record.PeerID); err != nil {
		return err
	}

	return nil
}

func (service *AdminService) CreateWireguardPeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (domain.WireguardPeer, error) {
	if interfaceID == "" {
		return domain.WireguardPeer{}, errors.New("interface id is required")
	}

	config, err := service.interfaceStore.Get(ctx, interfaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WireguardPeer{}, ErrInterfaceNotFound
		}
		return domain.WireguardPeer{}, err
	}

	if endpoint == "" {
		endpoint = config.Endpoint
	}

	interfaceRoutes, err := service.interfaceRouteStore.ListByInterface(ctx, interfaceID)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	normalizedAllowedIPs, err := normalizeCreateAllowedIPs(allowedIPs, interfaceRoutes)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	peer, err := service.repository.CreatePeer(ctx, interfaceID, endpoint, normalizedAllowedIPs)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	// Sync iptables rules for the newly created peer
	if err := service.repository.SyncPeerFirewallRules(ctx, interfaceID, peer.AllowedIP, normalizedAllowedIPs); err != nil {
		return domain.WireguardPeer{}, err
	}

	return peer, nil
}

func (service *AdminService) DeleteWireguardPeer(ctx context.Context, peerID string) error {
	if peerID == "" {
		return errors.New("peer id is required")
	}
	return service.repository.DeletePeer(ctx, peerID)
}

func (service *AdminService) UpdateWireguardPeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error {
	if interfaceID == "" || peerID == "" {
		return errors.New("interface id and peer id are required")
	}
	if len(allowedIPs) == 0 {
		return errors.New("allowed IPs are required")
	}
	if err := service.repository.UpdatePeerAllowedIPs(ctx, interfaceID, peerID, allowedIPs); err != nil {
		return err
	}

	record, err := service.peerStore.GetByPeerID(ctx, peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	updatedConfig := updatePeerConfigAllowedIPs(record.Config, allowedIPs)
	if updatedConfig == "" || updatedConfig == record.Config {
		return nil
	}
	return service.peerStore.UpdateConfig(ctx, peerID, updatedConfig)
}

func (service *AdminService) ListInterfaceRoutes(ctx context.Context, interfaceID string) ([]domain.InterfaceRoute, error) {
	if interfaceID == "" {
		return nil, errors.New("interface id is required")
	}
	return service.interfaceRouteStore.ListByInterface(ctx, interfaceID)
}

func (service *AdminService) CreateInterfaceRoute(ctx context.Context, interfaceID string, cidr string) error {
	if interfaceID == "" || cidr == "" {
		return errors.New("interface id and cidr are required")
	}
	if err := validateIPv4CIDR(cidr); err != nil {
		return err
	}
	if _, err := service.interfaceStore.Get(ctx, interfaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInterfaceNotFound
		}
		return err
	}
	if err := service.interfaceRouteStore.Create(ctx, interfaceID, cidr); err != nil {
		return err
	}
	return service.applyAllowedRoutes(ctx, interfaceID)
}

func (service *AdminService) DeleteInterfaceRoute(ctx context.Context, interfaceID string, cidr string) error {
	if interfaceID == "" || cidr == "" {
		return errors.New("interface id and cidr are required")
	}
	if err := service.interfaceRouteStore.Delete(ctx, interfaceID, cidr); err != nil {
		return err
	}
	return service.applyAllowedRoutes(ctx, interfaceID)
}

func (service *AdminService) ListPeerRoutes(ctx context.Context, peerID string) ([]domain.PeerRoute, error) {
	if peerID == "" {
		return nil, errors.New("peer id is required")
	}
	return service.peerRouteStore.ListByPeer(ctx, peerID)
}

func (service *AdminService) CreatePeerRoute(ctx context.Context, peerID string, cidr string) error {
	if peerID == "" || cidr == "" {
		return errors.New("peer id and cidr are required")
	}
	if err := validateIPv4CIDR(cidr); err != nil {
		return err
	}
	record, err := service.peerStore.GetByPeerID(ctx, peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPeerNotFound
		}
		return err
	}
	if err := service.peerRouteStore.Create(ctx, peerID, cidr); err != nil {
		return err
	}
	return service.applyAllowedRoutes(ctx, record.InterfaceID)
}

func (service *AdminService) DeletePeerRoute(ctx context.Context, peerID string, cidr string) error {
	if peerID == "" || cidr == "" {
		return errors.New("peer id and cidr are required")
	}
	record, err := service.peerStore.GetByPeerID(ctx, peerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPeerNotFound
		}
		return err
	}
	if err := service.peerRouteStore.Delete(ctx, peerID, cidr); err != nil {
		return err
	}
	return service.applyAllowedRoutes(ctx, record.InterfaceID)
}

func validateInterfaceConfig(config domain.InterfaceConfig) error {
	if config.ID == "" {
		return errors.New("interface id is required")
	}
	if config.Name == "" {
		return errors.New("name is required")
	}
	if config.Address == "" {
		return errors.New("address is required")
	}
	if config.ListenPort == 0 {
		return errors.New("listen port is required")
	}
	if config.MTU == 0 {
		return errors.New("mtu is required")
	}
	if config.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	return nil
}

func (service *AdminService) applyAllowedRoutes(ctx context.Context, interfaceID string) error {
	interfaceRoutes, err := service.interfaceRouteStore.ListByInterface(ctx, interfaceID)
	if err != nil {
		return err
	}

	peers, err := service.peerStore.ListByInterface(ctx, interfaceID)
	if err != nil {
		return err
	}

	for _, peer := range peers {
		peerRoutes, err := service.peerRouteStore.ListByPeer(ctx, peer.PeerID)
		if err != nil {
			return err
		}
		allowedIPs := buildAllowedIPs(peer.AllowedIP, interfaceRoutes, peerRoutes)
		if err := service.repository.UpdatePeerAllowedIPs(ctx, interfaceID, peer.PeerID, allowedIPs); err != nil {
			return err
		}

		// Sync iptables rules for this peer
		if err := service.repository.SyncPeerFirewallRules(ctx, interfaceID, peer.AllowedIP, allowedIPs); err != nil {
			return err
		}

		updatedConfig := updatePeerConfigAllowedIPs(peer.Config, allowedIPs)
		if updatedConfig != "" && updatedConfig != peer.Config {
			if err := service.peerStore.UpdateConfig(ctx, peer.PeerID, updatedConfig); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildAllowedIPs(peerAllowedIP string, interfaceRoutes []domain.InterfaceRoute, peerRoutes []domain.PeerRoute) []string {
	items := []string{}
	if peerAllowedIP != "" {
		items = append(items, peerAllowedIP)
	}

	interfaceCIDRs := extractRouteCIDRs(interfaceRoutes)
	peerCIDRs := extractPeerRouteCIDRs(peerRoutes)

	sort.Strings(interfaceCIDRs)
	sort.Strings(peerCIDRs)

	items = append(items, interfaceCIDRs...)
	items = append(items, peerCIDRs...)

	return dedupeStrings(items)
}

func updatePeerConfigAllowedIPs(config string, allowedIPs []string) string {
	if config == "" || len(allowedIPs) == 0 {
		return config
	}

	lines := strings.Split(config, "\n")
	updated := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "AllowedIPs =") {
			continue
		}
		prefixIndex := strings.Index(line, "AllowedIPs =")
		if prefixIndex == -1 {
			continue
		}
		lines[i] = line[:prefixIndex] + "AllowedIPs = " + strings.Join(allowedIPs, ", ")
		updated = true
	}
	if !updated {
		return config
	}
	return strings.Join(lines, "\n")
}

func normalizeCreateAllowedIPs(allowedIPs []string, interfaceRoutes []domain.InterfaceRoute) ([]string, error) {
	items := make([]string, 0, len(allowedIPs)+len(interfaceRoutes))
	for _, cidr := range allowedIPs {
		if cidr == "" {
			continue
		}
		if err := validateIPv4CIDR(cidr); err != nil {
			return nil, err
		}
		items = append(items, cidr)
	}

	items = append(items, extractRouteCIDRs(interfaceRoutes)...)
	sort.Strings(items)
	return dedupeStrings(items), nil
}

func extractRouteCIDRs(routes []domain.InterfaceRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, route := range routes {
		cidrs = append(cidrs, route.CIDR)
	}
	return cidrs
}

func extractPeerRouteCIDRs(routes []domain.PeerRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, route := range routes {
		cidrs = append(cidrs, route.CIDR)
	}
	return cidrs
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func validateIPv4CIDR(cidr string) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	if !prefix.Addr().Is4() {
		return errors.New("only IPv4 CIDR is supported")
	}
	return nil
}
