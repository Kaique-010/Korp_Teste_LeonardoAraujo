# Korp Teste — Sistema de Emissão de Notas Fiscais

Teste técnico: **sistema de emissão de notas fiscais** com **2 microsserviços Go**
(Estoque + Faturamento), **frontend Angular 19 standalone + Material**, mensageria RabbitMQ,
auditoria JSONB, idempotência, logs estruturados, e **preços em 2 modalidades
(à vista / a prazo / ambos)** com histórico SCD2.

- **Backend 1 (Estoque)**: Go, Gin, GORM, PostgreSQL, SQLite em memória (testes),
  RabbitMQ consumer/publisher.
- **Backend 2 (Faturamento)**: Go, Gin, GORM, PostgreSQL, RabbitMQ publisher/consumer,
  integração HTTP com Estoque, auditoria JSONB `nota_fiscal_eventos`.
- **Frontend**: Angular 19 standalone + Angular Material, tema Light/Dark com persistência
  em `localStorage`, header responsivo hambúrguer (<820px), CRUDs Produtos/Clientes/Notas,
  impressão assíncrona 202.
- **Infra**: Docker Compose (Postgres 17 + RabbitMQ 3.13 + PgAdmin).
- **Padrão de commits**: Conventional Commits em português (40 commits + pasta
  `commits/` com 1 stub por commit como prova do passo-a-passo).

---

## 1. Estrutura do Monorepo

```
Korp_Teste_LeonardoAraujo/
├── backendGo/
│   └── services/
│       ├── estoque/        ⭐ Microsserviço 1 — Produto + Preço + Movimento
│       │   ├── cmd/main.go
│       │   └── internal/
│       │       ├── config/           # env vars (PORT, DB, RABBIT, LOG_LEVEL)
│       │       ├── database/         # GORM + migrations + UNIQUE idempotency
│       │       ├── health/           # /health real (DB + Rabbit conectados)
│       │       ├── logging/          # logs JSON estruturados (middleware)
│       │       ├── broker/           # Rabbit reconexão automática + topologia
│       │       ├── messaging/        # contratos + filas korp.baixa, korp.resultado + DLX
│       │       ├── models/           # Produto, PrecoProduto (SCD2), MovimentoEstoque
│       │       ├── repositories/
│       │       ├── services/         # Produto + Movimento + consumer Baixa + publisher Resultado
│       │       ├── handlers/         # ProdutoHandler, MovimentoHandler, ErrorPadrao
│       │       └── routes/routes.go
│       │
│       └── faturamento/    ⭐ Microsserviço 2 — NotaFiscal + Item + Evento + Cliente
│           ├── cmd/main.go
│           └── internal/
│               ├── config/database/health/logging/broker/messaging   # idem estoque
│               ├── models/           # NotaFiscal (A/BERTA, FECHADA), Item, Evento (JSONB), Cliente
│               ├── repositories/     # NF, Item, Evento, Cliente (interfaces + GORM impl)
│               ├── services/         # NF + publisher BaixaSolicitada + consumer Resultado
│               ├── handlers/         # NotaFiscalHandler, ClienteHandler, EventoHandler, ErrorPadrao
│               └── routes/routes.go
│
├── frontend/                  ⭐ Angular 19 standalone + Material
│   ├── angular.json  (budgets ajustados, build OK)
│   ├── proxy.conf.json         4200 → 8081/8082 (/api/estoque, /api/faturamento)
│   └── src/
│       ├── app/
│       │   ├── components/     # app-header (hambúrguer), theme-toggle,
│       │   │                   # produto-form-dialog (toggle preço),
│       │   │                   # cliente-form-dialog, nova-nota-dialog,
│       │   │                   # item-form-dialog (toggle vista/prazo/manual)
│       │   │                   # confirm-dialog
│       │   ├── pages/          # Home (cards), Produtos, Clientes,
│       │   │                   # Notas (listagem), NotaDetalhe (itens + eventos JSONB)
│       │   ├── services/       # ProdutoService, ClienteService, NotaFiscalService,
│       │   │                   # ThemeService (signal + Renderer2 + localStorage)
│       │   ├── models/         # Produto (+preco_vista, preco_prazo),
│       │   │                   # NotaFiscal (+cliente_id, status),
│       │   │                   # NotaFiscalItem, NotaFiscalEvento, Cliente
│       │   ├── app.ts / app.html / app.scss
│       │   ├── app.routes.ts
│       │   └── app.config.ts   # provideHttpClient + provideAnimations
│       ├── index.html          lang="pt-BR", <meta charset="UTF-8">
│       └── styles.scss         tema Light/Dark via CSS vars
│
├── commits/                    ⭐ Prova passo-a-passo (1 stub por commit)
│   ├── commit001.stamp  …  commit040.stamp
│
├── docs/                       Documentação oficial
│   ├── arquitetura.md    (Partes 1..12)
│   ├── planejamento.md   (Sprints 1..11)
│   ├── commits.md        (mapa 40 commits × sprint × SDD § × Parte)
│   ├── rodar.md          (sequência completa de subida)
│   ├── guia-sprint7-mensageria.md
│   ├── terminal-historia-985-1010.md
│   ├── reescrever-historico.ps1     # Windows PowerShell 5
│   └── reescrever-historico.sh      # Linux / macOS / Git Bash
│
├── sdd_teste.md                SDD oficial §1..§26 (especificação do teste)
├── .env.example                PORT / DB / RABBITMQ_URL / LOG_LEVEL / SERVICE_NAME / VERSION
├── docker-compose.yml
└── README.md (este arquivo)
```

