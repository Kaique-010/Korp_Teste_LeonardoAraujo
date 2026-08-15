# Korp Teste — Plano de Sprints

## Sprint 1 — Fundação do Projeto ✅

**Objetivo:** estrutura base, infraestrutura local e ponto de partida comum.

- [x] Inicializar repositório Git (`git init`)
- [x] Criar `.gitignore` (Go, Node, IDE, .env)
- [x] Criar `README.md` com visão geral do projeto
- [x] Criar `docker-compose.yml` com:
  - [x] PostgreSQL do Estoque
  - [x] PostgreSQL do Faturamento
  - [x] RabbitMQ
- [x] Criar `backendGo/services/estoque/go.mod` (`go mod init`)
- [x] Criar `backendGo/services/faturamento/go.mod` (`go mod init`)
- [x] Criar `cmd/main.go` mínimo em cada serviço (HTTP servindo `/health`)
- [x] Configurar variáveis de ambiente (.env / config por serviço)
- [x] Definir padrão de abstração: interfaces de `services` e `repositories` no mesmo pacote, `handlers` dependendo apenas das interfaces
- [x] Rodar `docker compose up` e validar subida dos serviços

## Sprint 2 — Serviço de Estoque: Produto ✅

- [x] Model `Produto` (`internal/models`)
- [x] Migration da tabela `produtos`
- [x] Repository `ProdutoRepository` (`internal/repositories`)
- [x] Service com regras:
  - [x] Código único
  - [x] Saldo nunca negativo/inválido
- [x] Handlers e rotas:
  - [x] `POST /produtos`
  - [x] `GET /produtos`
  - [x] `GET /produtos/:id`
  - [x] `PUT /produtos/:id`
  - [x] `DELETE /produtos/:id`
- [x] Testes de unidade/integração do CRUD
- [x] Validação manual via API (ex.: curl/Postman)

## Sprint 3 — Serviço de Estoque: MovimentoEstoque ✅

> **Estratégia da trigger:** o saldo é atualizado pelo banco via trigger `trg_movimento_estoque_insert`
> (valida tipo/quantidade/produto/saldo e atualiza `produtos.saldo` na mesma transação, com
> `SELECT ... FOR UPDATE`). O Service valida antes (feedback rápido e testável) e a trigger é a
> garantia final no banco — inclusive contra SQL direto e concorrência (ver `docs/arquitetura.md`,
> Parte 3, e `migrations/0003_create_movimentos_trigger.up.sql`).

- [x] Model `MovimentoEstoque` (`internal/models`)
- [x] Migration da tabela `movimentos_estoque`
- [x] Operação `entrada` e `saida` no Service
- [x] Trigger no banco: valida saldo e atualiza `produtos.saldo` automaticamente (aplicada no startup)
- [x] Endpoint `POST /estoque/movimentos`
- [x] Saldo insuficiente → erro 409
- [x] Testes da operação de movimento
- [x] Validar que produto inexistente retorna 404
- [x] Validado com SQL direto no Postgres: trigger rejeita saída sem saldo e atualiza saldo

## Sprint 4 — Serviço de Faturamento: NotaFiscal e Itens ✅

> **Notas da implementação:** numeração sequencial via sequence do banco
> (`nota_fiscal_numero_seq`, criada no startup e na migration `0003`). Ao incluir item,
> o Faturamento consulta o Produto via HTTP no Estoque (404 se inexistente, 503 se o
> Estoque estiver indisponível) e grava snapshot de código/descrição.
> `total = quantidade × valor_unitario − desconto`.

- [x] Models `NotaFiscal` e `NotaFiscalItem` (`internal/models`)
- [x] Migrations das tabelas
- [x] Numeração sequencial da nota
- [x] API:
  - [x] `POST /notas` (status inicial `ABERTA`)
  - [x] `GET /notas`
  - [x] `GET /notas/:id`
  - [x] `POST /notas/:id/itens`
  - [x] `DELETE /notas/:id/itens/:item_id`
- [x] Regras:
  - [x] Só permite itens em nota `ABERTA`
  - [x] Snapshot de `codigo_produto` e `descricao_produto`
  - [x] Cálculo do `total` do item (quantidade × valor − desconto)
  - [x] Produto inexistente no Estoque não pode entrar na nota (consultar Estoque)
- [x] Testes
- [x] Validação manual via API (notas 1 e 2 sequenciais, item com total 25, 404 produto inexistente, remoção de item 204/404)

