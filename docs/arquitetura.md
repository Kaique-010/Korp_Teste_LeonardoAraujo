# Korp Teste — Detalhamento Técnico

Divisão lógica da SDD (`sdd_teste.md`) em módulos coesos para orientar a implementação incremental. Cada parte referencia as seções originais da SDD.

---

## Parte 1 — Fundamentos do Sistema

**Referência SDD:** 1, 2, 3, 26

- Objetivo: sistema de emissão de notas fiscais em microsserviços.
- Backend em Go; frontend em Angular (posterior ao backend).
- Mínimo de dois microsserviços: Estoque e Faturamento.
- Cada microsserviço possui banco próprio (PostgreSQL) e comunicação via HTTP/REST e RabbitMQ.
- Princípio de propriedade dos dados: Faturamento nunca acessa o banco de Estoque diretamente.
- Stack: Go, Gin, GORM, PostgreSQL, RabbitMQ, Docker Compose, Angular, RxJS, Angular Material.

## Parte 2 — Modelagem de Domínio

**Referência SDD:** 4, 10

- Estoque: `Produto` (1:N) `MovimentoEstoque`.
- Faturamento: `NotaFiscal` (1:N) `NotaFiscalItem` e (1:N) `NotaFiscalEvento`.
- Sem Foreign Key física entre bancos — `produto_id` é referência lógica.
- Padrão ERP: cabeçalho + itens + histórico de eventos.

## Parte 3 — Serviço de Estoque

**Referência SDD:** 5, 6, 11

- `Produto`: id, codigo (único, auto-gerado sequencial `PROD-000001` se não informado), descricao, saldo, timestamps.
  - Relacionamento 1:N com `PrecoProduto` (histórico); `precos []PrecoProduto` e campo virtual `preco_atual` no JSON.
- `PrecoProduto` — model separado com FK + SCD Tipo 2 (histórico de alterações):
  - Campos: `id`, `produto_id` (FK), `preco_vista`, `preco_prazo`, `vigente_em`, `fim_em` (nulo enquanto vigente), `criado_em`.
  - Ao inserir novo preço: em transação fecha o anterior (atualiza `fim_em = now()` do registro vigente) e cria novo com `vigente_em` informado ou now(). Histórico mantido por padrão (SCD2), consulta `preco_atual` sempre retorna o vigente (sem `fim_em`).
- `MovimentoEstoque`: id, produto_id, tipo (ENTRADA/SAIDA), quantidade, origem, referencia, idempotency_key.
  - Toda alteração de saldo gera movimentação.
- **Validação de delete**: `TemMovimentos()` no Repository retorna `409 Conflict` se o produto tiver movimentações (evita erro de FK no banco).
- **Trigger de estoque** (`trg_movimento_estoque_insert`): o saldo é atualizado pelo próprio banco na inserção de um movimento, na mesma transação. A trigger valida tipo, quantidade > 0, existência do produto e saldo suficiente, e usa `SELECT ... FOR UPDATE` para serializar operações concorrentes sobre o mesmo produto (base do Sprint 8). O Service valida primeiro para feedback rápido; a trigger é a garantia final no banco — funciona mesmo fora da aplicação.
- API:
  - Produtos: `POST /produtos {codigo?, descricao, saldo, preco_vista?, preco_prazo?, vigente_em?}`, `GET /produtos`, `GET /produtos/:id`, `PUT /produtos/:id`, `DELETE /produtos/:id` (409 se tem movimento).
  - Preços (rotas próprias): `GET /produtos/:id/precos` (histórico) e `POST /produtos/:id/precos {preco_vista, preco_prazo, vigente_em?}` (novo preço fecha o anterior).
  - Movimentos: `POST /estoque/movimentos {produto_id, tipo:ENTRADA|SAIDA, quantidade, origem, referencia?, idempotency_key?}`.
  - `GET /` — página inicial JSON descritiva com `service`, `version`, `links[]` e `rotas` auto-documentadas.

