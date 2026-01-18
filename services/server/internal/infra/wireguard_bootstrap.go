package infra

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/nomuken/william/services/server/internal/domain"
)

// BootstrapWireguard resets and restores wireguard state from the database.
func BootstrapWireguard(ctx context.Context, repository domain.WireguardRepository, interfaceStore domain.InterfaceStore, peerStore domain.PeerStore, interfaceRouteStore domain.InterfaceRouteStore, peerRouteStore domain.PeerRouteStore, runner CommandRunner) error {
	if runner == nil {
		runner = execRunner{}
	}

	excluded := loadExcludedInterfaces()

	interfaces, err := listWireguardInterfaces(ctx, runner)
	if err != nil {
		return err
	}

	for _, iface := range interfaces {
		if _, ok := excluded[iface]; ok {
			continue
		}
		if _, err := runner.Run(ctx, "ip", "link", "down", iface); err != nil {
			return err
		}
	}

	configs, err := interfaceStore.List(ctx)
	if err != nil {
		return err
	}

	for _, config := range configs {
		if _, err := repository.CreateInterface(ctx, config); err != nil {
			return err
		}

		interfaceRoutes, err := interfaceRouteStore.ListByInterface(ctx, config.ID)
		if err != nil {
			return err
		}

		peers, err := peerStore.ListByInterface(ctx, config.ID)
		if err != nil {
			return err
		}
		for _, peer := range peers {
			peerRoutes, err := peerRouteStore.ListByPeer(ctx, peer.PeerID)
			if err != nil {
				return err
			}
			allowedIPs := normalizeAllowedIPs(peer.AllowedIP, append(routeCIDRs(interfaceRoutes), peerRouteCIDRs(peerRoutes)...))
			if err := repository.UpdatePeerAllowedIPs(ctx, config.ID, peer.PeerID, allowedIPs); err != nil {
				return err
			}
		}
	}

	return nil
}

func listWireguardInterfaces(ctx context.Context, runner CommandRunner) ([]string, error) {
	output, err := runner.Run(ctx, "wg", "show", "interfaces")
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	return strings.Fields(output), nil
}

func loadExcludedInterfaces() map[string]struct{} {
	value := strings.TrimSpace(os.Getenv("WILLIAM_SYSTEM_WG_INTERFACES"))
	if value == "" {
		return map[string]struct{}{}
	}

	items := strings.Split(value, ",")
	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		result[name] = struct{}{}
	}

	return result
}

func BootstrapWireguardOrFatal(ctx context.Context, repository domain.WireguardRepository, interfaceStore domain.InterfaceStore, peerStore domain.PeerStore, interfaceRouteStore domain.InterfaceRouteStore, peerRouteStore domain.PeerRouteStore) {
	if err := BootstrapWireguard(ctx, repository, interfaceStore, peerStore, interfaceRouteStore, peerRouteStore, nil); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Fatalf("wireguard bootstrap failed: %v", err)
	}
}

func routeCIDRs(routes []domain.InterfaceRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, route := range routes {
		cidrs = append(cidrs, route.CIDR)
	}
	return cidrs
}

func peerRouteCIDRs(routes []domain.PeerRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, route := range routes {
		cidrs = append(cidrs, route.CIDR)
	}
	return cidrs
}
