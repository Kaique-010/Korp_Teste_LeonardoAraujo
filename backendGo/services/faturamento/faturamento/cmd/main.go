package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/broker"
	"github.com/korp-teste/backendGo/services/faturamento/internal/config"
	"github.com/korp-teste/backendGo/services/faturamento/internal/database"
	"github.com/korp-teste/backendGo/services/faturamento/internal/handlers"
	"github.com/korp-teste/backendGo/services/faturamento/internal/health"
	"github.com/korp-teste/backendGo/services/faturamento/internal/logging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/repositories"
	"github.com/korp-teste/backendGo/services/faturamento/internal/routes"
	"github.com/korp-teste/backendGo/services/faturamento/internal/services"
)

func main() {
	cfg := config.Load()
	logger := logging.New("faturamento", logging.Version)
	logger.SetLevel(logging.ParseLevel(cfg.LogLevel))
	logging.SetDefault(logger)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("falha ao conectar no banco: %v", err)
	}

	if err := db.AutoMigrate(&models.Cliente{}, &models.NotaFiscal{}, &models.NotaFiscalItem{}, &models.NotaFiscalEvento{}); err != nil {
		log.Fatalf("falha ao executar migrações: %v", err)
	}

	if err := database.ApplyNotaFiscalSequence(db); err != nil {
		log.Fatalf("falha ao criar sequência de numeração: %v", err)
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

	clienteRepo := repositories.NewClienteRepository(db)
	clienteService := services.NewClienteService(clienteRepo)
	clienteHandler := handlers.NewClienteHandler(clienteService)

	notaRepo := repositories.NewNotaFiscalRepository(db)
	eventoRepo := repositories.NewNotaFiscalEventoRepository(db)
	estoqueCliente := services.NewEstoqueCliente(cfg.EstoqueBaseURL)
	eventoService := services.NewNotaFiscalEventoService(eventoRepo)
	publicador := services.NewBaixaPublisher(getChannel)
	notaService := services.NewNotaFiscalService(notaRepo, estoqueCliente, eventoService, publicador, clienteRepo)
	notaHandler := handlers.NewNotaFiscalHandler(notaService)
	eventoHandler := handlers.NewNotaFiscalEventoHandler(notaService, eventoService)

	consumidorResultado := services.NewResultadoBaixaConsumer(getChannel, notaService)
	if err := consumidorResultado.Start(); err != nil {
		log.Fatalf("falha ao iniciar consumidor de resultado: %v", err)
	}
	defer consumidorResultado.Close()

	healthChecker := health.NewChecker("faturamento", logging.Version, func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return sqlDB.PingContext(ctx)
	}, b.IsHealthy)

	r := gin.Default()
	routes.Setup(r, routes.Handlers{
		NotaFiscal: notaHandler,
		Evento:     eventoHandler,
		Cliente:    clienteHandler,
		Health:     healthChecker,
	})

	if err := r.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