## Parte 4 — Serviço de Faturamento

**Referência SDD:** 7, 8, 9, 12

- `Cliente`: id, nome (único), timestamps.
  - FK opcional em `NotaFiscal.cliente_id` (Nota pertence a 0 ou 1 Cliente; Cliente tem 0..N notas).
  - Delete protegido: `TemNotas()` → 409 Conflict se cliente tiver notas (evita FK error).
  - CRUD completo: `POST /clientes`, `GET /clientes`, `GET /clientes/:id`, `PUT /clientes/:id`, `DELETE /clientes/:id`.
- `NotaFiscal`: id, numero, status (ABERTA/FECHADA), timestamps, `cliente_id *uint64` (FK), relacionamento `Cliente *Cliente` (preload).
  - Numeração sequencial garantida por sequence do banco (`nota_fiscal_numero_seq`).
  - Status só muda para `FECHADA` ao receber `ESTOQUE_BAIXADO` da fila de resultado (não ao Imprimir).
- `NotaFiscalItem`: id, nota_fiscal_id, produto_id, codigo_produto (snapshot), descricao_produto (snapshot), quantidade, valor_unitario, desconto, total, **`preco_vista` e `preco_prazo` (snapshot do PrecoProduto vigente no momento da inclusão)**.
  - Ao incluir item, o Faturamento consulta o Produto via HTTP no Estoque: 404 se inexistente, 503 se Estoque indisponível, e preenche `preco_vista`/`preco_prazo` a partir de `produto.preco_atual` (permite override por body).
  - Snapshot mantém código/descrição da emissão; `total = quantidade × valor_unitario − desconto`.
  - Itens só podem ser adicionados/removidos enquanto status = ABERTA (409 Conflict em notas fechadas).
- `NotaFiscalEvento`: id, nota_fiscal_id, tipo, descricao, referencia, dados (jsonb).
  - Auditoria do ciclo de vida: `NOTA_CRIADA`, `ITEM_ADICIONADO`, `ITEM_REMOVIDO`, `BAIXA_ESTOQUE_SOLICITADA`, `ESTOQUE_BAIXADO`, `NOTA_FECHADA`, `FALHA_ESTOQUE`. Service.Registrar(notaID, tipo, descricao, referencia, dados).
- API:
  - Clientes: `POST /clientes {nome}`, `GET /clientes`, `GET /clientes/:id`, `PUT /clientes/:id {nome}`, `DELETE /clientes/:id` (409 se tem notas).
  - Notas: `POST /notas {cliente_id?}`, `GET /notas`, `GET /notas/:id`, `GET /notas/:id/eventos`, `POST /notas/:id/itens {...}`, `DELETE /notas/:id/itens/:item_id`, `POST /notas/:id/imprimir`.
  - `GET /` — página inicial JSON descritiva com rotas documentadas.

## Parte 5 — Fluxos de Negócio

**Referência SDD:** 13, 14

- Criação: nota nasce `ABERTA` e recebe itens enquanto aberta.
- Impressão (`POST /notas/:id/imprimir`) — assíncrona (Sprint 7):
  1. Validar nota e status (`ABERTA`), deve ter itens (422 se vazia);
  2. Montar `BaixaSolicitada` `{nota_id, numero, itens:[{produto_id, quantidade}]}` e publicar no exchange `korp.baixa` (`baixa.solicitada`), persistente;
  3. Registrar `BAIXA_ESTOQUE_SOLICITADA` e retornar **202 Accepted** com a nota (que permanece `ABERTA`);
  4. Em falha de publicação: registrar `FALHA_ESTOQUE`, nota permanece `ABERTA`, retornar 503.
