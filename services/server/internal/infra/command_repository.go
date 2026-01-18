package infra

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/nomuken/william/services/server/internal/domain"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	RunWithInput(ctx context.Context, input string, name string, args ...string) (string, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	return formatCommandOutput(name, args, output, err)
}

func (execRunner) RunWithInput(ctx context.Context, input string, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = strings.NewReader(input)
	output, err := command.CombinedOutput()
	return formatCommandOutput(name, args, output, err)
}

type CommandWireguardRepository struct {
	runner CommandRunner
}

func NewCommandWireguardRepository() *CommandWireguardRepository {
	return &CommandWireguardRepository{runner: execRunner{}}
}

func (repo *CommandWireguardRepository) ListInterfaces(ctx context.Context) ([]domain.WireguardInterface, error) {
	// wg show interfaces
	output, err := repo.runner.Run(ctx, "wg", "show", "interfaces")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	interfaceNames := strings.Fields(output)
	interfaces := make([]domain.WireguardInterface, 0, len(interfaceNames))
	for _, name := range interfaceNames {
		item, err := repo.describeInterface(ctx, name)
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, item)
	}

	return interfaces, nil
}

func (repo *CommandWireguardRepository) GetInterface(ctx context.Context, interfaceID string) (domain.WireguardInterface, error) {
	return repo.describeInterface(ctx, interfaceID)
}

