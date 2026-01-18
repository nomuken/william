package connecthandler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	adminv1 "github.com/nomuken/william/services/server/gen/proto/admin/v1"
	"github.com/nomuken/william/services/server/internal/domain"
	"github.com/nomuken/william/services/server/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AdminHandler struct {
	adminUsecase usecase.AdminUsecase
}

func NewAdminHandler(adminUsecase usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: adminUsecase}
}

func (handler *AdminHandler) ListInterfaces(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[adminv1.ListAdminInterfacesResponse], error) {
	interfaces, err := handler.adminUsecase.ListInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*adminv1.AdminWireguardInterface, 0, len(interfaces))
	for _, item := range interfaces {
		items = append(items, &adminv1.AdminWireguardInterface{
			Id:         item.ID,
			Name:       item.Name,
			Address:    item.Address,
			ListenPort: item.ListenPort,
			PublicKey:  item.PublicKey,
			Mtu:        item.MTU,
			Endpoint:   item.Endpoint,
		})
	}

	return connect.NewResponse(&adminv1.ListAdminInterfacesResponse{Interfaces: items}), nil
}

func (handler *AdminHandler) GetInterface(ctx context.Context, req *connect.Request[adminv1.GetAdminInterfaceRequest]) (*connect.Response[adminv1.GetAdminInterfaceResponse], error) {
	iface, err := handler.adminUsecase.GetInterface(ctx, req.Msg.GetId())
	if err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	response := &adminv1.GetAdminInterfaceResponse{Interface: adminInterfaceToProto(iface)}
	return connect.NewResponse(response), nil
}

func (handler *AdminHandler) CreateInterface(ctx context.Context, req *connect.Request[adminv1.CreateAdminInterfaceRequest]) (*connect.Response[adminv1.CreateAdminInterfaceResponse], error) {
	config := domain.InterfaceConfig{
		ID:         req.Msg.GetName(),
		Name:       req.Msg.GetName(),
		Address:    req.Msg.GetAddress(),
		ListenPort: req.Msg.GetListenPort(),
		MTU:        req.Msg.GetMtu(),
		Endpoint:   req.Msg.GetEndpoint(),
	}
	iface, err := handler.adminUsecase.CreateInterface(ctx, config)
	if err != nil {
		return nil, err
	}

	response := &adminv1.CreateAdminInterfaceResponse{Interface: adminInterfaceToProto(iface)}
	return connect.NewResponse(response), nil
}

func (handler *AdminHandler) UpdateInterface(ctx context.Context, req *connect.Request[adminv1.UpdateAdminInterfaceRequest]) (*connect.Response[adminv1.UpdateAdminInterfaceResponse], error) {
	config := domain.InterfaceConfig{
		ID:         req.Msg.GetId(),
		Name:       req.Msg.GetName(),
		Address:    req.Msg.GetAddress(),
		ListenPort: req.Msg.GetListenPort(),
		MTU:        req.Msg.GetMtu(),
		Endpoint:   req.Msg.GetEndpoint(),
	}
	iface, err := handler.adminUsecase.UpdateInterface(ctx, config)
	if err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	response := &adminv1.UpdateAdminInterfaceResponse{Interface: adminInterfaceToProto(iface)}
	return connect.NewResponse(response), nil
}