- Consumo e fechamento assíncronos (Etapas 4–6): o **Estoque** consome `korp.baixa.solicitada`,
  executa a baixa por item (`SAIDA`, origem `FATURAMENTO`, ref `nota-<id>`, idempotency
  `nota-<id>-item-<produto>`) e publica o resultado; o **Faturamento** consome
  `korp.baixa.resultado` e fecha a nota em `ESTOQUE_BAIXADO` (ou registra `FALHA_ESTOQUE` em
  negada/indisponível).
- Sem transação distribuída: baixa assíncrona e atômica por item pelo Estoque; a `idempotency_key`
  evita baixa duplicada em reenvio (Sprint 8).

## Parte 6 — Resiliência e Consistência

**Referência SDD:** 15, 16, 17, 18, 19

- Falha do Estoque: nota permanece `ABERTA`, Faturamento registra o erro, usuário recebe feedback.
- Mensageria (RabbitMQ): desacopla operações; introduzida após o fluxo HTTP básico.
  - **Contrato (arquivo `internal/messaging/contract.go`, idêntico nos dois serviços):**
    - exchange `korp.baixa` (direct, durável); fila `korp.baixa.solicitada` (`baixa.solicitada`, consumida pelo Estoque); fila `korp.baixa.resultado` (`baixa.resultado`, consumida pelo Faturamento).
    - `BaixaSolicitada` `{nota_id, numero, itens:[{produto_id, quantidade}]}` — Faturamento → Estoque.
    - `BaixaResultado` `{nota_id, tipo, motivo?, ocorrido_em}` com `tipo` ∈ `ESTOQUE_BAIXADO` | `ESTOQUE_BAIXA_NEGADA` | `ESTOQUE_INDISPONIVEL` — Estoque → Faturamento.
  - **Fluxo do resultado:** `ESTOQUE_BAIXADO` fecha a nota (idempotente); `ESTOQUE_BAIXA_NEGADA`
    (regra de negócio: saldo/produto) e `ESTOQUE_INDISPONIVEL` (falha de infra após esgotar
    retries) deixam a nota `ABERTA` com evento `FALHA_ESTOQUE`. Se o Estoque estiver fora, a
    mensagem de baixa fica na fila e é entregue quando ele voltar (nada se perde).
  - **Retry/backoff:** consumidores usam `prefetch=1`, `autoAck=false` e até 5 tentativas com
    backoff exponencial (2s→30s) na própria carta; falha persistente vai para a **DLQ**.
  - **DLX/DLQ:** filas declaradas com `x-dead-letter-exchange: korp.baixa.dlx`; DLQs
    `korp.baixa.solicitada.dlq` e `korp.baixa.resultado.dlq` (auditoria/reprocessamento manual).
  - Topologia declarada de forma idempotente pelo `internal/messaging/topology.go` (`Declare`).
  - Guia de implementação: `docs/guia-sprint7-mensageria.md`.
- Idempotência: `idempotency_key` impede baixa duplicada em retry.
  - Índice único parcial `uq_movimentos_idempotency` (`WHERE idempotency_key <> ''`) no Postgres do Estoque.
  - Violação `23505` → `ErrDuplicado` → `409 CONFLICT_DUPLICADO` na API.
  - No consumidor, duplicado é **sucesso idempotente**: redelivery (ack perdido) não repete a baixa nem
    publica novo resultado — a mensagem é simplesmente confirmada (DLQ permanece vazia).
- Concorrência: transação + bloqueio no banco; duas notas sobre o mesmo saldo não podem ambas baixar
  (trigger com `SELECT ... FOR UPDATE`; validado com requisições paralelas).
- Consistência: transações locais por serviço; sem transação distribuída.

## Parte 6.1 — Observabilidade (Sprint 9)

- **Logs estruturados JSON** (`internal/logging`): nível (`LOG_LEVEL`), timestamp UTC, `service`,
  `version`, mensagem e campos por evento. Consumidores registram domínio (baixa, idempotência,
  resultados, falhas/DLQ); middleware Gin registra toda requisição (método, path, status, duração).
