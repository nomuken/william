package infra

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	adminv1 "github.com/nomuken/william/services/server/gen/proto/admin/v1"
	"github.com/nomuken/william/services/server/gen/proto/admin/v1/adminv1connect"
	"github.com/nomuken/william/services/server/internal/domain"
	"google.golang.org/protobuf/types/known/emptypb"
)

const adminServiceEndpoint = "http://admin-server:8081"

// AdminRPCWireguardRepository delegates wireguard operations to admin-server.
type AdminRPCWireguardRepository struct {
	client adminv1connect.WilliamAdminServiceClient
}

func NewAdminRPCWireguardRepository(httpClient *http.Client) *AdminRPCWireguardRepository {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := adminv1connect.NewWilliamAdminServiceClient(httpClient, adminServiceEndpoint)
	return &AdminRPCWireguardRepository{client: client}
}

func (repo *AdminRPCWireguardRepository) ListInterfaces(ctx context.Context) ([]domain.WireguardInterface, error) {
	response, err := repo.client.ListInterfaces(ctx, connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return nil, err
	}

	interfaces := make([]domain.WireguardInterface, 0, len(response.Msg.Interfaces))
	for _, iface := range response.Msg.Interfaces {
		interfaces = append(interfaces, domain.WireguardInterface{
			ID:         iface.GetId(),
			Name:       iface.GetName(),
			Address:    iface.GetAddress(),
			ListenPort: iface.GetListenPort(),
			PublicKey:  iface.GetPublicKey(),
			MTU:        iface.GetMtu(),
		})
	}

	return interfaces, nil
}

func (repo *AdminRPCWireguardRepository) GetInterface(ctx context.Context, interfaceID string) (domain.WireguardInterface, error) {
	response, err := repo.client.GetInterface(ctx, connect.NewRequest(&adminv1.GetAdminInterfaceRequest{Id: interfaceID}))
	if err != nil {
		return domain.WireguardInterface{}, err
	}
	iface := response.Msg.GetInterface()
	if iface == nil {
		return domain.WireguardInterface{}, errors.New("interface not found")
	}

	return domain.WireguardInterface{
		ID:         iface.GetId(),
		Name:       iface.GetName(),
		Address:    iface.GetAddress(),
		ListenPort: iface.GetListenPort(),
		PublicKey:  iface.GetPublicKey(),
		MTU:        iface.GetMtu(),
	}, nil
}

func (repo *AdminRPCWireguardRepository) CreateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	response, err := repo.client.CreateInterface(ctx, connect.NewRequest(&adminv1.CreateAdminInterfaceRequest{
		Name:       config.Name,
		Address:    config.Address,
		ListenPort: config.ListenPort,
		Mtu:        config.MTU,
		Endpoint:   config.Endpoint,
	}))
	if err != nil {
		return domain.WireguardInterface{}, err
	}

	iface := response.Msg.GetInterface()
	if iface == nil {
		return domain.WireguardInterface{}, errors.New("interface not created")
	}

	return domain.WireguardInterface{
		ID:         iface.GetId(),
		Name:       iface.GetName(),
		Address:    iface.GetAddress(),
		ListenPort: iface.GetListenPort(),
		PublicKey:  iface.GetPublicKey(),
		MTU:        iface.GetMtu(),
	}, nil
}

func (repo *AdminRPCWireguardRepository) UpdateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	response, err := repo.client.UpdateInterface(ctx, connect.NewRequest(&adminv1.UpdateAdminInterfaceRequest{
		Id:         config.ID,
		Name:       config.Name,
		Address:    config.Address,
		ListenPort: config.ListenPort,
		Mtu:        config.MTU,
		Endpoint:   config.Endpoint,
	}))
	if err != nil {
		return domain.WireguardInterface{}, err
	}

	iface := response.Msg.GetInterface()
	if iface == nil {
		return domain.WireguardInterface{}, errors.New("interface not updated")
	}

	return domain.WireguardInterface{
		ID:         iface.GetId(),
		Name:       iface.GetName(),
		Address:    iface.GetAddress(),
		ListenPort: iface.GetListenPort(),
		PublicKey:  iface.GetPublicKey(),
		MTU:        iface.GetMtu(),
	}, nil
}