func (handler *AdminHandler) DeleteInterface(ctx context.Context, req *connect.Request[adminv1.DeleteAdminInterfaceRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeleteInterface(ctx, req.Msg.GetId()); err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) ListAllowedEmails(ctx context.Context, req *connect.Request[adminv1.ListAllowedEmailsRequest]) (*connect.Response[adminv1.ListAllowedEmailsResponse], error) {
	emails, err := handler.adminUsecase.ListAllowedEmails(ctx, req.Msg.GetInterfaceId())
	if err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	items := make([]*adminv1.AllowedEmail, 0, len(emails))
	for _, email := range emails {
		items = append(items, &adminv1.AllowedEmail{
			InterfaceId: email.InterfaceID,
			Email:       email.Email,
			CreatedAt:   timestamppb.New(email.CreatedAt),
		})
	}

	return connect.NewResponse(&adminv1.ListAllowedEmailsResponse{Emails: items}), nil
}

func (handler *AdminHandler) CreateAllowedEmail(ctx context.Context, req *connect.Request[adminv1.CreateAllowedEmailRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.CreateAllowedEmail(ctx, req.Msg.GetInterfaceId(), req.Msg.GetEmail()); err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) DeleteAllowedEmail(ctx context.Context, req *connect.Request[adminv1.DeleteAllowedEmailRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeleteAllowedEmail(ctx, req.Msg.GetInterfaceId(), req.Msg.GetEmail()); err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) ListPeers(ctx context.Context, req *connect.Request[adminv1.ListAdminPeersRequest]) (*connect.Response[adminv1.ListAdminPeersResponse], error) {
	peers, err := handler.adminUsecase.ListPeers(ctx, req.Msg.GetInterfaceId())
	if err != nil {
		return nil, err
	}

	items := make([]*adminv1.AdminPeer, 0, len(peers))
	for _, peer := range peers {
		items = append(items, &adminv1.AdminPeer{
			PeerId:      peer.PeerID,
			Email:       peer.Email,
			InterfaceId: peer.InterfaceID,
			AllowedIp:   peer.AllowedIP,
			CreatedAt:   timestamppb.New(peer.CreatedAt),
		})
	}

	return connect.NewResponse(&adminv1.ListAdminPeersResponse{Peers: items}), nil
}

func (handler *AdminHandler) DeletePeer(ctx context.Context, req *connect.Request[adminv1.DeleteAdminPeerRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeletePeer(ctx, req.Msg.GetPeerId()); err != nil {
		if errors.Is(err, usecase.ErrPeerNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) CreateWireguardPeer(ctx context.Context, req *connect.Request[adminv1.CreateWireguardPeerRequest]) (*connect.Response[adminv1.CreateWireguardPeerResponse], error) {
	peer, err := handler.adminUsecase.CreateWireguardPeer(ctx, req.Msg.GetInterfaceId(), req.Msg.GetEndpoint(), req.Msg.GetAllowedIps())
	if err != nil {
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	response := &adminv1.CreateWireguardPeerResponse{
		InterfaceId: peer.InterfaceID,
		PeerId:      peer.ID,
		AllowedIp:   peer.AllowedIP,
		PeerConfig:  peer.Config,
	}
	return connect.NewResponse(response), nil
}

func (handler *AdminHandler) DeleteWireguardPeer(ctx context.Context, req *connect.Request[adminv1.DeleteWireguardPeerRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeleteWireguardPeer(ctx, req.Msg.GetPeerId()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) UpdateWireguardPeerAllowedIPs(ctx context.Context, req *connect.Request[adminv1.UpdateWireguardPeerAllowedIPsRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.UpdateWireguardPeerAllowedIPs(ctx, req.Msg.GetInterfaceId(), req.Msg.GetPeerId(), req.Msg.GetAllowedIps()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) ListInterfaceRoutes(ctx context.Context, req *connect.Request[adminv1.ListInterfaceRoutesRequest]) (*connect.Response[adminv1.ListInterfaceRoutesResponse], error) {
	routes, err := handler.adminUsecase.ListInterfaceRoutes(ctx, req.Msg.GetInterfaceId())
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.InterfaceRoute, 0, len(routes))
	for _, route := range routes {
		items = append(items, &adminv1.InterfaceRoute{
			InterfaceId: route.InterfaceID,
			Cidr:        route.CIDR,
			CreatedAt:   timestamppb.New(route.CreatedAt),
		})
	}
	return connect.NewResponse(&adminv1.ListInterfaceRoutesResponse{Routes: items}), nil
}

func (handler *AdminHandler) CreateInterfaceRoute(ctx context.Context, req *connect.Request[adminv1.CreateInterfaceRouteRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.CreateInterfaceRoute(ctx, req.Msg.GetInterfaceId(), req.Msg.GetCidr()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) DeleteInterfaceRoute(ctx context.Context, req *connect.Request[adminv1.DeleteInterfaceRouteRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeleteInterfaceRoute(ctx, req.Msg.GetInterfaceId(), req.Msg.GetCidr()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) ListPeerRoutes(ctx context.Context, req *connect.Request[adminv1.ListPeerRoutesRequest]) (*connect.Response[adminv1.ListPeerRoutesResponse], error) {
	routes, err := handler.adminUsecase.ListPeerRoutes(ctx, req.Msg.GetPeerId())
	if err != nil {
		return nil, err
	}
	items := make([]*adminv1.PeerRoute, 0, len(routes))
	for _, route := range routes {
		items = append(items, &adminv1.PeerRoute{
			PeerId:    route.PeerID,
			Cidr:      route.CIDR,
			CreatedAt: timestamppb.New(route.CreatedAt),
		})
	}
	return connect.NewResponse(&adminv1.ListPeerRoutesResponse{Routes: items}), nil
}

func (handler *AdminHandler) CreatePeerRoute(ctx context.Context, req *connect.Request[adminv1.CreatePeerRouteRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.CreatePeerRoute(ctx, req.Msg.GetPeerId(), req.Msg.GetCidr()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) DeletePeerRoute(ctx context.Context, req *connect.Request[adminv1.DeletePeerRouteRequest]) (*connect.Response[emptypb.Empty], error) {
	if err := handler.adminUsecase.DeletePeerRoute(ctx, req.Msg.GetPeerId(), req.Msg.GetCidr()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (handler *AdminHandler) ListPeerStats(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[adminv1.ListPeerStatsResponse], error) {
	stats, err := handler.adminUsecase.ListPeerStats(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*adminv1.PeerStat, 0, len(stats))
	for _, stat := range stats {
		items = append(items, &adminv1.PeerStat{
			PeerId:          stat.PeerID,
			InterfaceId:     stat.InterfaceID,
			RxBytes:         stat.RxBytes,
			TxBytes:         stat.TxBytes,
			LastHandshakeAt: stat.LastHandshakeAt,
		})
	}
	return connect.NewResponse(&adminv1.ListPeerStatsResponse{Stats: items}), nil
}

func (handler *AdminHandler) ListWireguardConfigs(ctx context.Context, req *connect.Request[adminv1.ListWireguardConfigsRequest]) (*connect.Response[adminv1.ListWireguardConfigsResponse], error) {
	configs, err := handler.adminUsecase.ListWireguardConfigs(ctx, req.Msg.GetInterfaceId())
	if err != nil {
		return nil, err
	}

	items := make([]*adminv1.WireguardConfig, 0, len(configs))
	for _, config := range configs {
		items = append(items, &adminv1.WireguardConfig{
			InterfaceId: config.InterfaceID,
			Config:      config.Config,
		})
	}
	return connect.NewResponse(&adminv1.ListWireguardConfigsResponse{Configs: items}), nil
}

func (handler *AdminHandler) GetFirewallRules(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[adminv1.GetFirewallRulesResponse], error) {
	rules, err := handler.adminUsecase.GetFirewallRules(ctx)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&adminv1.GetFirewallRulesResponse{Rules: rules}), nil
}

func adminInterfaceToProto(item domain.AdminInterface) *adminv1.AdminWireguardInterface {
	return &adminv1.AdminWireguardInterface{
		Id:         item.ID,
		Name:       item.Name,
		Address:    item.Address,
		ListenPort: item.ListenPort,
		PublicKey:  item.PublicKey,
		Mtu:        item.MTU,
		Endpoint:   item.Endpoint,
	}
}
