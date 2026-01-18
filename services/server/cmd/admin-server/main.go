package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/nomuken/william/services/server/gen/proto/admin/v1/adminv1connect"
	"github.com/nomuken/william/services/server/internal/domain"
	"github.com/nomuken/william/services/server/internal/infra"
	"github.com/nomuken/william/services/server/internal/transport/connecthandler"
	"github.com/nomuken/william/services/server/internal/usecase"
)

func main() {
	database, err := infra.OpenDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := infra.RunMigrations(database); err != nil {
		log.Fatal(err)
	}

	peerStore := infra.NewSQLPeerStore(database)
	interfaceStore := infra.NewSQLInterfaceStore(database)
	allowedEmailStore := infra.NewSQLAllowedEmailStore(database)
	interfaceRouteStore := infra.NewSQLInterfaceRouteStore(database)
	peerRouteStore := infra.NewSQLPeerRouteStore(database)

	devMode := os.Getenv("WILLIAM_DEV") == "1"
	var repository domain.WireguardRepository
	if devMode {
		repository = infra.NewMockWireguardRepository(interfaceStore, peerStore)
	} else {
		repository = infra.NewCommandWireguardRepository()
		infra.BootstrapWireguardOrFatal(context.Background(), repository, interfaceStore, peerStore, interfaceRouteStore, peerRouteStore)
	}

	adminService := usecase.NewAdminService(repository, peerStore, interfaceStore, allowedEmailStore, interfaceRouteStore, peerRouteStore)

	adminHandler := connecthandler.NewAdminHandler(adminService)

	adminPath, adminConnectHandler := adminv1connect.NewWilliamAdminServiceHandler(adminHandler)
	mux := http.NewServeMux()
	mux.Handle(adminPath, adminConnectHandler)

	addr := os.Getenv("WILLIAM_ADMIN_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	log.Printf("William admin server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