---

## 2. Como rodar (sequência completa)

Passo-a-passo completo em [docs/rodar.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/docs/rodar.md).
Resumo:

```bash
# 1. Infra (3 containers: Postgres + RabbitMQ + PgAdmin)
docker compose up -d

# 2. Estoque (porta 8081)
cd backendGo/services/estoque      ; go run ./cmd

# 3. Faturamento (porta 8082) — OUTRO terminal
cd backendGo/services/faturamento  ; go run ./cmd

# 4. Health checks (esperado 200 ambos)
curl http://localhost:8081/health
curl http://localhost:8082/health

# 5. Frontend (opcional, porta 4200 — dev proxy → 8081/8082)
cd frontend ; npm install          # só 1ª vez
cd frontend ; npm start
```

Variáveis de ambiente: copie `.env.example` para `.env` (valores padrão já
funcionam com o `docker-compose.yml` local).

---

## 3. Arquitetura (3 camadas padrão)

| Camada       | Responsabilidade                                                                   |
| ------------ | ---------------------------------------------------------------------------------- |
| **Handlers** | Gin: valida entrada, monta status HTTP, formata erro padrão. **Zero regra.**       |
| **Services** | Regras de negócio (transações GORM, idempotência, integrações HTTP/Rabbit, retry). |
| **Repos**    | Acesso a dados via GORM + interfaces (mockável em testes SQLite memória).          |

---

## 4. Endpoints (HTTP REST + JSON)

### 4.1. Estoque (porta `8081`)

```http
# Produtos (código automático PROD-XXXXXX se não enviar)
GET    /produtos
POST   /produtos             { codigo?, descricao, saldo, preco_vista?, preco_prazo? }
GET    /produtos/:id
PUT    /produtos/:id         { descricao, saldo, preco_vista?, preco_prazo? }
DELETE /produtos/:id         # 409 se houver movimentos vinculados

# Preços do Produto — histórico SCD2 (vigência, fim_em null = vigente)
GET    /produtos/:id/precos                # histórico completo (do mais novo ao mais antigo)
POST   /produtos/:id/precos                # nova versão vigente (fecha a anterior com fim_em = agora)
       { preco_vista, preco_prazo, vigente_em? }

# Movimentos (ENTRADA / SAÍDA)
POST   /estoque/movimentos    { produto_id, tipo:ENTRADA|SAIDA, quantidade, origem,
                                 referencia?, idempotency_key? }
                              # 409 UNIQUE uq_movimentos_idempotency se idempotency_key duplicada
                              # 409 CONCORRÊNCIA se saldo=0 no SAÍDA (trigger FOR UPDATE)

# Observabilidade
GET    /health                # 200 só se DB e Rabbit conectados (não só "pong")
GET    /                       # HATEOAS: links para todos endpoints
```

### 4.2. Faturamento (porta `8082`)

