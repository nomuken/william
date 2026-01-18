package connecthandler

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	williamv1 "github.com/nomuken/william/services/server/gen/proto/server/v1"
	"github.com/nomuken/william/services/server/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type WilliamHandler struct {
	wireguardUsecase usecase.WireguardUsecase
}

func NewWilliamHandler(wireguardUsecase usecase.WireguardUsecase) *WilliamHandler {
	return &WilliamHandler{wireguardUsecase: wireguardUsecase}
}

func (handler *WilliamHandler) ListWireguardInterfaces(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[williamv1.ListWireguardInterfacesResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}

	interfaces, err := handler.wireguardUsecase.ListInterfaces(ctx, email)
	if err != nil {
		return nil, err
	}

	items := make([]*williamv1.WireguardInterface, 0, len(interfaces))
	for _, item := range interfaces {
		items = append(items, &williamv1.WireguardInterface{
			Id:         item.ID,
			Name:       item.Name,
			Address:    item.Address,
			ListenPort: item.ListenPort,
			PublicKey:  item.PublicKey,
			Mtu:        item.MTU,
		})
	}

	response := &williamv1.ListWireguardInterfacesResponse{Interfaces: items}
	return connect.NewResponse(response), nil
}

func (handler *WilliamHandler) CreateWireguardPeer(ctx context.Context, req *connect.Request[williamv1.CreateWireguardPeerRequest]) (*connect.Response[williamv1.CreateWireguardPeerResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}

	peer, err := handler.wireguardUsecase.CreatePeer(ctx, email, req.Msg.GetWireguardInterfaceId())
	if err != nil {
		if errors.Is(err, usecase.ErrPeerAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		if errors.Is(err, usecase.ErrEmailNotAllowed) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	response := &williamv1.CreateWireguardPeerResponse{
		PeerId:     peer.ID,
		PeerConfig: peer.Config,
	}
	return connect.NewResponse(response), nil
}

func (handler *WilliamHandler) GetMyWireguardPeer(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[williamv1.GetMyWireguardPeerResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}

	record, err := handler.wireguardUsecase.GetPeerByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, usecase.ErrPeerNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, usecase.ErrPeerForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, err
	}

	response := &williamv1.GetMyWireguardPeerResponse{
		PeerId:     record.PeerID,
		PeerConfig: record.Config,
	}
	return connect.NewResponse(response), nil
}

func (handler *WilliamHandler) GetMyWireguardPeerByInterface(ctx context.Context, req *connect.Request[williamv1.GetMyWireguardPeerByInterfaceRequest]) (*connect.Response[williamv1.GetMyWireguardPeerByInterfaceResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}

	record, err := handler.wireguardUsecase.GetPeerByEmailAndInterface(ctx, email, req.Msg.GetInterfaceId())
	if err != nil {
		if errors.Is(err, usecase.ErrPeerNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, usecase.ErrPeerForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, usecase.ErrInterfaceNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	response := &williamv1.GetMyWireguardPeerByInterfaceResponse{
		PeerId:     record.PeerID,
		PeerConfig: record.Config,
	}
	return connect.NewResponse(response), nil
}

func (handler *WilliamHandler) DeleteWireguardPeer(ctx context.Context, req *connect.Request[williamv1.DeleteWireguardPeerRequest]) (*connect.Response[williamv1.DeleteWireguardPeerResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}

	if err := handler.wireguardUsecase.DeletePeer(ctx, email, req.Msg.GetPeerId()); err != nil {
		if errors.Is(err, usecase.ErrPeerForbidden) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		if errors.Is(err, usecase.ErrPeerNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}

	return connect.NewResponse(&williamv1.DeleteWireguardPeerResponse{}), nil
}

func (handler *WilliamHandler) ListPeerStatuses(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[williamv1.ListPeerStatusesResponse], error) {
	email, err := emailFromHeader(req)
	if err != nil {
		return nil, err
	}
	statuses, err := handler.wireguardUsecase.ListPeerStatuses(ctx, email)
	if err != nil {
		return nil, err
	}

	items := make([]*williamv1.PeerStatus, 0, len(statuses))
	for _, stat := range statuses {
		items = append(items, &williamv1.PeerStatus{
			PeerId:          stat.PeerID,
			InterfaceId:     stat.InterfaceID,
			InterfaceName:   stat.InterfaceName,
			RxBytes:         stat.RxBytes,
			TxBytes:         stat.TxBytes,
			LastHandshakeAt: stat.LastHandshakeAt,
		})
	}

	return connect.NewResponse(&williamv1.ListPeerStatusesResponse{Statuses: items}), nil
}

func emailFromHeader(request interface{ Header() http.Header }) (string, error) {
	email := request.Header().Get("X-Email")
	if email == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("X-Email header is required"))
	}
	return email, nil
}
