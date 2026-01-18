package infra

import "errors"

var ErrInterfaceNotFound = errors.New("wireguard interface not found")
var ErrPeerNotFound = errors.New("wireguard peer not found")
