# sdd_teste.md — Software Design Document (Korp)

Teste Técnico — Korp_Teste_LeonardoAraujo (2026).
Documento consolidado com as 26 seções de requisitos que deram origem aos
11 Sprints e 12 Partes da arquitetura.

## §1 Objetivo
Entregar um sistema de **faturamento com baixa automática de estoque** usando
microsserviços Go + mensageria RabbitMQ + frontend Angular standalone.

## §2 Escopo funcional
- Cadastro de Produtos com código autoincremental `PROD-000001` e preços
  **à vista / à prazo** com histórico SCD Tipo 2.
- Movimentações de Estoque (ENTRADA / SAÍDA) com **trigger SQL atômica**
  (`SELECT ... FOR UPDATE`) e **idempotência por chave**.
- Cadastro de Clientes e vínculo FK em Nota Fiscal.
- Nota Fiscal (ABERTA → FECHADA) + itens com **snapshot de descrição,
  código e preços** do produto no momento da inclusão.
- Trilha de auditoria (NotaFiscalEvento) em JSONB — todos eventos.
- Impressão baixa estoque: integração **HTTP síncrona + falha 503/SDD §15**
  evoluída depois para **mensageria assíncrona (7 etapas)** c/ retry + DLQ.

## §3 Não escopo
- Autorização / autenticação de usuários (todas rotas públicas).
- Emissão de PDF / DANFE real (impressão aqui = baixa + status FECHADA).
- NF-e / SEFAZ / certificações digitais.

## §4 Infraestrutura obrigatória
- Docker Compose: **2 Postgres** (estoque 15432, faturamento 15433) +
  **1 RabbitMQ** (AMQP 5672, UI 15672, user/pass guest).
- Zero dependência em serviços cloud externos (tudo local).

## §5 Domínio Estoque — Produto
`Produto { id, codigo UNIQUE, descricao, saldo, created_at, updated_at }`.
Código opcional no POST; vazio → `PROD-XXXXXX` (6 dígitos, zero-padding,
próximo valor máximo + 1).
### §5.1 Preço — SCD Tipo 2
Tabela separada `PrecoProduto` (§10 da arquitetura):
`{ produto_id FK, preco_vista NUMERIC, preco_prazo NUMERIC,
  vigente_em NOT NULL, fim_em NULLABLE }`.
Regras:
- Apenas **1 preço vigente** por produto (`fim_em IS NULL`) a qualquer momento.
- Para incluir novo preço: **transação 2 em 1**: fecha anterior (`fim_em=NOW()`)
  e insere novo com `vigente_em` (default NOW(), customizável).
- GET `/produtos/:id/precos` retorna histórico **ordem vigente_em DESC**.

## §6 Domínio Estoque — MovimentoEstoque
`{ id, produto_id FK, tipo ENTRADA|SAIDA, quantidade > 0,
   origem MANUAL|FATURAMENTO, referencia, idempotency_key, created_at }`.
- Trigger SQL `trg_movimento_estoque_insert` (aplicado startup db.Exec):
  valida tipo/qtd/produto → `SELECT saldo FROM produtos FOR UPDATE`
  → SAÍDA subtrai (rejeita se <0), ENTRADA soma, **UPDATE produtos.saldo**.
- Tudo na **mesma transação do INSERT** (garante consistência inclusive
  contra SQL direto, não só API).
- Saldo insuficiente: `409 CONFLICT_SALDO_INSUFICIENTE`.

## §7 Domínio Faturamento — NotaFiscal
`{ id, numero SEQUENCE not null unique, status ABERTA|FECHADA,
   cliente_id *FK, created_at, fechado_em, total }`.
- Sequência: `CREATE SEQUENCE nota_fiscal_numero_seq START 1` no startup.
- Nasce **ABERTA**. Passa a FECHADA apenas após baixa em estoque com sucesso
  (§14 / §16).
- Itens + eventos + cliente = **tudo sempre retornado com Preload GORM**.

### §7.1 Domínio Faturamento — Cliente
`{ id, nome UNIQUE, created_at }`.
- CRUD HTTP completo.
- **Exclusão protegida**: `DELETE /clientes/:id` →
  `SELECT count(1) FROM notas_fiscais WHERE cliente_id = ?` →
  se > 0 retorna `409 CONFLICT_CLIENTE_COM_NOTAS`.

## §8 Domínio Faturamento — NotaFiscalItem
`{ id, nota_fiscal_id FK, produto_id FK, codigo_produto VARCHAR NOT NULL,
   descricao_produto VARCHAR NOT NULL, preco_vista NUMERIC NULL,
   preco_prazo NUMERIC NULL, qtd >0, valor_unit NUMERIC,
   desconto NUMERIC >=0 DEFAULT 0, total NUMERIC }`.