## Sprint 5 — Faturamento: NotaFiscalEvento ✅

**Objetivo:** auditoria do ciclo de vida da nota.

> **Notas da implementação:** eventos de auditoria persistidos em `nota_fiscal_eventos`
> (campo `dados` em JSONB). `NOTA_CRIADA`, `ITEM_ADICIONADO` e `ITEM_REMOVIDO` são emitidos
> automaticamente pelo `NotaFiscalService`; os eventos do fluxo de impressão
> (`BAIXA_ESTOQUE_SOLICITADA`, `ESTOQUE_BAIXADO`, `FALHA_ESTOQUE`, `NOTA_FECHADA`) já têm
> constantes definidas e serão registrados no Sprint 6. Registro é best-effort (falha é logada,
> não aborta a operação principal). Auditoria consultável via `GET /notas/:id/eventos`.

- [x] Model `NotaFiscalEvento` (`internal/models`)
- [x] Migration da tabela `nota_fiscal_eventos`
- [x] Service de eventos
- [x] Registrar eventos:
  - [x] `NOTA_CRIADA`
  - [x] `ITEM_ADICIONADO`
  - [x] `ITEM_REMOVIDO`
  - [ ] `BAIXA_ESTOQUE_SOLICITADA` (Sprint 6)
  - [ ] `ESTOQUE_BAIXADO` (Sprint 6)
  - [ ] `FALHA_ESTOQUE` (Sprint 6)
  - [ ] `NOTA_FECHADA` (Sprint 6)
- [x] Campo `dados` (jsonb) para payloads específicos
- [x] Testes
- [x] Validação manual via API (trilha NOTA_CRIADA → ITEM_ADICIONADO → ITEM_REMOVIDO com `dados`, 404 para nota inexistente, JSONB conferido direto no banco)

## Sprint 6 — Integração Faturamento ↔ Estoque (Fluxo HTTP) ✅

> **Notas da implementação:** `POST /notas/:id/imprimir` valida nota/status/itens, solicita baixa
> `SAIDA` por item via `POST /estoque/movimentos` (origem `FATURAMENTO`, `referencia` `nota-<id>` e
> `idempotency_key` `nota-<id>-item-<id>`) e só então fecha a nota. Estoque baixou saldo é conferido
> no banco do Estoque. Falha (Estoque indisponível → 503; saldo insuficiente → 409) deixa a nota
> `ABERTA`, registra `FALHA_ESTOQUE` e retorna feedback — cenário obrigatório de falha de
> microsserviço demonstrado (§15). Baixa é síncrona por item; falha no meio pode deixar baixa
> parcial (sem transação distribuída), recuperável por reenvio graças à `idempotency_key` (Sprint 8).
>
> **Nota (Sprint 7, Etapa 3):** o fluxo síncrono descrito acima foi substituído por assíncrono —
> `imprimir` agora publica `BaixaSolicitada` na fila e retorna 202 (ver Sprint 7).

- [x] Cliente HTTP no Faturamento para consumir a API do Estoque
- [x] Endpoint `POST /notas/:id/imprimir`
- [x] Fluxo:
  - [x] Validar nota e status (deve estar `ABERTA`)
  - [x] Identificar itens da nota
  - [x] Solicitar baixa ao Estoque
  - [x] Processar resultado
  - [x] Fechar nota (`FECHADA`) somente após sucesso
- [x] Bloqueio: nota não aberta não pode imprimir
- [x] Falha do Estoque (indisponível):
  - [x] Nota permanece `ABERTA`
  - [x] Registrar `FALHA_ESTOQUE`
  - [x] Retornar 503 com feedback ao usuário
- [x] Testes do fluxo feliz e do fluxo de falha
- [x] Demonstrar o cenário de falha de microsserviço (requisito obrigatório)
- [x] Validação manual via API: feliz (saldo 5→2, nota FECHADA, evento NOTA_FECHADA), 409 fechada, 409 saldo insuficiente, 503 com Estoque derrubado (nota ABERTA + FALHA_ESTOQUE)

## Sprint 7 — Mensageria (RabbitMQ)

**Objetivo:** desacoplar a baixa de estoque via fila.

