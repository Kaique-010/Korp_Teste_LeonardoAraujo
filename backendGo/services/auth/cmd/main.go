package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"korp-teste/auth/internal/config"
	"korp-teste/auth/internal/database"
	"korp-teste/auth/internal/handlers"
	"korp-teste/auth/internal/jwt"
	"korp-teste/auth/internal/repositories"
	"korp-teste/auth/internal/routes"
	"korp-teste/auth/internal/services"
)

const (
	SeedAdminNome  = "Administrador Korp"
	SeedAdminEmail = "admin@korp.local"
	SeedAdminSenha = "korp26"
)

func main() {
	ctx := context.Background()

	// ============================================================
	// CONFIGURAÇÃO
	// ============================================================

	cfg := config.Load()

	// Valida a configuração da chave JWT antes de iniciar o serviço.
	jwtSecret := []byte(cfg.JWTSecret)

	if err := jwt.ValidateJWTSecret(jwtSecret); err != nil {
		log.Fatal(err)
	}

	// ============================================================
	// BANCO DE DADOS
	// ============================================================

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal(err)
	}

	// ============================================================
	// REPOSITORY
	// ============================================================

	usuarioRepository := repositories.NewUsuarioRepository(db)

	// ============================================================
	// SERVICES
	// ============================================================

	usuarioService := services.NewUsuarioService(
		usuarioRepository,
	)

	jwtService := jwt.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTExpiration,
	)

	// ============================================================
	// SEED — Usuário administrador padrão (apenas se tabela vazia)
	// ============================================================

	criado, admin, err := usuarioService.CriarSeVazio(
		ctx,
		SeedAdminNome,
		SeedAdminEmail,
		SeedAdminSenha,
	)
	if err != nil {
		log.Fatalf("falha no seed do admin: %v", err)
	}
	if criado {
		log.Printf(
			"[SEED] Usuário admin criado: id=%d nome=%q email=%q senha=%q",
			admin.ID, admin.Nome, admin.Email, SeedAdminSenha,
		)
	} else {
		log.Printf("[SEED] Tabela de usuários já possui registros — nada feito.")
	}

	// ============================================================
	// HANDLERS
	// ============================================================

	usuarioHandler := handlers.NewUsuarioHandler(
		jwtService,
		usuarioService,
	)

	// ============================================================
	// CORS
	// ============================================================

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{
		"http://localhost:4200",
	}

	corsConfig.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	}

	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
	}

	corsConfig.ExposeHeaders = []string{
		"Content-Length",
	}

	corsConfig.AllowCredentials = false
	corsConfig.MaxAge = 12 * time.Hour

	// ============================================================
	// HTTP / GIN
	// ============================================================

	router := gin.Default()

	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "auth",
		})
	})

	// Rotas da aplicação
	routes.Setup(router, usuarioHandler)

	// ============================================================
	// START
	// ============================================================

	log.Printf(
		"Auth Service iniciado na porta %s",
		cfg.Port,
	)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