```http
# Clientes (FK em NotaFiscal.cliente_id)
GET    /clientes
POST   /clientes             { nome }
GET    /clientes/:id
PUT    /clientes/:id         { nome }
DELETE /clientes/:id         # 409 se cliente tiver notas fiscais vinculadas

# Notas Fiscais (ABERTA / FECHADA; número sequencial automático por ano)
GET    /notas
POST   /notas                { cliente_id? }
GET    /notas/:id            # retorna itens + resumo totais
GET    /notas/:id/eventos    # auditoria JSONB: ESTOQUE_BAIXADO, FALHA_ESTOQUE, NOTA_FECHADA...
POST   /notas/:id/itens      { produto_id, descricao?, quantidade,
                                 preco_unitario?, preco_vista?, preco_prazo?, desconto? }
                               # preenche automaticamente se existir preço vigente no produto
DELETE /notas/:id/itens/:item_id
POST   /notas/:id/imprimir   # 202 Accepted → publica BaixaSolicitada na korp.baixa
                               # → Estoque consome, baixa, publica ResultadoBaixa
                               # → Faturamento consome e fecha nota (idempotente, só fecha 1x)

# Observabilidade
GET    /health
GET    /                       # HATEOAS: links para clientes + notas + imprimir
```

---

## 5. Preços do Produto (à vista / a prazo / ambos — SCD2)

É o item que você pediu que o README atualizasse com destaque:

1. **Armazenamento dual** — todo produto tem os 2 campos no model (`Produto`):
   - `preco_vista  DECIMAL(18,2) NULL`
   - `preco_prazo  DECIMAL(18,2) NULL`
2. **3 modalidades no formulário** (frontend `produto-form-dialog`):
   - **À Vista:** só `preco_vista` obrigatório, `preco_prazo` aceita nulo.
   - **A Prazo:** só `preco_prazo` obrigatório, `preco_vista` aceita nulo.
   - **Ambos:** os 2 campos obrigatórios.
3. **Histórico de preços SCD2** — tabela `preco_produtos`:
   - `produto_id`, `preco_vista`, `preco_prazo`, `vigente_em`, `fim_em` (NULL = vigente)
   - Endpoint `POST /produtos/:id/precos` **fecha a versão vigente** (seta `fim_em = NOW()`)
     e cria uma nova linha. Nunca atualiza a linha vigente no lugar (garante trilha de
     auditoria fiscal).
4. **Auto-preenchimento no Item da Nota** (frontend `item-form-dialog`):
   - Toggle **Preço: `A VISTA` / `A PRAZO` / `MANUAL`**.
   - `A VISTA` → busca `preco_vista` vigente do produto e preenche `preco_unitario`.
   - `A PRAZO` → busca `preco_prazo` vigente do produto e preenche `preco_unitario`.
   - `MANUAL` → campo livre, usuário digita.
5. **Regra de negócio do backend:** se não enviar `preco_unitario` no `POST /notas/:id/itens`,
   o service busca a versão vigente e preenche (vista primeiro, fallback prazo).

---

## 6. Impressão (fluxo assíncrono completo — RabbitMQ, etapas Sprint 7)

| Etapa | O que acontece                                                                                      |
| :---: | --------------------------------------------------------------------------------------------------- |
|  1/7  | `POST /notas/:id/imprimir` recebe a requisição → **202 Accepted**                                   |
|  2/7  | Faturamento publica **`BaixaSolicitada`** na exchange `korp.direct` routing-key `korp.baixa`        |
|  3/7  | Estoque consome `korp.baixa` (prefetch 1, `autoAck=false`)                                          |
|  4/7  | Estoque executa a baixa (transação `FOR UPDATE`, saldo>0 → insere movimento)                        |
|  5/7  | Estoque publica **`ResultadoBaixa`** (APROVADO / NEGADO / INDISPONIVEL) em `korp.resultado`         |
|  6/7  | Faturamento consome `korp.resultado` + **fecha a nota idempotentemente** (só muda FECHAR se ABERTA) |
|  7/7  | Em caso de erro: retry 2s → 4s → 8s → 16s → 30s → DLX → DLQ auditoria `korp.*.dlq`                  |

**Idempotência (Sprint 8):** a mesma nota enviada 2x para imprimir → 2 redeliveries da
`ResultadoBaixa` → o service verifica: `if nota.Status == FECHADA => Ack sem side effect`.

---

## 7. Observabilidade (Sprint 9)

