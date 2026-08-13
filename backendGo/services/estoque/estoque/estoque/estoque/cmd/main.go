package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/estoque/internal/broker"
	"github.com/korp-teste/backendGo/services/estoque/internal/config"
	"github.com/korp-teste/backendGo/services/estoque/internal/database"
	"github.com/korp-teste/backendGo/services/estoque/internal/handlers"
	"github.com/korp-teste/backendGo/services/estoque/internal/health"
	"github.com/korp-teste/backendGo/services/estoque/internal/logging"
	"github.com/korp-teste/backendGo/services/estoque/internal/messaging"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/repositories"
	"github.com/korp-teste/backendGo/services/estoque/internal/routes"
	"github.com/korp-teste/backendGo/services/estoque/internal/services"
)

func main() {
	cfg := config.Load()
	logger := logging.New("estoque", logging.Version)
	logger.SetLevel(logging.ParseLevel(cfg.LogLevel))
	logging.SetDefault(logger)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("falha ao conectar no banco: %v", err)
	}

	if err := db.AutoMigrate(&models.Produto{}, &models.PrecoProduto{}, &models.MovimentoEstoque{}); err != nil {
		log.Fatalf("falha ao executar migrações: %v", err)
	}

	if err := database.ApplyMovimentoTrigger(db); err != nil {
		log.Fatalf("falha ao criar trigger de estoque: %v", err)
	}

	if err := database.ApplyMovimentoConstraints(db); err != nil {
		log.Fatalf("falha ao criar constraints de idempotência: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("falha ao obter handle do banco: %v", err)
	}

	b, err := broker.Connect(cfg.RabbitMQURL, func(ch *amqp.Channel) error {
		return messaging.Declare(ch)
	})
	if err != nil {
		log.Fatalf("falha ao conectar no RabbitMQ: %v", err)
	}
	defer b.Close()

	if err := messaging.Declare(b.Channel); err != nil {
		log.Fatalf("falha ao declarar topologia de mensageria: %v", err)
	}

	getChannel := func() *amqp.Channel {
		_ = b.IsHealthy()
		return b.ChannelSafe()
	}

	produtoRepo := repositories.NewProdutoRepository(db)
	produtoService := services.NewProdutoService(produtoRepo)
	produtoHandler := handlers.NewProdutoHandler(produtoService)

	movimentoRepo := repositories.NewMovimentoRepository(db)
	movimentoService := services.NewMovimentoService(movimentoRepo, produtoRepo)
	movimentoHandler := handlers.NewMovimentoHandler(movimentoService)

	publicadorResultado := services.NewResultadoBaixaPublisher(getChannel)
	consumidorBaixa := services.NewBaixaConsumer(getChannel, movimentoService, publicadorResultado)
	if err := consumidorBaixa.Start(); err != nil {
		log.Fatalf("falha ao iniciar consumidor de baixa: %v", err)
	}
	defer consumidorBaixa.Close()

	healthChecker := health.NewChecker("estoque", logging.Version, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return sqlDB.PingContext(ctx)
	}, b.IsHealthy)

	r := gin.Default()
	routes.Setup(r, routes.Handlers{
		Produto:   produtoHandler,
		Movimento: movimentoHandler,
		Health:    healthChecker,
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