> **Progresso incremental (etapas):** Etapa 1 ✅ — lib `amqp091-go`, `internal/broker` (Connect/Close)
> e `RABBITMQ_URL` nos dois serviços; demo `cmd/ping` publica/consome 3 mensagens. Etapa 2 ✅ —
> `internal/messaging` (contrato JSON + topologia) nos dois serviços: exchange `korp.baixa`
> (direct) com filas `korp.baixa.solicitada` (`baixa.solicitada`) e `korp.baixa.resultado`
> (`baixa.resultado`); testes de contrato travam o JSON; topologia declarada e conferida na UI.
> Etapa 3 ✅ — `BaixaPublisher` (interface + impl AMQP) no Faturamento; `Imprimir` monta
> `BaixaSolicitada`, publica na fila e retorna **202 Accepted** (a nota permanece `ABERTA`; em
> falha de publicação registra `FALHA_ESTOQUE` e retorna 503); wiring em `main.go`; testes e
> validação manual OK (mensagem `{"nota_id":11,"numero":11,"itens":[...]}` vista na fila).
> Etapa 4 ✅ — Estoque consome `BaixaSolicitada` e executa a baixa via `MovimentoService`
> (Tipo `SAIDA`, origem `FATURAMENTO`, ref `nota-<id>`, idempotency `nota-<id>-item-<produto>`);
> lógica pura `tratarMensagem` testável. Etapa 5 ✅ — `ResultadoBaixaPublisher` publica
> `ESTOQUE_BAIXADO` / `ESTOQUE_BAIXA_NEGADA` (+motivo) / `ESTOQUE_INDISPONIVEL` (após esgotar
> tentativas) em `korp.baixa.resultado`. Etapa 6 ✅ — Faturamento consome o resultado e
> `ProcessarResultadoBaixa`: `ESTOQUE_BAIXADO` fecha a nota (+`ESTOQUE_BAIXADO`/`NOTA_FECHADA`,
> idempotente), negada/indisponível registra `FALHA_ESTOQUE` e mantém `ABERTA`.
> Etapa 7 ✅ — retry com backoff exponencial (2s→30s, máx 5 tentativas) e DLX/DLQ
> (`korp.baixa.dlx`, filas `*.dlq`) para recuperação de falhas; testes unitários dos dois lados;
> validação manual completa (feliz, negada, indisponível com persistência da fila, DLQ). Guia de
> implementação: `docs/guia-sprint7-mensageria.md`.

- [x] Adicionar RabbitMQ ao `docker-compose.yml` (se ainda não estiver)
- [x] Configurar conexão AMQP nos dois serviços
- [x] Declarar exchanges/filas/rotas
- [x] Faturamento publica `ESTOQUE_BAIXA_SOLICITADA`
- [x] Estoque consome a mensagem e executa a baixa
- [x] Estoque publica resultado: `ESTOQUE_BAIXADO` / `ESTOQUE_BAIXA_NEGADA` / `ESTOQUE_INDISPONIVEL`
- [x] Faturamento consome o resultado e fecha a nota
- [x] Retry de consumo com backoff
- [x] Recuperação de falhas (dead-letter / reprocessamento)
- [x] Testes do fluxo assíncrono

## Sprint 8 — Robustez: Idempotência e Concorrência ✅

**Objetivo:** garantir que baixas não dupliquem e o saldo não estoure em acesso simultâneo.

- [x] `idempotency_key` no `MovimentoEstoque` (Sprint 3; consumidor já preenche desde o Sprint 7)
- [x] Índice único parcial `uq_movimentos_idempotency` (WHERE `idempotency_key <> ''`) via `ApplyMovimentoConstraints`
- [x] Estoque rejeita baixa duplicada para a mesma chave: `23505` → `ErrDuplicado` → `409 CONFLICT_DUPLICADO`
- [x] Consumidor trata duplicado como **sucesso idempotente** (redelivery/ack perdido não repete resultado nem baixa)
- [x] Concorrência garantida pela trigger com `SELECT ... FOR UPDATE` (transação no banco)
- [x] Cenário de teste: saldo 1, duas chamadas paralelas de quantidade 1
  - [x] Primeira baixa: sucesso (201)
  - [x] Segunda baixa: rejeitada (409 saldo insuficiente)
  - [x] Saldo final = 0
- [x] Testes unitários (mapeamento 23505, service 409, consumidor com redelivery total/parcial) + integração
- [x] Validação manual: HTTP duplicado (409 CONFLICT_DUPLICADO), concorrência paralela e redelivery Rabbit (2 mensagens idênticas → 1 movimento, 1 resultado, DLQ vazia)

> **Nota:** a trigger já cobria concorrência e validações (Sprint 3). O Sprint 8 enxuto adicionou apenas o que a fila _at-least-once_ exigia: o índice único e o tratamento idempotente do redelivery.