- **Logs JSON estruturados** em ambos backends (request_id, status, latência_ms, path, method,
  error se existir) via middleware Gin antes do handler.
- **Formato de erro padronizado** `{"error":{"code":"NOTA_FECHADA","message":"..."}}` em todos
  handlers.
- **Health checks reais** (não pong): checa `DB.Ping()` + estado da conexão RabbitMQ (canal
  aberto).
- **Env vars obrigatórias** em `.env.example`: `PORT / DB_HOST / DB_PORT / DB_NAME / DB_USER /
DB_PASS / RABBITMQ_URL / LOG_LEVEL (info|debug) / SERVICE_NAME / VERSION`.
- **Reconexão broker automática:** publisher/consumer tenta reconectar a cada 2s se a conexão
  do `amqp.Dial` cair (não deixa o processo morrer).

---

## 8. Frontend Angular 19 (Sprints 10 + 11 UX + UTF-8)

| Feature                                    | Onde                                                                                                                                                          |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Scaffold standalone + Material + proxy** | `app.config.ts` (provideHttpClient, provideAnimations); `proxy.conf.json` → `/api/estoque/*` porta 8081, `/api/faturamento/*` 8082.                           |
| **HomePage (cards)**                       | `pages/home` — 4 cards: Produtos / Clientes / Notas Abertas / Eventos de auditoria.                                                                           |
| **Header responsivo**                      | `components/app-header` — hambúrguer em telas <820px. Nav: Home / Produtos / Clientes / Notas + ThemeToggle.                                                  |
| **Tema Light/Dark**                        | `services/theme.service.ts` (Signal + `Renderer2` + `localStorage`). Troca atributo `data-theme` em `<body>` + CSS vars. Toggle em `components/theme-toggle`. |
| **CRUD Produtos**                          | `pages/produtos` + `components/produto-form-dialog`: **toggle tipo preço (Vista / Prazo / Ambos)** + campo `Código PROD-XXXXXX` automático.                   |
| **CRUD Clientes**                          | `pages/clientes` + `components/cliente-form-dialog`: apenas `nome`. 409 exclusão se notas vinculadas.                                                         |
| **CRUD Notas**                             | `pages/notas` (listagem com número, cliente, status); `pages/nota-detalhe` com 3 abas: Itens / Eventos JSONB / Imprimir 202.                                  |
| **Item Nota auto-preenche**                | `components/item-form-dialog`: **toggle (Vista / Prazo / Manual)** — busca preço vigente do produto se Vista/Prazo.                                           |
| **UTF-8 em toda a stack**                  | Go middlewares adicionam `Content-Type: application/json; charset=utf-8`; `index.html` com `<html lang="pt-BR">` e `<meta charset="UTF-8">`.                  |
| **Builds validados**                       | `ng build` OK, budgets ajustados em `angular.json` (não avisa warning de initial 2MB).                                                                        |

---

## 9. História de commits (40 passos, 11 Sprints)

A história completa (mapa sprint × commit × hash curto × data × referência SDD / Parte da
arquitetura / Sprint do planejamento) está em:

- **[docs/commits.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/docs/commits.md)** — mapa oficial 40 commits
- **[docs/terminal-historia-985-1010.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/docs/terminal-historia-985-1010.md)** — página da história reconstruída
- **Pasta `commits/` na raiz** — 40 arquivos `commit001.stamp` … `commit040.stamp` como prova
  material de cada etapa (1 arquivo criado em cada commit para forçar o diff e garantir que
  nenhum passo foi pulado).

---

## 10. Referências dos 3 documentos que guiaram TUDO

| Documento                        | Caminho                                                                                                                          | O que cobre                                                                  |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| **Especificação do teste (SDD)** | [sdd_teste.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/sdd_teste.md)                 | 26 seções §1..§26 (objetivo → glossário). Obrigatoriedades do teste técnico. |
| **Arquitetura**                  | [docs/arquitetura.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/docs/arquitetura.md)   | 12 Partes (fundação → UI + tema + header responsivo).                        |
| **Planejamento 11 Sprints**      | [docs/planejamento.md](file:///c:/Users/leoka/OneDrive/%C3%81rea%20de%20Trabalho/Korp_Teste_LeonardoAraujo/docs/planejamento.md) | S1..S11 com checklists de entrega por sprint.                                |
