package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/m04kA/SMC-SellerService/internal/api/handlers/create_company"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/create_service"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/delete_company"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/delete_service"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/get_company"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/get_service"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/list_companies"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/list_services"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/update_company"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers/update_service"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/config"
	companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
	serviceRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/service"
	"github.com/m04kA/SMC-SellerService/internal/integrations/priceservice"
	"github.com/m04kA/SMC-SellerService/internal/integrations/userservice"
	companiesService "github.com/m04kA/SMC-SellerService/internal/service/companies"
	servicesService "github.com/m04kA/SMC-SellerService/internal/service/services"
	"github.com/m04kA/SMC-SellerService/pkg/dbmetrics"
	"github.com/m04kA/SMC-SellerService/pkg/logger"
	"github.com/m04kA/SMC-SellerService/pkg/metrics"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := logger.New(cfg.Logs.File)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting SMC-SellerService...")
	log.Info("Configuration loaded from config.toml")

	// Инициализируем метрики (если включены)
	var metricsCollector *metrics.Metrics
	var wrappedDB *dbmetrics.DB
	stopMetricsCh := make(chan struct{})

	if cfg.Metrics.Enabled {
		metricsCollector = metrics.New(cfg.Metrics.ServiceName)
		log.Info("Metrics enabled at %s", cfg.Metrics.Path)
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Настраиваем connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: %v", err)
	}
	log.Info("Successfully connected to database (host=%s, port=%d, db=%s)",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Инициализируем PriceService клиент
	priceClient := priceservice.NewClient(cfg.PriceService.BaseURL, log)
	log.Info("PriceService client initialized (base_url=%s)", cfg.PriceService.BaseURL)

	// Инициализируем UserService клиент
	userClient := userservice.NewClient(cfg.UserService.BaseURL, log)
	log.Info("UserService client initialized (base_url=%s)", cfg.UserService.BaseURL)

	// Инициализируем репозитории и сервисы (с метриками или без)
	var companySvc *companiesService.Service
	var serviceSvc *servicesService.Service

	if cfg.Metrics.Enabled {
		wrappedDB = dbmetrics.WrapWithDefault(db, metricsCollector, cfg.Metrics.ServiceName, stopMetricsCh)
		log.Info("Database metrics collection started")

		// Инициализируем репозитории с обёрткой метрик
		companyRepository := companyRepo.NewRepository(wrappedDB)
		serviceRepository := serviceRepo.NewRepository(wrappedDB)

		companySvc = companiesService.NewService(companyRepository, userClient)
		serviceSvc = servicesService.NewService(serviceRepository, companyRepository, priceClient)
	} else {
		// Инициализируем репозитории без метрик
		companyRepository := companyRepo.NewRepository(db)
		serviceRepository := serviceRepo.NewRepository(db)

		companySvc = companiesService.NewService(companyRepository, userClient)
		serviceSvc = servicesService.NewService(serviceRepository, companyRepository, priceClient)
	}

	// Инициализируем handlers для компаний
	createCompanyHandler := create_company.NewHandler(companySvc, log)
	getCompanyHandler := get_company.NewHandler(companySvc, log)
	listCompaniesHandler := list_companies.NewHandler(companySvc, log)
	updateCompanyHandler := update_company.NewHandler(companySvc, log)
	deleteCompanyHandler := delete_company.NewHandler(companySvc, log)

	// Инициализируем handlers для услуг
	createServiceHandler := create_service.NewHandler(serviceSvc, log)
	getServiceHandler := get_service.NewHandler(serviceSvc, log)
	listServicesHandler := list_services.NewHandler(serviceSvc, log)
	updateServiceHandler := update_service.NewHandler(serviceSvc, log)
	deleteServiceHandler := delete_service.NewHandler(serviceSvc, log)

	// Настраиваем роутер
	r := mux.NewRouter()

	// Добавляем metrics middleware (если метрики включены)
	if cfg.Metrics.Enabled {
		r.Use(middleware.MetricsMiddleware(metricsCollector, cfg.Metrics.ServiceName))
		log.Info("HTTP metrics middleware enabled")
	}

	// Metrics endpoint (публичный, без аутентификации)
	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler()).Methods(http.MethodGet)
		log.Info("Prometheus metrics endpoint exposed at %s", cfg.Metrics.Path)
	}

	// API prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public routes (с опциональной аутентификацией для получения дополнительных данных)
	public := api.PathPrefix("").Subrouter()
	public.Use(middleware.OptionalAuth)

	// Public routes для компаний
	public.HandleFunc("/companies", listCompaniesHandler.Handle).Methods(http.MethodGet)
	public.HandleFunc("/companies/{id}", getCompanyHandler.Handle).Methods(http.MethodGet)

	// Public routes для услуг
	public.HandleFunc("/companies/{company_id}/services", listServicesHandler.Handle).Methods(http.MethodGet)
	public.HandleFunc("/companies/{company_id}/services/{service_id}", getServiceHandler.Handle).Methods(http.MethodGet)

	// Protected routes (требуют X-User-ID и X-User-Role)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth)

	// Protected routes для компаний
	protected.HandleFunc("/companies", createCompanyHandler.Handle).Methods(http.MethodPost)
	protected.HandleFunc("/companies/{id}", updateCompanyHandler.Handle).Methods(http.MethodPut)
	protected.HandleFunc("/companies/{id}", deleteCompanyHandler.Handle).Methods(http.MethodDelete)

	// Protected routes для услуг
	protected.HandleFunc("/companies/{company_id}/services", createServiceHandler.Handle).Methods(http.MethodPost)
	protected.HandleFunc("/companies/{company_id}/services/{service_id}", updateServiceHandler.Handle).Methods(http.MethodPut)
	protected.HandleFunc("/companies/{company_id}/services/{service_id}", deleteServiceHandler.Handle).Methods(http.MethodDelete)

	// Создаем HTTP сервер
	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Останавливаем сбор метрик connection pool
	if cfg.Metrics.Enabled {
		close(stopMetricsCh)
		log.Info("Metrics collection stopped")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}
