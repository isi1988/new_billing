package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"

	"new-billing/frontend"
	"new-billing/internal/api"
	"new-billing/internal/config"
	"new-billing/internal/database"
	"new-billing/internal/hostname"
	"new-billing/internal/models"
	"new-billing/internal/service"
	"new-billing/internal/telegram"

	"github.com/gorilla/mux"

	// --- SWAGGER ИМПОРТЫ ---
	httpSwagger "github.com/swaggo/http-swagger"
	_ "new-billing/docs" // Важно! Импортируем сгенерированную документацию
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("FATAL: Failed to load config: %v", err)
	}

	db := database.Connect(cfg.Database)
	database.Migrate(db)

	// Заполняем базовые данные при первом запуске
	database.SeedBasicData(db)

	// Трафик теперь будет обрабатываться реально через функцию ProcessIPTraffic
	// в зависимости от ваших NetFlow данных

	flowService := service.NewFlowService(db, &cfg.Nfcapd)
	go flowService.StartProcessing()
	
	// Запускаем воркер для резолвинга IP адресов
	hostnameWorker := hostname.NewHostnameWorker(db)
	go hostnameWorker.StartWorker()

	r := mux.NewRouter()

	// --- Инициализация сервисов ---
	telegramService := telegram.NewTelegramService(&cfg.Telegram)

	// --- Инициализация обработчиков ---
	authHandler := &api.AuthHandler{DB: db, Cfg: cfg}
	billingHandler := &api.BillingHandler{DB: db, TelegramService: telegramService, HostnameWorker: hostnameWorker}
	netflowHandler := api.NewAPIHandler(db)

	// --- SWAGGER ROUTE ---
	// Этот маршрут должен идти до определения /api, чтобы не было конфликтов
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// --- Публичные роуты ---
	// Мы не хотим документировать login в общем списке, так как это отдельная операция
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// --- Защищенные роуты API ---
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(api.AuthMiddleware(cfg))
	// Роуты для NetFlow (доступны всем аутентифицированным пользователям)
	apiRouter.HandleFunc("/flows/search", netflowHandler.SearchFlows).Methods("GET")
	apiRouter.HandleFunc("/flows/aggregate", netflowHandler.AggregateFlows).Methods("GET")

	// --- Роуты для Менеджеров и Админов ---
	managerRouter := apiRouter.PathPrefix("").Subrouter()
	managerRouter.Use(api.RoleRequired(models.ManagerRole, models.AdminRole))

	// CRUD для Тарифов
	managerRouter.HandleFunc("/tariffs", billingHandler.GetTariffs).Methods("GET")
	managerRouter.HandleFunc("/tariffs", billingHandler.CreateTariff).Methods("POST")
	managerRouter.HandleFunc("/tariffs/{id:[0-9]+}", billingHandler.GetTariffByID).Methods("GET")
	managerRouter.HandleFunc("/tariffs/{id:[0-9]+}", billingHandler.UpdateTariff).Methods("PUT")
	managerRouter.HandleFunc("/tariffs/{id:[0-9]+}", billingHandler.DeleteTariff).Methods("DELETE")

	// CRUD для Клиентов
	managerRouter.HandleFunc("/clients", billingHandler.GetClients).Methods("GET")
	managerRouter.HandleFunc("/clients", billingHandler.CreateClient).Methods("POST")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}", billingHandler.GetClientByID).Methods("GET")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}", billingHandler.UpdateClient).Methods("PUT")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}", billingHandler.DeleteClient).Methods("DELETE")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}/block", billingHandler.BlockClient).Methods("POST")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}/unblock", billingHandler.UnblockClient).Methods("POST")

	// Договоры по клиенту
	managerRouter.HandleFunc("/clients/{client_id:[0-9]+}/contracts", billingHandler.GetContractsByClient).Methods("GET")
	managerRouter.HandleFunc("/clients/{id:[0-9]+}/ips", billingHandler.GetClientIPs).Methods("GET")

	// CRUD для Оборудования
	managerRouter.HandleFunc("/equipment", billingHandler.GetAllEquipment).Methods("GET")
	managerRouter.HandleFunc("/equipment", billingHandler.CreateEquipment).Methods("POST")
	managerRouter.HandleFunc("/equipment/{id:[0-9]+}", billingHandler.GetEquipmentByID).Methods("GET")
	managerRouter.HandleFunc("/equipment/{id:[0-9]+}", billingHandler.UpdateEquipment).Methods("PUT")
	managerRouter.HandleFunc("/equipment/{id:[0-9]+}", billingHandler.DeleteEquipment).Methods("DELETE")

	// CRUD для Договоров
	managerRouter.HandleFunc("/contracts", billingHandler.GetContracts).Methods("GET")
	managerRouter.HandleFunc("/contracts", billingHandler.CreateContract).Methods("POST")
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}", billingHandler.GetContractByID).Methods("GET")
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}", billingHandler.UpdateContract).Methods("PUT")
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}", billingHandler.DeleteContract).Methods("DELETE")
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}/block", billingHandler.BlockContract).Methods("POST")
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}/unblock", billingHandler.UnblockContract).Methods("POST")

	// CRUD для Подключений
	managerRouter.HandleFunc("/connections", billingHandler.GetConnections).Methods("GET")
	managerRouter.HandleFunc("/connections", billingHandler.CreateConnection).Methods("POST")
	managerRouter.HandleFunc("/connections/{id:[0-9]+}", billingHandler.GetConnectionByID).Methods("GET")
	managerRouter.HandleFunc("/connections/{id:[0-9]+}", billingHandler.UpdateConnection).Methods("PUT")
	managerRouter.HandleFunc("/connections/{id:[0-9]+}", billingHandler.DeleteConnection).Methods("DELETE")
	managerRouter.HandleFunc("/connections/{id:[0-9]+}/block", billingHandler.BlockConnection).Methods("POST")
	managerRouter.HandleFunc("/connections/{id:[0-9]+}/unblock", billingHandler.UnblockConnection).Methods("POST")

	// Подключения по договору
	managerRouter.HandleFunc("/contracts/{contract_id:[0-9]+}/connections", billingHandler.GetConnectionsByContract).Methods("GET")

	// Дашборд трафика
	managerRouter.HandleFunc("/traffic", billingHandler.GetTrafficData).Methods("GET")
	managerRouter.HandleFunc("/traffic/stats", billingHandler.GetTrafficStats).Methods("GET")
	managerRouter.HandleFunc("/traffic/export", billingHandler.ExportTrafficCSV).Methods("GET")

	// Статистика по договорам
	managerRouter.HandleFunc("/contracts/{id:[0-9]+}/stats", billingHandler.GetContractStats).Methods("GET")
	// Статистика по подключениям
	managerRouter.HandleFunc("/connections/{id:[0-9]+}/stats", billingHandler.GetConnectionStats).Methods("GET")
	
	// Системная информация
	managerRouter.HandleFunc("/system/info", billingHandler.GetSystemInfo).Methods("GET")
	
	// Информация об IP адресах
	managerRouter.HandleFunc("/ip/{ip}/info", billingHandler.GetIPInfo).Methods("GET")

	// CRUD для Доработок/Issues
	managerRouter.HandleFunc("/issues", billingHandler.GetIssues).Methods("GET")
	managerRouter.HandleFunc("/issues", billingHandler.CreateIssue).Methods("POST")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}", billingHandler.GetIssueByID).Methods("GET")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}", billingHandler.UpdateIssue).Methods("PUT")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}", billingHandler.DeleteIssue).Methods("DELETE")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}/resolve", billingHandler.ResolveIssue).Methods("POST")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}/unresolve", billingHandler.UnresolveIssue).Methods("POST")
	managerRouter.HandleFunc("/issues/{id:[0-9]+}/history", billingHandler.GetIssueHistory).Methods("GET")

	// --- Роуты только для Админов ---
	adminRouter := apiRouter.PathPrefix("").Subrouter()
	adminRouter.Use(api.RoleRequired(models.AdminRole))

	// CRUD для Пользователей
	adminRouter.HandleFunc("/users", billingHandler.GetUsers).Methods("GET")
	adminRouter.HandleFunc("/users", billingHandler.CreateUser).Methods("POST")
	adminRouter.HandleFunc("/users/{id:[0-9]+}", billingHandler.GetUserByID).Methods("GET")
	adminRouter.HandleFunc("/users/{id:[0-9]+}", billingHandler.UpdateUser).Methods("PUT")
	adminRouter.HandleFunc("/users/{id:[0-9]+}", billingHandler.DeleteUser).Methods("DELETE")

	// --- Раздача фронтенда ---
	staticContent, _ := fs.Sub(frontend.StaticFiles, "dist")
	r.PathPrefix("/").Handler(http.FileServer(http.FS(spaFileSystem{staticContent})))

	log.Printf("Starting server on port %s", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(cfg.Server.Port, r))
}

// spaFileSystem - без изменений
type spaFileSystem struct{ root fs.FS }

func (sfs spaFileSystem) Open(name string) (fs.File, error) {
	file, err := sfs.root.Open(name)
	if os.IsNotExist(err) {
		return sfs.root.Open("index.html")
	}
	return file, err
}