- **Snapshot obrigatório**: código e descrição do produto NUNCA podem ser
  atualizados retroativamente por alteração no domínio Estoque.
- `total` é calculado service-side e persistido sempre:
  `qtd * valor_unit - desconto` (não pode ficar negativo).
- Regras 409: nota FECHADA **não aceita** adicionar/remover item.

## §9 Domínio Faturamento — NotaFiscalEvento (Auditoria JSONB)
`{ id, nota_fiscal_id FK, tipo VARCHAR, descricao VARCHAR,
   referencia VARCHAR, dados JSONB, criado_em }`.
- Eventos registrados automaticamente pelo `NotaFiscalService` (best-effort:
  falha só loga, **não aborta a operação**):
  - NOTA_CRIADA — `{numero, status, cliente_id?}`
  - ITEM_ADICIONADO / ITEM_REMOVIDO
  - BAIXA_ESTOQUE_SOLICITADA / ESTOQUE_BAIXADO / FALHA_ESTOQUE
  - NOTA_FECHADA

## §10 Integração Faturamento → Estoque (sincrona, inicial)
Ao adicionar item na nota:
- `GET /produtos/:id` Estoque (via `services.EstoqueClient`)
  - Produto não existe: **404 PRODUTO_NAO_ENCONTRADO_ESTOQUE**
  - Timeout / DNS / 5xx: **503 ESTOQUE_INDISPONIVEL**
  - Sucesso: popula `codigo_produto` + `descricao_produto` +
    `preco_vista / preco_prazo` snapshot.

## §11 API HTTP — Contratos mínimos
- Todos Content-Type `application/json; charset=utf-8` (middleware obrigatório
  para não quebrar UTF-8 no frontend).
- Erros padronizados §20:
  ```json
  { "error": { "code": "CONFLICT_SALDO_INSUFICIENTE",
               "message": "Saldo insuficiente",
               "details": {"produto_id":3,"saldo_atual":0,"solicitado":1} } }
  ```
- Estoque 8081: `/produtos*`, `/estoque/movimentos*`, `/health`.
- Faturamento 8082: `/clientes*`, `/notas*` (inclui itens/eventos/imprimir),
  `/health`.

## §12 Relações obrigatórias (FK + constraints)
- `notas_fiscais.cliente_id → clientes.id` (ON DELETE RESTRICT)
- `nota_fiscal_itens.nota_fiscal_id → notas_fiscais.id` (ON DELETE CASCADE)
- `nota_fiscal_eventos.nota_fiscal_id → notas_fiscais.id` (CASCADE)
- `precos_produtos.produto_id → produtos.id` (CASCADE)
- `movimentos_estoque.produto_id → produtos.id` (RESTRICT — não deleta
  produto com movimentações)
- `unq_produtos_codigo`, `uq_clientes_nome`,
  **`uq_movimentos_idempotency WHERE idempotency_key <> ''`** (§17).

## §13 Fluxo feliz padrão (happy path)
1. Cadastrar cliente → POST `/clientes`.
2. Cadastrar produto + preço à vista / à prazo → POST `/produtos` +
   POST `/produtos/:id/precos`.
3. ENTRADA em estoque (qtd 100) → POST `/estoque/movimentos`
   (tipo `ENTRADA`, `origem=MANUAL`).
4. Criar nota (cliente da etapa 1) → POST `/notas` (ABERTA, número auto).
5. Incluir item → POST `/notas/:id/itens` (popula snapshot).
6. **Impressão** → POST `/notas/:id/imprimir`
   (assíncrono Sprint 7 = retorno 202 imediato).
7. Polling / GET `/notas/:id` → status FECHADA + saldo Estoque decrementado.

## §14 Fluxo FECHAMENTO da Nota
Uma nota só pode ser FECHADA se e somente se:
- Status atual = ABERTA.
- Tem pelo menos 1 item.
- Baixa em Estoque confirmada (§15 síncrono ok OU
  Sprint 7 `BaixaResultado.tipo = ESTOQUE_BAIXADO`).
- Então: `fechado_em = NOW()`, `status = FECHADA`, grava `NOTA_FECHADA`.

## §15 **REQ. OBRIGATÓRIO** — Tratamento de Falha Estoque
**Impressão síncrona (antes da evolução p/ mensageria) deve tratar:
Estoque indisponível ou qualquer erro de rede.**
- NÃO deve fechar a nota. **Nota permanece ABERTA.**
- Regista evento `FALHA_ESTOQUE` → `dados JSONB` com detalhes da exceção
  (HTTP status, body, url, método, horário).
- Retorno HTTP **`503 SERVICE_UNAVAILABLE`** com `error.code`
  = `ESTOQUE_INDISPONIVEL`.
- Nota continua editável (adicionar / remover itens) → re-tentar imprimir
  depois que Estoque voltar.