- **Health checks reais** (`internal/health`): `GET /health` verifica banco (`PingContext`) e
  RabbitMQ (conexão/canal abertos). 200 `{"status":"ok","checks":{...}}` ou 503 `degraded` com o
  detalhe de cada dependência.
- Nível e configuração por variáveis de ambiente (`LOG_LEVEL`, `PORT`, `DB_*`, `RABBITMQ_URL`).

## Parte 7 — Padrões de API e Tratamento de Erros

**Referência SDD:** 20

- 400 (dados inválidos), 404 (não encontrado), 409 (conflito), 422 (regra de negócio), 500 (erro inesperado), 503 (dependente indisponível).
- Formato de erro único: `{ "error": { "code", "message", "details" } }`.

## Parte 8 — Arquitetura Interna do Backend

**Referência SDD:** 21, 22

- Camadas: HTTP → Handler → Service → Repository → Database.
- `Handler`: recebe HTTP, valida entrada, chama Service, monta resposta.
- `Service`: regras de negócio (ex.: fluxo de impressão).
- `Repository`: acesso a dados.
- `Model`: entidades persistidas.
- Estrutura de pastas: `backendGo/services/estoque` e `backendGo/services/faturamento`, cada um com `cmd/`, `internal/{handlers,services,repositories,models,routes}`, `migrations/`.
- Abstração: dentro de cada serviço, `services` e `repositories` definem interface + implementação no mesmo pacote; `handlers` dependem apenas da interface, mantendo banco e regras de negócio trocáveis/testáveis.

## Parte 9 — Estratégia de Desenvolvimento

**Referência SDD:** 23

- Implementação incremental em fases: Estoque → Faturamento → Integração → Mensageria → Robustez → Infra → Angular.
- Detalhamento em checklist: ver `docs/planejamento.md`.

## Parte 10 — Requisitos, Entrega e Critérios

**Referência SDD:** 24, 25

- Obrigatórios: 2 microsserviços, produto (código/descrição/saldo), nota (numeração, status, múltiplos itens), impressão (indicador, fechamento, bloqueio de nota não aberta, atualização de saldo), persistência real, tratamento de falha entre serviços.
- Opcionais: concorrência, IA, idempotência.
- Entrega: repositório público `Korp_Teste_SeuNome`, vídeo demonstrativo, documento técnico, prazo de até 7 dias.

---

---

## Estrutura de pastas

```text
Korp_Teste_LeonardoAraujo/
│
├── backendGo/
│   └── services/
│       ├── estoque/
│       │   ├── cmd/
│       │   ├── internal/
│       │   │   ├── handlers/          → recebe HTTP, chama interface de service
│       │   │   ├── services/          → interface + implementação das regras de negócio
│       │   │   ├── repositories/      → interface + implementação de acesso ao banco
│       │   │   ├── models/            → entidades persistidas
│       │   │   └── routes/            → registro de rotas
│       │   ├── migrations/
│       │   └── go.mod
│       │
│       └── faturamento/
│           ├── cmd/
│           ├── internal/
│           │   ├── handlers/
│           │   ├── services/
│           │   ├── repositories/
│           │   ├── models/
│           │   └── routes/
│           ├── migrations/
│           └── go.mod
│
├── frontend/                → Angular (após o backend)
│
├── docker-compose.yml
│
└── docs/
    ├── arquitetura.md
    └── planejamento.md
```

## Mapa de dependências entre partes

```text
Parte 1 (Fundamentos)
   └── Parte 2 (Domínio) ── orienta a modelagem
         ├── Parte 3 (Estoque)       ── implementa primeiro
         └── Parte 4 (Faturamento)
               └── Parte 5 (Fluxos) ── integração entre os dois
                     └── Parte 6 (Resiliência) ── falhas, mensageria, robustez
Parte 7 (Erros) e Parte 8 (Camadas) ── aplicadas em toda a implementação
Parte 9 (Estratégia) ── sprints
Parte 10 (Requisitos/Entrega) ── validação final
```

