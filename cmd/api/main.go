package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"financas-sah-api/internal/auth"
	"financas-sah-api/internal/config"
	"financas-sah-api/internal/database"
	"financas-sah-api/internal/handlers"
	"financas-sah-api/internal/middleware"
	"financas-sah-api/internal/repository"
	"financas-sah-api/internal/services"
	"financas-sah-api/migrations"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	log.Println("🎀 Inicializando Finanças Sah API (Hello Kitty Edition)...")

	// 1. Carregar Configurações
	cfg := config.LoadConfig()

	// 2. Conectar ao Banco de Dados PostgreSQL
	db, err := database.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Erro fatal: Banco de dados inacessível: %v", err)
	}
	defer db.Close()

	// 3. Executar Migrações Automáticas
	if err := db.RunMigrations(migrations.FS); err != nil {
		log.Fatalf("❌ Erro ao rodar migrações: %v", err)
	}

	// 4. Inicializar Gerenciadores e Repositórios
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(db)
	cycleRepo := repository.NewCycleRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	dashRepo := repository.NewDashboardRepository(db, cycleRepo, txRepo)

	// 5. Inicializar Serviços
	authService := services.NewAuthService(userRepo, cycleRepo, jwtManager)
	cycleService := services.NewCycleService(cycleRepo)
	catService := services.NewCategoryService(catRepo)
	txService := services.NewTransactionService(txRepo, catRepo, cycleRepo)
	dashService := services.NewDashboardService(dashRepo)

	// 6. Inicializar Handlers
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(authService)
	cycleHandler := handlers.NewCycleHandler(cycleService)
	catHandler := handlers.NewCategoryHandler(catService)
	txHandler := handlers.NewTransactionHandler(txService)
	dashHandler := handlers.NewDashboardHandler(dashService)

	// 7. Configurar Roteador Chi
	r := chi.NewRouter()

	// Middlewares Globais
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.EnableCORS)
	r.Use(middleware.RequestLogger)

	// Rotas da API v1
	r.Route("/api/v1", func(api chi.Router) {
		// Health & Boas-vindas
		api.Get("/health", healthHandler.CheckHealth)
		api.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.RespondSuccess(w, http.StatusOK, "Bem-vinda ao Finanças Sah API 🎀", map[string]string{
				"version": "1.0.0",
				"theme":   "Hello Kitty Edition 🎀",
			})
		})

		// Rotas Públicas de Autenticação
		api.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/register", authHandler.Register)
			authRouter.Post("/login", authHandler.Login)
		})

		// Rotas Protegidas por JWT
		api.Group(func(protected chi.Router) {
			protected.Use(middleware.AuthMiddleware(jwtManager))

			// Dados da Conta & Preferências
			protected.Get("/auth/me", authHandler.GetMe)
			protected.Put("/auth/settings", authHandler.UpdateSettings)

			// Ciclos Financeiros & Fechamento de Mês
			protected.Route("/cycles", func(cycleRouter chi.Router) {
				cycleRouter.Get("/current", cycleHandler.GetCurrentCycle)
				cycleRouter.Get("/", cycleHandler.ListCycles)
				cycleRouter.Get("/{id}", cycleHandler.GetCycleByID)
				cycleRouter.Post("/close", cycleHandler.CloseCycle)
			})

			// Categorias / Tópicos Customizáveis
			protected.Route("/categories", func(catRouter chi.Router) {
				catRouter.Get("/", catHandler.List)
				catRouter.Post("/", catHandler.Create)
				catRouter.Put("/{id}", catHandler.Update)
				catRouter.Delete("/{id}", catHandler.Delete)
			})

			// Transações / Lançamentos Financeiros
			protected.Route("/transactions", func(txRouter chi.Router) {
				txRouter.Get("/", txHandler.List)
				txRouter.Post("/", txHandler.Create)
				txRouter.Get("/{id}", txHandler.GetByID)
				txRouter.Put("/{id}", txHandler.Update)
				txRouter.Delete("/{id}", txHandler.Delete)
			})

			// Dashboard & Analytics
			protected.Route("/dashboard", func(dashRouter chi.Router) {
				dashRouter.Get("/summary", dashHandler.GetSummary)
			})
		})
	})

	// 8. Iniciar Servidor HTTP com Graceful Shutdown
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Servidor rodando na porta %s (http://localhost:%s)", cfg.Port, cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erro fatal no servidor: %v", err)
		}
	}()

	// Aguardar sinal de encerramento do SO
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🌸 Encerrando Finanças Sah API de forma segura...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Falha ao encerrar servidor: %v", err)
	}

	log.Println("✨ Finanças Sah API finalizada com sucesso. Até logo!")
}
