package main

import (
	"log"
	"log/slog"
	"net/http"

	"auth-service/internal/core/config"
	"auth-service/internal/core/database"
	"auth-service/internal/core/logx"
	"auth-service/internal/core/metricsx"
	"auth-service/internal/modules/auth/httpapi"

	"github.com/gorilla/mux"
)

func main() {
	appCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, closeLogFile, err := logx.New(appCfg.ServiceName, appCfg.LogDir, appCfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := closeLogFile(); err != nil {
			slog.Error("failed to close log file", "log_type", "app_error", "error", err.Error())
		}
	}()

	if appCfg.MetricsEnabled {
		stopMetricsPush := metricsx.StartPushLoop(
			appCfg.ServiceName,
			appCfg.PushGatewayURL,
			appCfg.MetricsJob,
			appCfg.MetricsInstance,
			appCfg.MetricsInterval,
		)
		defer stopMetricsPush()
	}

	dbManager, err := database.NewManager(appCfg)
	if err != nil {
		logger.Error("failed to init database manager", "log_type", "startup", "error", err.Error())
		log.Fatalf("Failed to init database manager: %v", err)
	}
	defer dbManager.CloseAll()

	if err := dbManager.EnsureDomainReady(appCfg.DefaultDomain); err != nil {
		logger.Error("failed to prepare default domain database", "log_type", "startup", "error", err.Error())
		log.Fatalf("Failed to prepare default domain DB: %v", err)
	}

	router := mux.NewRouter()
	httpapi.RegisterRoutes(router, dbManager, appCfg)
	handler := httpapi.DomainMiddleware(appCfg)(logx.RecoverMiddleware()(logx.RequestMiddleware()(router)))

	logger.Info("server starting",
		"log_type", "startup",
		"port", appCfg.Port,
		"default_domain", appCfg.DefaultDomain,
		"domains", appCfg.DomainList(),
	)
	if err := http.ListenAndServe(":"+appCfg.Port, handler); err != nil {
		logger.Error("server stopped with error", "log_type", "startup", "error", err.Error())
		log.Fatalf("Server error: %v", err)
	}
}