---

## Parte 11 — Backend Go: Frameworks, Bibliotecas e Detalhes Internos

Versão: **Go 1.26.5**. Cada microsserviço é um módulo Go independente (`go.mod` próprio, `internal/` com limites de visibilidade).

### 11.1 Frameworks e principais dependências (ambos serviços)

| Dependência                      | Versão    | Papel                                                                                        |
| -------------------------------- | --------- | -------------------------------------------------------------------------------------------- |
| `github.com/gin-gonic/gin`       | `v1.12.0` | HTTP framework (middleware Logger/Recovery, routing, binding JSON, contexto `*gin.Context`). |
| `gorm.io/gorm`                   | `v1.31.2` | ORM: models, AutoMigrate, transações (`db.Begin/Commit/Rollback`), associações, `Preload`.   |
| `gorm.io/driver/postgres`        | `v1.6.2`  | Dialeto PostgreSQL via `jackc/pgx/v5`.                                                       |
| `github.com/rabbitmq/amqp091-go` | `v1.13.0` | Cliente RabbitMQ (pub/consume, channel QoS, confirm mode).                                   |
| `github.com/stretchr/testify`    | `v1.11.1` | Testes unitários (`assert`, `require`, `suite`).                                             |
| `gorm.io/datatypes`              | `v1.2.7`  | `JSON` (jsonb) — utilizado em `NotaFiscalEvento.dados` no Faturamento.                       |
| `github.com/jackc/pgx/v5`        | `v5.10.0` | Driver nativo PostgreSQL (Estoque).                                                          |
| `github.com/glebarez/sqlite`     | `v1.11.0` | SQLite em memória para testes de repository/service (sem dependência externa).               |

### 11.2 Camadas e fluxo de uma requisição (exemplo: criar produto com preço)

```
   HTTP POST /produtos JSON body
         │
         ▼
  routes/routes.go  ── gin.RouterGroup: registra handler + middleware
         │
         ▼
  handlers/ProdutoHandler.Create
         │  ShouldBindJSON → input do service (tipagem forte)
         │  writeError(c, err) mapeia AppError → status + JSON
         ▼
  services/ProdutoService.Create(CreateProdutoInput)
         │  → ProximoCodigo() repo (gera PROD-XXXXXX se vazio)
         │  → Cria Produto; se input tiver preco_vista/preco_prazo:
         │      repo.AdicionarPreco(produto.ID, preco)
         │         → abre transação,
         │           UPDATE precos_produtos SET fim_em=now() WHERE produto_id=? AND fim_em IS NULL
         │           INSERT novo PrecoProduto
         │         → COMMIT
         ▼
  repositories/ProdutoRepository + PrecoProduto via GORM
         ▼
   PostgreSQL (estoque_db:15432) — trigger na tabela movimentos
```

### 11.3 RabbitMQ: Broker com reconexão automática + ChannelGetter

Arquitetura tolerante a restart do RabbitMQ:

- `internal/broker/broker.go`
  - Campos: `sync.Mutex`, `url string`, `conn *amqp.Connection`, `ch *amqp.Channel`, `onReconnect func(ch) error`.
  - `Connect(url, cb)`: dial + `Confirm(false)` (publisher confirms), salva callback `onReconnect` que re-declara a topologia `messaging.Declare(ch)` e reinicia consumers.
  - `IsHealthy()`: detecta channel/conn fechado e reconecta on-the-fly com 3 retries (backoff exponencial 750ms); `/health` nunca fica degraded permanentemente.
  - `watch()` goroutine: `conn.NotifyClose()` → ao receber evento → `open()` reconecta + roda callback.
  - `ChannelSafe()` getter (thread-safe) — **publishers e consumers NÃO guardam `*amqp.Channel` diretamente**.