func (repo *AdminRPCWireguardRepository) DeleteInterface(ctx context.Context, interfaceID string) error {
	_, err := repo.client.DeleteInterface(ctx, connect.NewRequest(&adminv1.DeleteAdminInterfaceRequest{Id: interfaceID}))
	return err
}

func (repo *AdminRPCWireguardRepository) UpdatePeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error {
	_, err := repo.client.UpdateWireguardPeerAllowedIPs(ctx, connect.NewRequest(&adminv1.UpdateWireguardPeerAllowedIPsRequest{
		InterfaceId: interfaceID,
		PeerId:      peerID,
		AllowedIps:  allowedIPs,
	}))
	return err
}

func (repo *AdminRPCWireguardRepository) CreatePeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (domain.WireguardPeer, error) {
	response, err := repo.client.CreateWireguardPeer(ctx, connect.NewRequest(&adminv1.CreateWireguardPeerRequest{
		InterfaceId: interfaceID,
		Endpoint:    endpoint,
		AllowedIps:  allowedIPs,
	}))
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	return domain.WireguardPeer{
		ID:          response.Msg.GetPeerId(),
		InterfaceID: response.Msg.GetInterfaceId(),
		AllowedIP:   response.Msg.GetAllowedIp(),
		Config:      response.Msg.GetPeerConfig(),
	}, nil
}

func (repo *AdminRPCWireguardRepository) DeletePeer(ctx context.Context, peerID string) error {
	_, err := repo.client.DeleteWireguardPeer(ctx, connect.NewRequest(&adminv1.DeleteWireguardPeerRequest{PeerId: peerID}))
	return err
}

func (repo *AdminRPCWireguardRepository) ListPeerStats(ctx context.Context) ([]domain.PeerStat, error) {
	response, err := repo.client.ListPeerStats(ctx, connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return nil, err
	}

	stats := make([]domain.PeerStat, 0, len(response.Msg.Stats))
	for _, stat := range response.Msg.Stats {
		stats = append(stats, domain.PeerStat{
			PeerID:          stat.GetPeerId(),
			InterfaceID:     stat.GetInterfaceId(),
			RxBytes:         stat.GetRxBytes(),
			TxBytes:         stat.GetTxBytes(),
			LastHandshakeAt: stat.GetLastHandshakeAt(),
		})
	}

	return stats, nil
}

func (repo *AdminRPCWireguardRepository) ListConfigs(ctx context.Context, interfaceID string) ([]domain.WireguardConfig, error) {
	response, err := repo.client.ListWireguardConfigs(ctx, connect.NewRequest(&adminv1.ListWireguardConfigsRequest{InterfaceId: interfaceID}))
	if err != nil {
		return nil, err
	}

	configs := make([]domain.WireguardConfig, 0, len(response.Msg.Configs))
	for _, config := range response.Msg.Configs {
		configs = append(configs, domain.WireguardConfig{
			InterfaceID: config.GetInterfaceId(),
			Config:      config.GetConfig(),
		})
	}

	return configs, nil
}

func (repo *AdminRPCWireguardRepository) ListFirewallRules(ctx context.Context) (string, error) {
	response, err := repo.client.GetFirewallRules(ctx, connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return "", err
	}

	return response.Msg.GetRules(), nil
}

// EnsureFirewallChain is not supported for RPC repository
func (repo *AdminRPCWireguardRepository) EnsureFirewallChain(ctx context.Context) error {
	return errors.New("firewall chain management is not supported for RPC repository")
}

// SyncPeerFirewallRules is not supported for RPC repository
func (repo *AdminRPCWireguardRepository) SyncPeerFirewallRules(ctx context.Context, interfaceID string, peerAllowedIP string, allowedIPs []string) error {
	return errors.New("firewall rule sync is not supported for RPC repository")
}

// RemovePeerFirewallRules is not supported for RPC repository
func (repo *AdminRPCWireguardRepository) RemovePeerFirewallRules(ctx context.Context, peerAllowedIP string) error {
	return errors.New("firewall rule removal is not supported for RPC repository")
}