## Sprint 9 — Erros, Logs e Observabilidade ✅

**Objetivo:** qualidade de API e operação.

- [x] Formato de erro padrão `{ "error": { code, message, details } }` (Sprint 2/3, `apperrors`)
- [x] Status codes corretos:
  - [x] 400 dados inválidos
  - [x] 404 não encontrado
  - [x] 409 conflito (estoque/duplicidade/estado)
  - [x] 422 regra de negócio
  - [x] 500 erro inesperado
  - [x] 503 dependente indisponível
- [x] Logs estruturados JSON por níveis (`internal/logging`, nível por `LOG_LEVEL`) com serviço/versão/timestamp
- [x] Health checks reais nos dois serviços (`GET /health` verifica DB `PingContext` + RabbitMQ `IsClosed`; 200 ok / 503 degradado)
- [x] Configuração por ambiente (dev/prod via env vars)
- [x] Middleware Gin de log de requisições (método, path, status, duração)
- [x] Revisão final dos tratamentos de erro da API
- [x] Validação manual: `/health` 200 → DB derrubado → 503 degradado → DB de volta → 200

## Sprint 10 — Frontend Angular

**Objetivo:** interface para os fluxos de produtos, notas e impressão.

- [ ] Criar projeto Angular (`ng new`)
- [ ] Adicionar Angular Material
- [ ] Estrutura: componentes, services, guards, environment
- [ ] `HttpClient` + services para as APIs
- [ ] RxJS para fluxos reativos
- [ ] Tela de Produtos (listar, criar, editar, excluir)
- [ ] Tela de Notas (listar, criar, ver detalhes)
- [ ] Tela de Itens (adicionar/remover produtos na nota)
- [ ] Ação de Impressão:
  - [ ] Indicador de processamento
  - [ ] Feedback de sucesso/falha (503 do Estoque)
  - [ ] Atualização do status/saldo na tela
- [ ] Forms reativos + validações
- [ ] Tratamento de erros na UI

## Sprint 11 — Finalização e Entrega

**Objetivo:** validar requisitos e preparar a entrega.

- [ ] Rodar todos os testes (backend + frontend)
- [ ] Teste manual completo do fluxo: criar produto → criar nota → itens → imprimir → saldo atualizado
- [ ] Testar falha do Estoque e feedback ao usuário
- [ ] Testar idempotência e concorrência
- [ ] Finalizar `README.md`:
  - [ ] Como rodar (docker compose)
  - [ ] Endpoints documentados
  - [ ] Arquitetura resumida
  - [ ] Decisões técnicas justificadas
- [ ] Gravar vídeo demonstrativo (telas + funcionalidades)
- [ ] Revisar contra a seção 24 da SDD (requisitos obrigatórios)
- [ ] Publicar repositório público `Korp_Teste_SeuNome`

---

## Checkpoints obrigatórios (SDD §24)

Antes da entrega, confirmar:

- [ ] 2 microsserviços (Estoque e Faturamento) funcionando
- [ ] Produto com código, descrição e saldo
- [ ] Nota com numeração sequencial, status ABERTA/FECHADA e múltiplos produtos
- [ ] Impressão com indicador de processamento, fechamento da nota, bloqueio de nota não aberta e atualização do saldo
- [ ] Persistência real (PostgreSQL)
- [ ] Tratamento de falha entre microsserviços demonstrado

---

## Sprint Extra — Autenticação (Auth) ✅

> **Extra fora do escopo da SDD**: microsserviço de autenticação JWT independente
> com cadastro/login de usuários, tela de login no Angular e integração
> no header. Acesso geral (home, páginas) permanece SEM bloqueio por login
> (acesso público), mas o mecanismo está pronto para uso futuro.

### Backend Go — serviço `auth` (porta 8083)

- [x] Banco próprio: PostgreSQL `auth_db` (container `korp-postgres-auth`, porta `15434:5432`)
- [x] Volume declarado no `docker-compose.yml`: `pg_auth_data`
- [x] Model `Usuario` (`id`, `nome`, `email` UNIQUE, `senha_hash` bcrypt, `ativo`)
- [x] Migration `0001_create_usuarios.up.sql` (tabela + índice)
- [x] Repository pattern com interface: `Criar`, `BuscarPorEmail`, `BuscarPorID`, `Atualizar`, `Contar`
- [x] Service:
  - [x] `Criar()` — valida email/senha obrigatórios, gera hash bcrypt (`golang.org/x/crypto/bcrypt`)
  - [x] `Autenticar()` — busca por email, verifica ativo, compara hash bcrypt
  - [x] `CriarSeVazio()` — idempotente: só insere se tabela vazia (usado no seed)
  - [x] Erros tipados: `ErrEmailObrigatorio`, `ErrSenhaObrigatoria`, `ErrUsuarioInativo`, `ErrCredencialInvalida`