func (repo *CommandWireguardRepository) CreateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	if _, err := repo.runner.Run(ctx, "ip", "link", "add", "dev", config.ID, "type", "wireguard"); err != nil {
		return domain.WireguardInterface{}, err
	}

	privateKey, err := repo.runner.Run(ctx, "wg", "genkey")
	if err != nil {
		return domain.WireguardInterface{}, err
	}
	privateKey = strings.TrimSpace(privateKey)

	listenPort := strconv.FormatUint(uint64(config.ListenPort), 10)
	if _, err := repo.runner.RunWithInput(ctx, privateKey+"\n", "wg", "set", config.ID, "private-key", "/dev/fd/0", "listen-port", listenPort); err != nil {
		return domain.WireguardInterface{}, err
	}

	if _, err := repo.runner.Run(ctx, "ip", "address", "add", config.Address, "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	if _, err := repo.runner.Run(ctx, "ip", "link", "set", "mtu", strconv.FormatUint(uint64(config.MTU), 10), "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	if _, err := repo.runner.Run(ctx, "ip", "link", "set", "up", "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	return repo.describeInterface(ctx, config.ID)
}

func (repo *CommandWireguardRepository) UpdateInterface(ctx context.Context, config domain.InterfaceConfig) (domain.WireguardInterface, error) {
	if _, err := repo.runner.Run(ctx, "ip", "address", "replace", config.Address, "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	listenPort := strconv.FormatUint(uint64(config.ListenPort), 10)
	if _, err := repo.runner.Run(ctx, "wg", "set", config.ID, "listen-port", listenPort); err != nil {
		return domain.WireguardInterface{}, err
	}

	if _, err := repo.runner.Run(ctx, "ip", "link", "set", "mtu", strconv.FormatUint(uint64(config.MTU), 10), "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	if _, err := repo.runner.Run(ctx, "ip", "link", "set", "up", "dev", config.ID); err != nil {
		return domain.WireguardInterface{}, err
	}

	return repo.describeInterface(ctx, config.ID)
}

func (repo *CommandWireguardRepository) DeleteInterface(ctx context.Context, interfaceID string) error {
	_, err := repo.runner.Run(ctx, "ip", "link", "delete", "dev", interfaceID)
	return err
}

func (repo *CommandWireguardRepository) CreatePeer(ctx context.Context, interfaceID string, endpoint string, allowedIPs []string) (domain.WireguardPeer, error) {
	interfaceInfo, err := repo.describeInterface(ctx, interfaceID)
	if err != nil {
		if errors.Is(err, ErrInterfaceNotFound) {
			return domain.WireguardPeer{}, err
		}
		return domain.WireguardPeer{}, err
	}

	prefix, interfaceAddr, err := repo.interfacePrefix(ctx, interfaceID)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	usedAddrs, err := repo.listAllowedIPs(ctx, interfaceID)
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	usedAddrs[interfaceAddr] = struct{}{}
	usedAddrs[prefix.Masked().Addr()] = struct{}{}

	peerAddr, err := nextAvailableIPv4(prefix, interfaceAddr, usedAddrs)
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	allowedIP := fmt.Sprintf("%s/32", peerAddr.String())

	// wg genkey
	privateKey, err := repo.runner.Run(ctx, "wg", "genkey")
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	privateKey = strings.TrimSpace(privateKey)

	// wg pubkey
	publicKey, err := repo.runner.RunWithInput(ctx, privateKey+"\n", "wg", "pubkey")
	if err != nil {
		return domain.WireguardPeer{}, err
	}
	publicKey = strings.TrimSpace(publicKey)

	allowedIPs = normalizeAllowedIPs(allowedIP, allowedIPs)
	if _, err := repo.runner.Run(ctx, "wg", "set", interfaceID, "peer", publicKey, "allowed-ips", strings.Join(allowedIPs, ",")); err != nil {
		return domain.WireguardPeer{}, err
	}

	if endpoint == "" {
		return domain.WireguardPeer{}, errors.New("endpoint is required")
	}
	config, err := buildPeerConfig(privateKey, allowedIP, interfaceInfo.PublicKey, interfaceInfo.ListenPort, endpoint, allowedIPs)
	if err != nil {
		return domain.WireguardPeer{}, err
	}

	log.Printf("wireguard peer created: interface=%s peer=%s ip=%s", interfaceID, publicKey, allowedIP)
	return domain.WireguardPeer{
		ID:          publicKey,
		InterfaceID: interfaceID,
		AllowedIP:   allowedIP,
		Config:      config,
	}, nil
}

func (repo *CommandWireguardRepository) UpdatePeerAllowedIPs(ctx context.Context, interfaceID string, peerID string, allowedIPs []string) error {
	if len(allowedIPs) == 0 {
		return errors.New("allowed IPs are required")
	}
	_, err := repo.runner.Run(ctx, "wg", "set", interfaceID, "peer", peerID, "allowed-ips", strings.Join(allowedIPs, ","))
	return err
}

func (repo *CommandWireguardRepository) DeletePeer(ctx context.Context, peerID string) error {
	interfaces, err := repo.ListInterfaces(ctx)
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		// wg show <interface> peers
		peers, err := repo.runner.Run(ctx, "wg", "show", iface.Name, "peers")
		if err != nil {
			return err
		}
		for _, peer := range strings.Fields(peers) {
			if peer != peerID {
				continue
			}
			// wg set <interface> peer <peerID> remove
			if _, err := repo.runner.Run(ctx, "wg", "set", iface.Name, "peer", peerID, "remove"); err != nil {
				return err
			}
			return nil
		}
	}

	return ErrPeerNotFound
}

func (repo *CommandWireguardRepository) ListPeerStats(ctx context.Context) ([]domain.PeerStat, error) {
	interfaces, err := repo.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	statsByPeer := make(map[string]domain.PeerStat)
	for _, iface := range interfaces {
		transfers, err := repo.readPeerTransfers(ctx, iface.Name)
		if err != nil {
			return nil, err
		}
		handshakes, err := repo.readPeerHandshakes(ctx, iface.Name)
		if err != nil {
			return nil, err
		}

		for peerID, transfer := range transfers {
			statsByPeer[peerID] = domain.PeerStat{
				PeerID:          peerID,
				InterfaceID:     iface.ID,
				RxBytes:         transfer.rxBytes,
				TxBytes:         transfer.txBytes,
				LastHandshakeAt: handshakes[peerID],
			}
		}
		for peerID, handshake := range handshakes {
			if _, ok := statsByPeer[peerID]; ok {
				continue
			}
			statsByPeer[peerID] = domain.PeerStat{
				PeerID:          peerID,
				InterfaceID:     iface.ID,
				LastHandshakeAt: handshake,
			}
		}
	}

	stats := make([]domain.PeerStat, 0, len(statsByPeer))
	for _, stat := range statsByPeer {
		stats = append(stats, stat)
	}
	return stats, nil
}

func (repo *CommandWireguardRepository) ListConfigs(ctx context.Context, interfaceID string) ([]domain.WireguardConfig, error) {
	interfaces, err := repo.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	configs := make([]domain.WireguardConfig, 0, len(interfaces))
	for _, iface := range interfaces {
		if interfaceID != "" && iface.ID != interfaceID {
			continue
		}
		configText, err := repo.runner.Run(ctx, "wg", "showconf", iface.Name)
		if err != nil {
			return nil, err
		}
		configs = append(configs, domain.WireguardConfig{
			InterfaceID: iface.ID,
			Config:      strings.TrimSpace(configText),
		})
	}

	if interfaceID != "" && len(configs) == 0 {
		return nil, ErrInterfaceNotFound
	}

	return configs, nil
}

func (repo *CommandWireguardRepository) ListFirewallRules(ctx context.Context) (string, error) {
	rules, err := repo.runner.Run(ctx, "iptables", "-S", "WILLIAM_FWD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(rules), nil
}

type peerTransfer struct {
	rxBytes uint64
	txBytes uint64
}

func (repo *CommandWireguardRepository) readPeerTransfers(ctx context.Context, interfaceID string) (map[string]peerTransfer, error) {
	output, err := repo.runner.Run(ctx, "wg", "show", interfaceID, "transfer")
	if err != nil {
		return nil, err
	}

	transfers := make(map[string]peerTransfer)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		rxBytes, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}
		txBytes, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}
		transfers[fields[0]] = peerTransfer{rxBytes: rxBytes, txBytes: txBytes}
	}

	return transfers, nil
}