- **Pattern ChannelGetter** `type ChannelGetter func() *amqp.Channel`
  - Todos publishers (`BaixaPublisher`, `BaixaResultadoPublisher`) e consumers (`BaixaConsumer`, `BaixaResultadoConsumer`) recebem getter no construtor.
  - A cada `Publicar*()` ou `Restart()`: chamam `getChannel()` — se channel caiu, `IsHealthy()` reconecta.
- Consumers tem `Restart() error` (mutex + context cancel + WaitGroup) → onReconnect chama `consumer.Restart()` sobre o channel novo.

### 11.4 Contratos internos (interfaces tipadas)

```go
// estoque/internal/services
type ProdutoRepository interface {
    Create(p *models.Produto) error; GetByID(id); List()
    Update(p); Delete(id); ProximoCodigo() (string, error)
    AdicionarPreco(produtoID uint64, preco *models.PrecoProduto) error
    PrecoAtual(produtoID); ListarPrecos(produtoID)
    TemMovimentos(produtoID) (bool, error)
}

// faturamento/internal/services
type NotaFiscalService interface {
    Criar(...CriarNotaInput) (*models.NotaFiscal, error)
    Obter(id); Listar()
    AdicionarItem(notaID, item) (*models.NotaFiscal, error)
    RemoverItem(notaID, itemID); Imprimir(notaID)
    ProcessarResultadoBaixa(resultado messaging.BaixaResultado) error
}
type BaixaPublisher interface { PublicarSolicitacao(messaging.BaixaSolicitada) error }
type ClienteRepository interface {
    Create, GetByID, List, Update, Delete
    TemNotas(clienteID uint64) (bool, error)
}
```

---

## Parte 12 — Frontend Angular: RxJS, Ciclos de Vida e Camadas

Versões: **Angular 21.2.x**, **RxJS ~7.8.0**, **TypeScript ~5.9.2**, **Angular Material 21.2.14**.

### 12.1 Dependências visuais e de runtime (`package.json`)

| Lib                                  | Versão    | Propósito                                                                                                                                                        |
| ------------------------------------ | --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@angular/material` + `@angular/cdk` | `21.2.14` | Design System: tabelas (`mat-table`), formulários (`mat-form-field`, `mat-input`, `mat-select`), diálogos (`MatDialog`), snack-bars (`MatSnackBar`), paginação.  |
| `@angular/animations`                | `21.2.20` | Dependência do Material (trigger fade, slide).                                                                                                                   |
| `rxjs`                               | `~7.8.0`  | Streams: `HttpClient` retorna `Observable<T>`; `BehaviorSubject` stores (estado); `pipe(catchError, tap, retry, switchMap, debounceTime, distinctUntilChanged)`. |
| `@angular/router`                    | `21.2.0`  | Lazy loading de Features Modules (ex: `ProdutosModule`, `NotasModule`, `ClientesModule`), `canActivate` guards, params resolver.                                 |
| `@angular/forms`                     | `21.2.0`  | Reactive Forms (`FormGroup`, `FormArray`, `Validators`, async validators) e Template-driven.                                                                     |
| `tslib`                              | `^2.3.0`  | Runtime helpers (tslib: `__awaiter`, `__decorate`, etc.).                                                                                                        |

Dev tools: `@angular/cli@21.2.21` (schematics, `ng serve`/`ng build` com Vite/Esbuild via `@angular/build`), `prettier@3.8.1`.

### 12.2 Camadas recomendadas (Frontend)

```
frontend/src/app
│
├── core/
│   ├── models/            → interfaces: Produto, PrecoProdutoView, Cliente, NotaFiscal, NotaFiscalItem, NotaFiscalEvento
│   ├── services/          → HttpClient API layer (ProdutoService, ClienteService, NotaService, PrecoService)
│   ├── guards/            → rotas protegidas
│   └── interceptors/      → HttpInterceptor: base URL, tratamento de erro global (apperrors mapeados para MatSnackBar)
│
├── shared/
│   ├── components/        → Reusáveis: PageHeader, ConfirmDialog, LoadingSpinner, ErroMensagem
│   └── material.module.ts → imports/exports do Material (concentrado)
│
└── features/
    ├── produtos/
    │   ├── listagem/      (ProdutosListComponent com <mat-table> + paginação)
    │   ├── formulario/    (ProdutoFormComponent: ReactiveForms + PrecoFormGroup aninhado)
    │   └── historico-precos (dialog com timeline PrecoProduto vigência/fim_em)
    ├── clientes/          (CRUD similar)
    └── notas/
        ├── listagem
        ├── detalhe/       (itens: FormArray de itens; botão IMPRIMIR dispara POST /notas/:id/imprimir + polling GET /notas/:id até FECHADA)
        └── eventos/       (tabela de eventos da nota)
