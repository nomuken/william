package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nomuken/william/services/server/gen/proto/server/v1/williamv1connect"
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

	repository := infra.NewAdminRPCWireguardRepository(nil)
	peerStore := infra.NewSQLPeerStore(database)
	interfaceStore := infra.NewSQLInterfaceStore(database)
	allowedEmailStore := infra.NewSQLAllowedEmailStore(database)
	interfaceRouteStore := infra.NewSQLInterfaceRouteStore(database)
	wireguardService := usecase.NewWireguardService(repository, peerStore, interfaceStore, allowedEmailStore, interfaceRouteStore)

	userHandler := connecthandler.NewWilliamHandler(wireguardService)

	path, connectHandler := williamv1connect.NewWilliamServiceHandler(userHandler)
	mux := http.NewServeMux()
	mux.Handle(path, connectHandler)

	addr := os.Getenv("WILLIAM_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("William server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