- [x] JWT (`github.com/golang-jwt/jwt/v5`):
  - [x] `JWTService` (segredo + duração configuráveis por env)
  - [x] `ValidateJWTSecret()` — valida segredo ≥ 16 caracteres antes de iniciar
  - [x] `Generate(userID, email)` → claims `{sub, email}`, HS256, `IssuedAt` + `ExpiresAt` (padrão 1h)
  - [x] `Validate(tokenString)` → parse com HMAC, retorna `Claims` preenchidos ou erro
- [x] Handlers (`UsuarioHandler`):
  - [x] `POST /auth/usuarios` — cadastro (`nome`, `email`, `senha` ≥ 6)
  - [x] `POST /auth/login` → `{ access_token, token_type: "Bearer", user: {id,nome,email} }`
  - [x] Tratamento de erro: `400` (dados inválidos), `401` (creds/usuário inativo), `500` (interno)
- [x] Config (por env com fallback):
  - [x] `PORT=8083`, `DB_HOST=localhost`, `DB_PORT=15434`, `DB_NAME=auth_db`, `DB_USER=auth`, `DB_PASSWORD=auth`
  - [x] `JWT_SECRET="korp-teste-super-secret-key-2026"` (fallback; ≥ 16 chars)
  - [x] `JWT_EXPIRATION=1h` (suporta `15m`, `24h`, etc via `time.ParseDuration`)
- [x] CORS liberado para `http://localhost:4200` (methods: GET/POST/PUT/DELETE/OPTIONS; headers: Authorization)
- [x] **Seed automático no startup**: se a tabela `usuarios` estiver vazia, cria o admin padrão:
  - **Email**: `admin@korp.local`
  - **Senha**: `korp26`
  - **Nome**: `Administrador Korp`
  - Log impresso no startup com id/credenciais.
- [x] Rotas finais:
  - `GET  /health` — health check
  - `POST /auth/usuarios` — criar usuário
  - `POST /auth/login` — login → JWT

### Frontend Angular — Autenticação

- [x] `AuthService` (`src/app/services/auth.service.ts`):
  - [x] `login(email, senha)` → POST `http://localhost:8083/auth/login`
  - [x] Após sucesso: salva `korp_token` (JWT) e `korp_user` no `localStorage`
  - [x] Helpers públicos: `getToken()`, `getUser()`, `isLoggedIn()`, `logout()`
- [x] `LoginComponent` (`src/app/pages/login/`):
  - [x] Formulário template-driven com `ngModel` (email + senha)
  - [x] Estados `loading` (botão desabilita) e `erroMsg` (exibe msg em box vermelho)
  - [x] UI responsiva: inputs com foco azul, botão azul (`#007bff`), hover e disabled
  - [x] Sucesso → `router.navigate(['/home'])`
  - [x] Erro → exibe `err.error.error` ou mensagem padrão
- [x] `AppHeaderComponent` (header público):
  - [x] **Desktop**: mostra botão "Login" (ícone + texto) quando deslogado; quando logado: `{nome_usuario}` + botão ícone `logout`
  - [x] **Mobile** (menu hambúrguer): item "Login" quando deslogado; separador + item "Sair" quando logado
  - [x] Clique em Sair → `auth.logout()` + rota `/login`
  - [x] `MatDividerModule` importado para separador do menu mobile
- [x] **Acesso público (sem guard)**:
  - [x] Rota raiz `""` redireciona → `/home` (não bloqueia)
  - [x] Rota curinga `"**"` também cai em `/home`
  - [x] Todas as páginas (home, produtos, clientes, notas) acessíveis sem token
  - [x] Mecanismo de auth está **pronto** para receber `canActivate` guards e `HttpInterceptor` Bearer futuramente.

### Usuário padrão (seed do primeiro start)

| Campo  | Valor              |
| ------ | ------------------ |
| Nome   | Administrador Korp |
| Email  | admin@korp.local   |
| Senha  | `korp26`           |
| Ativo? | Sim                |