```

### 12.3 RxJS — padrões usados em telas

- **Listagem (com refresh após criar/editar):** `BehaviorSubject<Produto[]>` no Service + `$lista = subject.asObservable()`. Componentes assinam `lista$ | async` no template (Angular `async` pipe faz `unsubscribe` automático no destroy).
- **Debounce em search/busca:**
  ```ts
  this.searchControl.valueChanges
    .pipe(
      debounceTime(300),
      distinctUntilChanged(),
      switchMap((term) => this.service.buscar(term)),
      catchError((err) => {
        this.snack.open(err.error.error.message)
        return of([])
      }),
    )
    .subscribe((items) => (this.dataSource.data = items))
  ```
- **Polling da Nota após Imprimir (enquanto ABERTA → FECHADA):**
  ```ts
  interval(2000)
    .pipe(
      take(15), // timeout ~30s
      concatMap(() => this.notaService.obter(id)),
      takeWhile((n) => n.status === 'ABERTA', true),
    )
    .subscribe(
      (n) => (this.nota = n),
      (err) => {},
      () => this.atualizarStatusVisual(),
    )
  ```

### 12.4 Ciclos de vida (Lifecycle Hooks) de componentes

| Hook              | Uso típico                                                                                                            |
| ----------------- | --------------------------------------------------------------------------------------------------------------------- |
| `ngOnInit`        | Inicializar FormGroup, assinar `route.params` (resolver de Nota/Produto), carregar lista primeira vez.                |
| `ngOnChanges`     | Atualizar view quando `@Input()` muda (ex: atualizar `FormArray` de itens quando a nota é re-carregada).              |
| `ngOnDestroy`     | `unsubscribe` manual de subscriptions sem `async` pipe (ex: polling que não usa `takeWhile`/`takeUntil(onDestroy$)`). |
| `ngAfterViewInit` | Ajustar `MatPaginator`/`MatSort` no `MatTableDataSource` (garantir que ViewChild existe).                             |

### 12.5 Tratamento de erros HTTP (app/core/interceptors/error.interceptor.ts)

- 400 (Bad Request) → extrai `error.message` para mensagem amigável.
- 404 → navega para página 404 ou desfaz formulário.
- 409 Conflict → exibe SnackBar "Conflito: registro possui dependências" (produto com movimento, cliente com nota).
- 422 / 503 → notifica usuário com retry automático em streams que usam `pipe(retry(2))`.
- 500 → tela de erro + informar suporte.

### 12.6 Build e Dev

- **Local:** `npm start` (`ng serve --port 4200`) — aponta para backends `http://localhost:8081` (Estoque) e `http://localhost:8082` (Faturamento) via `environment.ts`.
- **Produção:** `ng build --configuration=production` (output em `dist/korp-frontend/browser`) + `ASSET_PREFIX` configurável; integração com Nginx ou CDN.
- **Testes unitários (opcional):** `ng test` (Karma/Jasmine por padrão; substituível por Jest).
