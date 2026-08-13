# Korp Teste — Sistema de Emissão de Notas Fiscais

Sistema de emissão de notas fiscais em arquitetura de microsserviços.

- **Backend:** Go, Gin, GORM, PostgreSQL, RabbitMQ
- **Frontend:** Angular (em desenvolvimento após o backend)
- **Infra:** Docker Compose

## Microsserviços

| Serviço      | Pasta                          | Domínio                          |
| ------------ | ------------------------------ | -------------------------------- |
| Estoque      | `backendGo/services/estoque`   | Produto, MovimentoEstoque        |
| Faturamento  | `backendGo/services/faturamento` | NotaFiscal, Item, Evento         |

## Como rodar

```bash
docker compose up -d
```

Cada microsserviço é um módulo Go independente:

```bash
cd backendGo/services/estoque
go run ./cmd
```

## Documentação

- `docs/rodar.md` — como subir bancos, microsserviços e frontend
- `docs/arquitetura.md` — divisão lógica da SDD e estrutura do projeto
- `docs/planejamento.md` — sprints em checklist

## Endpoints

### Estoque

```http
POST /produtos
GET  /produtos
GET  /produtos/:id
PUT  /produtos/:id
DELETE /produtos/:id
POST /estoque/movimentos
```

### Faturamento

```http
POST /notas
GET  /notas
GET  /notas/:id
POST /notas/:id/itens
DELETE /notas/:id/itens/:item_id
POST /notas/:id/imprimir
```
