package domain

import (
	"context"
	"time"
)

type WireguardInterface struct {
	ID         string
	Name       string
	Address    string
	ListenPort uint32
	PublicKey  string
	MTU        uint32
}

type InterfaceConfig struct {
	ID         string
	Name       string
	Address    string
	ListenPort uint32
	MTU        uint32
	Endpoint   string
}

type AdminInterface struct {
	ID         string
	Name       string
	Address    string
	ListenPort uint32
	PublicKey  string
	MTU        uint32
	Endpoint   string
}

type WireguardPeer struct {
	ID          string
	InterfaceID string
	AllowedIP   string
	Config      string
}

type PeerStat struct {
	PeerID          string
	InterfaceID     string
	RxBytes         uint64
	TxBytes         uint64
	LastHandshakeAt int64
}

type PeerStatus struct {
	PeerID          string
	InterfaceID     string
	InterfaceName   string
	RxBytes         uint64
	TxBytes         uint64
	LastHandshakeAt int64
}

type WireguardConfig struct {
	InterfaceID string
	Config      string
}

type WireguardRepository interface {
	ListInterfaces(ctx context.Context) ([]WireguardInterface, error)
	GetInterface(ctx context.Context, interfaceID string) (WireguardInterface, error)
	CreateInterface(ctx context.Context, config InterfaceConfig) (WireguardInterface, error)
	UpdateInterface(ctx context.Context, config InterfaceConfig) (WireguardInterface, error)
	DeleteInterface(ctx context.Context, interfaceID string) error
	CreatePeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (WireguardPeer, error)
	UpdatePeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error
	DeletePeer(ctx context.Context, peerID string) error
	ListPeerStats(ctx context.Context) ([]PeerStat, error)
	ListFirewallRules(ctx context.Context) (string, error)
	ListConfigs(ctx context.Context, interfaceID string) ([]WireguardConfig, error)
	EnsureFirewallChain(ctx context.Context) error
	SyncPeerFirewallRules(ctx context.Context, interfaceID string, peerAllowedIP string, allowedIPs []string) error
	RemovePeerFirewallRules(ctx context.Context, peerAllowedIP string) error
}

type PeerRecord struct {
	Email       string
	PeerID      string
	InterfaceID string
	AllowedIP   string
	Config      string
	CreatedAt   time.Time
}

type PeerStore interface {
	GetByEmail(ctx context.Context, email string) (PeerRecord, error)
	GetByEmailAndInterface(ctx context.Context, email string, interfaceID string) (PeerRecord, error)
	GetByPeerID(ctx context.Context, peerID string) (PeerRecord, error)
	List(ctx context.Context) ([]PeerRecord, error)
	ListByEmail(ctx context.Context, email string) ([]PeerRecord, error)
	ListByInterface(ctx context.Context, interfaceID string) ([]PeerRecord, error)

	Create(ctx context.Context, record PeerRecord) error
	UpdateConfig(ctx context.Context, peerID string, config string) error
	DeleteByPeerID(ctx context.Context, peerID string) error
	DeleteByInterface(ctx context.Context, interfaceID string) error
}

type InterfaceStore interface {
	Get(ctx context.Context, id string) (InterfaceConfig, error)
	List(ctx context.Context) ([]InterfaceConfig, error)
	Create(ctx context.Context, config InterfaceConfig) error
	Update(ctx context.Context, config InterfaceConfig) error
	Delete(ctx context.Context, id string) error
}

type AllowedEmailStore interface {
	ListByInterface(ctx context.Context, interfaceID string) ([]AllowedEmail, error)
	ListInterfaceIDsByEmail(ctx context.Context, email string) ([]string, error)
	Exists(ctx context.Context, interfaceID string, email string) (bool, error)
	Create(ctx context.Context, interfaceID string, email string) error
	Delete(ctx context.Context, interfaceID string, email string) error
	DeleteByInterface(ctx context.Context, interfaceID string) error
}

type AllowedEmail struct {
	InterfaceID string
	Email       string
	CreatedAt   time.Time
}

type InterfaceRoute struct {
	InterfaceID string
	CIDR        string
	CreatedAt   time.Time
}

type PeerRoute struct {
	PeerID    string
	CIDR      string
	CreatedAt time.Time
}

type InterfaceRouteStore interface {
	ListByInterface(ctx context.Context, interfaceID string) ([]InterfaceRoute, error)
	Create(ctx context.Context, interfaceID string, cidr string) error
	Delete(ctx context.Context, interfaceID string, cidr string) error
	DeleteByInterface(ctx context.Context, interfaceID string) error
}

type PeerRouteStore interface {
	ListByPeer(ctx context.Context, peerID string) ([]PeerRoute, error)
	Create(ctx context.Context, peerID string, cidr string) error
	Delete(ctx context.Context, peerID string, cidr string) error
	DeleteByPeer(ctx context.Context, peerID string) error
}
