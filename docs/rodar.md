# Como rodar o projeto

Guia para subir a infraestrutura (banco + mensageria) e os microsserviços localmente.

## Pré-requisitos

- **Docker Desktop** (com Docker Compose) — para Postgres e RabbitMQ
- **Go 1.26+** — para os microsserviços
- **Node.js 22+ e npm** — para o frontend Angular (Sprint 10)

## Resumo rápido (sequência completa de subida)

```bash
# 1. infra (3 containers)
docker compose up -d
docker compose ps              # esperar todos "healthy"

# 2. microsserviços (2 terminais separados)
cd backendGo/services/estoque      ; go run ./cmd
cd backendGo/services/faturamento  ; go run ./cmd

# 3. health
curl http://localhost:8081/health
curl http://localhost:8082/health

# 4. frontend (opcional)
cd frontend ; npm install       # apenas na 1a vez
cd frontend ; npm start         # dev proxy em 4200
```