func (repo *CommandWireguardRepository) readPeerHandshakes(ctx context.Context, interfaceID string) (map[string]int64, error) {
	output, err := repo.runner.Run(ctx, "wg", "show", interfaceID, "latest-handshakes")
	if err != nil {
		return nil, err
	}

	handshakes := make(map[string]int64)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}
		handshakes[fields[0]] = value
	}

	return handshakes, nil
}

func (repo *CommandWireguardRepository) describeInterface(ctx context.Context, name string) (domain.WireguardInterface, error) {
	// wg show <interface> public-key
	publicKey, err := repo.runner.Run(ctx, "wg", "show", name, "public-key")
	if err != nil {
		return domain.WireguardInterface{}, err
	}
	publicKey = strings.TrimSpace(publicKey)

	// wg show <interface> listen-port
	listenPortOutput, err := repo.runner.Run(ctx, "wg", "show", name, "listen-port")
	if err != nil {
		return domain.WireguardInterface{}, err
	}
	listenPort, err := strconv.ParseUint(strings.TrimSpace(listenPortOutput), 10, 32)
	if err != nil {
		return domain.WireguardInterface{}, fmt.Errorf("parse listen port: %w", err)
	}

	prefix, _, err := repo.interfacePrefix(ctx, name)
	if err != nil {
		return domain.WireguardInterface{}, err
	}

	mtu, err := repo.interfaceMTU(ctx, name)
	if err != nil {
		return domain.WireguardInterface{}, err
	}

	return domain.WireguardInterface{
		ID:         name,
		Name:       name,
		Address:    prefix.String(),
		ListenPort: uint32(listenPort),
		PublicKey:  publicKey,
		MTU:        uint32(mtu),
	}, nil
}