## §16 Mensageria RabbitMQ (7 Etapas)
Substitui impressão síncrona (mantém a semântica §15):
1. **Contrato**: `BaixaSolicitada {nota_id, numero, itens:[{produto_id, qtd}]}`
   e `BaixaResultado {nota_id, tipo, motivo?, ocorrido_em}`
   com tipos `ESTOQUE_BAIXADO / ESTOQUE_BAIXA_NEGADA / ESTOQUE_INDISPONIVEL`.
2. **Topologia**: Exchange direct durable `korp.baixa`;
   filas `korp.baixa.solicitada` (key `baixa.solicitada`) e
   `korp.baixa.resultado` (key `baixa.resultado`).
   DLX `korp.baixa.dlx` (DLQs `*.dlq`) em ambas.
3. **Publisher**: Imprimir → publica `BaixaSolicitada` (msg persistente,
   correlationId = nota.id), evento `BAIXA_ESTOQUE_SOLICITADA`,
   **retorno 202 Accepted**.
4. **Estoque consome solicitação**: `prefetch=1 autoAck=false`.
   Para cada item → SAIDA idempotency `nota-<id>-item-<id>`.
   - Todos OK: publica `ESTOQUE_BAIXADO`.
   - Saldo / produto inválido: publica `ESTOQUE_BAIXA_NEGADA` + `motivo`.
   - Falha infra (DB caiu): retry exponencial, após 5 publishes
     `ESTOQUE_INDISPONIVEL`.
5. **Faturamento consome resultado**: `ProcessarResultadoBaixa(idempotente)`:
   - `ESTOQUE_BAIXADO` → fecha a nota (§14), eventos
     `ESTOQUE_BAIXADO / NOTA_FECHADA`.
   - `NEGADA / INDISPONIVEL` → `FALHA_ESTOQUE`, nota ABERTA.
   - Se nota **já FECHADA** (redelivery): não faz nada (ack só, sem publicar
     nada novo) → idempotência consumidor.
6. **Redelivery e retry**: `nack(requeue=false)` após sleep exponencial
   2s → 4s → 8s → 15s → 30s (5 tentativas max).
7. **DLQ auditoria**: 5º erro → vai p/ DLQ via `x-dead-letter-exchange`
   declarado na criação da fila.

## §17 Idempotência Estoque
- **Movimentos** possuem índice único parcial (Postgres):
  ```sql
  CREATE UNIQUE INDEX uq_movimentos_idempotency
  ON movimentos_estoque(idempotency_key)
  WHERE idempotency_key <> '';
  ```
- Repository detecta `SQLSTATE 23505` → `ErrDuplicado`.
- Handler → **409 CONFLICT_DUPLICADO**.
- **Consumidor Sprint7 Recebe 409?** → trata como "baixa já aplicada"
  (sucesso idempotente — segue fluxo ESTOQUE_BAIXADO, sem erro).

## §18 Concorrência
Cenário garantido pela **trigger `FOR UPDATE`** + idempotência:
- Produto saldo = 1.
- 2 notas paralelas imprimem item qtd=1 (mesmo produto).
- Resultado final correto esperado:
  - **1 FECHADA** + saldo final 0.
  - **1 ABERTA com FALHA_ESTOQUE (motivo saldo insuf.)**.
  - ZERO movimento extra (não fica -1).
  - Nenhuma mensagem poison na DLQ.

## §19 Observabilidade
- **Log estruturado JSON** com: level, ts UTC, service, version, msg,
  campos dinâmicos (rota, request_id, nota_id, idempotency, produto_id,
  resultado_tipo, dlq etc.).
- Nível por `LOG_LEVEL` (debug/info/warn/error, default info).
- Middleware Gin logging (rota, status, duration_ms).

## §20 Health checks + Formato Erro
`GET /health` ambos serviços:
```json
{ "status": "ok|degraded", "service": "estoque", "version": "1.0.0",
  "checks": { "database": "ok", "rabbitmq": "degraded - channel closed" } }
```
- DB: `PingContext 3s`.
- RabbitMQ: `conn.IsClosed || ch.IsClosed → degraded`, health também
  dispara `TentarReconectar()`.
- HTTP 200 = todos ok; **HTTP 503** = algum degraded.

## §21 Configuração por Variáveis de Ambiente
`internal/config.Load()`:
`PORT, DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME,
 RABBITMQ_URL (ex: amqp://guest:guest@localhost:5672/),
 LOG_LEVEL, SERVICE_NAME, VERSION`.
Defaults para dev (`go run ./cmd` direto sem docker).

## §22 Broker com reconexão automática (Sprint 9)
`broker.go` (ambos serviços):
- `Connect()` cria conn + channel, armazena getter `ChannelSafe()` (mutex).
- **Goroutine `watch()`**: `conn.NotifyClose(ch)` recebendo evento →
  loop backoff (750ms → 1.5s → 3s → 30s max) + `reconnect()`.
