package main

import (
	"fmt"
	"log"
	"net/http"

	"auth-service/internal/core/config"
	"auth-service/internal/core/database"
	"auth-service/internal/modules/auth/httpapi"
	"github.com/gorilla/mux"
)

func main() {
	appCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbManager, err := database.NewManager(appCfg)
	if err != nil {
		log.Fatalf("Failed to init database manager: %v", err)
	}
	defer dbManager.CloseAll()

	if err := dbManager.EnsureDomainReady(appCfg.DefaultDomain); err != nil {
		log.Fatalf("Failed to prepare default domain DB: %v", err)
	}

	router := mux.NewRouter()
	httpapi.RegisterRoutes(router, dbManager, appCfg)

	fmt.Printf("Server starting on port %s\n", appCfg.Port)
	fmt.Printf("Default domain: %s\n", appCfg.DefaultDomain)
	fmt.Printf("Configured domains: %v\n", appCfg.DomainList())
	if err := http.ListenAndServe(":"+appCfg.Port, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