func (repo *CommandWireguardRepository) interfacePrefix(ctx context.Context, name string) (netip.Prefix, netip.Addr, error) {
	// ip -4 addr show dev <interface>
	output, err := repo.runner.Run(ctx, "ip", "-4", "addr", "show", "dev", name)
	if err != nil {
		if strings.Contains(output, "does not exist") {
			return netip.Prefix{}, netip.Addr{}, ErrInterfaceNotFound
		}
		return netip.Prefix{}, netip.Addr{}, err
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for index, field := range fields {
			if field == "inet" && index+1 < len(fields) {
				prefix, err := netip.ParsePrefix(fields[index+1])
				if err != nil {
					return netip.Prefix{}, netip.Addr{}, fmt.Errorf("parse interface address: %w", err)
				}
				return prefix, prefix.Addr(), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return netip.Prefix{}, netip.Addr{}, err
	}

	return netip.Prefix{}, netip.Addr{}, ErrInterfaceNotFound
}

func (repo *CommandWireguardRepository) interfaceMTU(ctx context.Context, name string) (int, error) {
	// ip link show dev <interface>
	output, err := repo.runner.Run(ctx, "ip", "link", "show", "dev", name)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(output)
	for index, field := range fields {
		if field == "mtu" && index+1 < len(fields) {
			mtu, err := strconv.Atoi(fields[index+1])
			if err != nil {
				return 0, fmt.Errorf("parse mtu: %w", err)
			}
			return mtu, nil
		}
	}

	return 0, errors.New("mtu not found")
}

func (repo *CommandWireguardRepository) listAllowedIPs(ctx context.Context, interfaceID string) (map[netip.Addr]struct{}, error) {
	// wg show <interface> allowed-ips
	output, err := repo.runner.Run(ctx, "wg", "show", interfaceID, "allowed-ips")
	if err != nil {
		return nil, err
	}

	used := make(map[netip.Addr]struct{})
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		for _, item := range strings.Split(fields[1], ",") {
			prefix, err := netip.ParsePrefix(strings.TrimSpace(item))
			if err != nil {
				continue
			}
			if prefix.Addr().Is4() {
				used[prefix.Addr()] = struct{}{}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return used, nil
}

func nextAvailableIPv4(prefix netip.Prefix, interfaceAddr netip.Addr, used map[netip.Addr]struct{}) (netip.Addr, error) {
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

//go:embed templates/peer.conf.tmpl
var peerTemplateFS embed.FS

var peerTemplate = template.Must(template.ParseFS(peerTemplateFS, "templates/peer.conf.tmpl"))

type peerTemplateData struct {
	PrivateKey      string
	Address         string
	ServerPublicKey string
	Endpoint        string
	ListenPort      uint32
	AllowedIPs      string
}

func buildPeerConfig(privateKey, address, serverPublicKey string, listenPort uint32, endpoint string, allowedIPs []string) (string, error) {
	data := peerTemplateData{
		PrivateKey:      privateKey,
		Address:         address,
		ServerPublicKey: serverPublicKey,
		Endpoint:        endpoint,
		ListenPort:      listenPort,
		AllowedIPs:      strings.Join(allowedIPs, ", "),
	}

	var buffer bytes.Buffer
	if err := peerTemplate.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func normalizeAllowedIPs(peerAllowedIP string, allowedIPs []string) []string {
	items := make([]string, 0, len(allowedIPs)+1)
	seen := make(map[string]struct{}, len(allowedIPs)+1)

	if peerAllowedIP != "" {
		items = append(items, peerAllowedIP)
		seen[peerAllowedIP] = struct{}{}
	}

	for _, item := range allowedIPs {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		items = append(items, trimmed)
	}

	return items
}

func formatCommandOutput(name string, args []string, output []byte, err error) (string, error) {
	trimmed := strings.TrimSpace(string(output))
	if err != nil {
		command := strings.TrimSpace(name + " " + strings.Join(args, " "))
		return trimmed, fmt.Errorf("command failed: %s: %w", command, err)
	}
	return trimmed, nil
}

// EnsureFirewallChain creates the WILLIAM_FWD chain if it doesn't exist and ensures it's called from FORWARD chain
func (repo *CommandWireguardRepository) EnsureFirewallChain(ctx context.Context) error {
	// Check if WILLIAM_FWD chain exists
	if _, err := repo.runner.Run(ctx, "iptables", "-L", "WILLIAM_FWD", "-n"); err != nil {
		// Chain doesn't exist, create it
		if _, err := repo.runner.Run(ctx, "iptables", "-N", "WILLIAM_FWD"); err != nil {
			return fmt.Errorf("create WILLIAM_FWD chain: %w", err)
		}
	}

	// Check if FORWARD chain calls WILLIAM_FWD
	output, err := repo.runner.Run(ctx, "iptables", "-S", "FORWARD")
	if err != nil {
		return fmt.Errorf("check FORWARD chain: %w", err)
	}

	if !strings.Contains(output, "-A FORWARD -j WILLIAM_FWD") {
		// Add jump to WILLIAM_FWD at the beginning of FORWARD chain
		if _, err := repo.runner.Run(ctx, "iptables", "-I", "FORWARD", "1", "-j", "WILLIAM_FWD"); err != nil {
			return fmt.Errorf("add WILLIAM_FWD to FORWARD chain: %w", err)
		}
	}

	return nil
}

// SyncPeerFirewallRules synchronizes iptables rules for a specific peer
func (repo *CommandWireguardRepository) SyncPeerFirewallRules(ctx context.Context, interfaceID string, peerAllowedIP string, allowedIPs []string) error {
	// First, ensure the firewall chain exists
	if err := repo.EnsureFirewallChain(ctx); err != nil {
		return err
	}

	// Remove old rules for this peer
	if err := repo.RemovePeerFirewallRules(ctx, peerAllowedIP); err != nil {
		return err
	}

	// Extract source IP from CIDR (e.g., "10.0.0.2/32" -> "10.0.0.2")
	sourceIP := strings.TrimSuffix(peerAllowedIP, "/32")

	// Add new rules for each allowed destination
	for _, destCIDR := range allowedIPs {
		// Skip the peer's own IP
		if destCIDR == peerAllowedIP {
			continue
		}

		// Add rule: allow traffic from peer to destination
		args := []string{
			"-A", "WILLIAM_FWD",
			"-i", interfaceID,
			"-s", sourceIP,
			"-d", destCIDR,
			"-j", "ACCEPT",
		}
		if _, err := repo.runner.Run(ctx, "iptables", args...); err != nil {
			return fmt.Errorf("add firewall rule for %s -> %s: %w", sourceIP, destCIDR, err)
		}
	}

	return nil
}

// RemovePeerFirewallRules removes all iptables rules associated with a peer's IP
func (repo *CommandWireguardRepository) RemovePeerFirewallRules(ctx context.Context, peerAllowedIP string) error {
	// Extract source IP from CIDR
	sourceIP := strings.TrimSuffix(peerAllowedIP, "/32")

	// Get current rules
	output, err := repo.runner.Run(ctx, "iptables", "-S", "WILLIAM_FWD")
	if err != nil {
		// Chain might not exist yet
		return nil
	}

	// Parse rules and delete matching ones
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		// Look for rules with this source IP
		if !strings.Contains(line, "-s "+sourceIP) {
			continue
		}

		// Convert "-A WILLIAM_FWD ..." to "-D WILLIAM_FWD ..."
		if !strings.HasPrefix(line, "-A ") {
			continue
		}
		deleteRule := "-D " + strings.TrimPrefix(line, "-A ")

		// Split into args and execute
		args := strings.Fields(deleteRule)
		if _, err := repo.runner.Run(ctx, "iptables", args...); err != nil {
			// Rule might have been already deleted, continue
			continue
		}
	}

	return scanner.Err()
}