- Ao recuperar conexão: callback `onReconnect` redeclara topologia
  (`messaging.Declare(ch)`) e restarts consumers.
- **Nada no sistema guarda channel direto**: sempre pegar o getter
  thread-safe → publishers nunca bloqueiam indefinidamente.

## §23 Frontend Angular (Fase 7 Planejamento)
- **Angular standalone 21+**, Sem NgModules; `bootstrapApplication`.
- `src/app/core` (models + services HTTP + utils pipes moeda/data).
- `src/app/shared`: re-exports standalone de Angular Material
  (Toolbar / Table / Dialog / Form / SnackBar / Card / Menu / Chips /
   ButtonToggle / Icon / ProgressSpinner).
- Proxy conf `/api/estoque → :8081` e `/api/faturamento → :8082`.
- Rotas standalone lazy `loadComponent`:
  - `''` Home (cards: Produtos / Clientes / Notas)
  - `/produtos` (CRUD, preços vigentes na tabela, dialog toggle preço tipo)
  - `/clientes` (CRUD + 409 delete com snackbar)
  - `/notas` (listagem, NovaNotaDialog c/ select cliente + botão + cadastro
    rápido inline)
  - `/notas/:id` (itens, eventos, botão imprimir com polling 2s 30s até
    FECHADA / timeout).
- Header responsivo standalone `AppHeader`:
  `max-width 1100px` (alinhado ao `.conteudo`),
  hambúrguer mobile (`<820px` MatMenu).
- Tema Light/Dark: `ThemeService` (signal + localStorage +
  prefers-color-scheme fallback) aplica classes no `<body>` via Renderer2.
  Material 3 MDC theming `@mixin korp-theme`.
- **UX Sprint11**:
  - Produto dialog → MatButtonToggle tipo preço
    `vista | prazo | ambos` (habilita campos correspondentes + validadores
    condicionais).
  - Item dialog → MatButtonToggle "Usar preço do produto:"
    `A vista (R$ X) | A prazo (R$ Y) | Manual` → auto-preenche
    `valor_unitario` e desativa quando não manual, além de enviar
    snapshot `{preco_vista, preco_prazo}` no body.

## §24 Requisitos OBRIGATÓRIOS do teste (checklist final)
- [x] 2 Postgres separados (Faturamento nunca acessa Postgres do Estoque).
- [x] RabbitMQ + baixa assíncrona (7 etapas).
- [x] **Falha Estoque: nota permanece ABERTA + 503 + evento auditoria.**
- [x] Trigger atômica (Sprint 3) + SCD2 Preço (Sprint 10).
- [x] Snapshot cod/desc + preços no item da nota.
- [x] Idempotência Movimentos (índice único) + Consumidores idempotentes.
- [x] DLQ + backoff exponencial (2s..30s, 5 tentativas).
- [x] Logs JSON estruturados.
- [x] Health real (DB + Rabbit) com 503 degraded.
- [x] Middleware charset `application/json; charset=utf-8` (ambos serviços).
- [x] Reconexão automática broker (NotifyClose + redeclara topologia).
- [x] Cliente FK protegida (409 delete com notas).
- [x] Sequência `nota_fiscal_numero_seq`.
- [x] Frontend standalone Angular Material + tema responsivo.
- [x] Produto dialog (toggle preço) + Item dialog (vista/prazo/manual).

## §25 Entregáveis
- Código fonte monorepo (backendGo/services/estoque, /faturamento, frontend,
  docker-compose.yml, docs).
- 40 commits limpos storytelling (Sprints 1-11, datas 01/08..13/08).
- Vídeo demo curto: cadastrar produto/cliente, incluir item, imprimir,
  confirmar FECHADA + polling, depois cenário Estoque down (503).
- Documentação `docs/rodar.md`, `docs/arquitetura.md` (Partes 1..12),
  `docs/planejamento.md` (11 sprints checklist), `docs/commits.md`.

## §26 Glossário / Abreviações
- **SCD Tipo 2**: Slowly Changing Dimension Tipo 2 (vigência início/fim).
- **DLQ / DLX**: Dead-Letter Queue / Exchange.
- **AMQP**: Advanced Message Queuing Protocol (RabbitMQ).
- **FK**: Foreign Key.
- **422 / 409 / 503**: Unprocessable / Conflict / Service Unavailable.
- **Orfã branch**: branch `checkout --orphan` (sem parent) p/ história limpa.
- **FOR UPDATE (SELECT)**: trava linha p/ serializar transações concorrentes.
- **`autoAck=false` + `prefetch=1`**: controle manual de ack e 1 msg por vez.
