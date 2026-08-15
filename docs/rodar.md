# Como rodar o projeto

Guia para subir a infraestrutura (banco + mensageria) e os microsserviços localmente.

## Pré-requisitos

- **Docker Desktop** (com Docker Compose) — para Postgres e RabbitMQ
- **Go 1.26+** — para os microsserviços
- **Node.js 22+ e npm** — para o frontend Angular (Sprint 10)

## Resumo rápido (sequência completa de subida)

```bash
# 1. infra (4 containers: estoque, faturamento, auth DB + RabbitMQ)
docker compose up -d
docker compose ps              # esperar todos "healthy"

PS C:\Users\leoka\Korp_Teste_LeonardoAraujo> docker compose up -d
[+] up 4/4
 ✔ Container korp-postgres-auth        Started                                            0.9s
 ✔ Container korp-postgres-faturamento Started                                            1.0s
 ✔ Container korp-postgres-estoque     Started                                            1.0s
 ✔ Container korp-rabbitmq             Started                                            1.0s
PS C:\Users\leoka\Korp_Teste_LeonardoAraujo>



# 2. microsserviços (3 terminais separados)
cd backendGo/services/estoque      ; go run ./cmd   # porta 8081
cd backendGo/services/faturamento  ; go run ./cmd   # porta 8082
cd backendGo/services/auth         ; go run ./cmd   # porta 8083 (cria usuário admin na 1a vez)

# 3. health
curl http://localhost:8081/health   # estoque
curl http://localhost:8082/health   # faturamento
curl http://localhost:8083/health   # auth

# 4. frontend (opcional)
cd frontend ; npm install       # apenas na 1a vez
cd frontend ; npm start         # dev proxy em 4200
```

---

## 1. Infraestrutura — Docker Compose

Sobe 4 containers:

| Container                   | Imagem                         | Porta host                   | Usuário/Senha             | DB               |
| --------------------------- | ------------------------------ | ---------------------------- | ------------------------- | ---------------- |
| `korp-postgres-estoque`     | `postgres:16-alpine`           | `15432`                      | `estoque/estoque`         | `estoque_db`     |
| `korp-postgres-faturamento` | `postgres:16-alpine`           | `15433`                      | `faturamento/faturamento` | `faturamento_db` |
| `korp-postgres-auth`        | `postgres:16-alpine`           | `15434`                      | `auth/auth`               | `auth_db`        |
| `korp-rabbitmq`             | `rabbitmq:3-management-alpine` | `5672` (AMQP) + `15672` (UI) | `korp/korp`               | —                |

Comando:

```bash
docker compose up -d
docker compose ps
```

Aguardar o campo `STATUS` de todos aparecer `(healthy)`. Em caso de problema, use `docker compose logs -f nome-container`.

---

## 2. Backend — microsserviços

Rode cada serviço em um **terminal separado**:

### 2.1 Estoque (porta 8081)

```bash
cd backendGo/services/estoque
go run ./cmd
```

- Migration auto-aplicada (`migrations/*.up.sql`) no startup via `RunMigrations`.
- Trigger `trg_movimento_estoque_insert` instalada (atualiza saldo automaticamente).

### 2.2 Faturamento (porta 8082)

```bash
cd backendGo/services/faturamento
go run ./cmd
```

- Sequence `nota_fiscal_numero_seq` criada pela migration (numeração automática de notas).
- Consome/publica em filas do RabbitMQ após impressão.

### 2.3 Auth — Autenticação (porta 8083) ✅ EXTRA

```bash
cd backendGo/services/auth
go run ./cmd
```

- Migration `0001_create_usuarios.up.sql` aplicada no startup.
- **Seed automático (idempotente)**: se a tabela `usuarios` estiver vazia, cria o admin padrão:
  - **Email**: `admin@korp.local`
  - **Senha**: `korp26`
  - **Nome**: `Administrador Korp`
- Log no startup informa se o seed foi aplicado ou pulado.
- JWT: `HS256`, expira em `1h`. Fallback `JWT_SECRET = "korp-teste-super-secret-key-2026"` (dev).

#### Endpoints do Auth

| Método | Rota             | Corpo de entrada                                         | Descrição                                                             |
| ------ | ---------------- | -------------------------------------------------------- | --------------------------------------------------------------------- |
| GET    | `/health`        | —                                                        | Health check do serviço auth                                          |
| POST   | `/auth/usuarios` | `{ "nome": "...", "email": "...", "senha": "≥6 chars" }` | Cria um novo usuário cadastrado                                       |
| POST   | `/auth/login`    | `{ "email": "...", "senha": "..." }`                     | Retorna `{ access_token, token_type:"Bearer", user:{id,nome,email} }` |

Exemplo com curl do **login admin padrão**:

```bash
curl -X POST http://localhost:8083/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{ \"email\": \"admin@korp.local\", \"senha\": \"korp26\" }"
```

#### Validar o JWT retornado (opcional)

Cole o `access_token` em https://jwt.io — o payload terá `sub` (UserID), `email`, `iat` (issued at) e `exp` (expiração).

---

## 3. Frontend Angular (porta 4200)

```bash
cd frontend
npm install        # 1a vez (instala Angular + Material)
npm start          # dev server com proxy para 8081/8082
```

Abra http://localhost:4200

### Proxy (dev)

O `proxy.conf.json` reescreve `/api/estoque/*` → `http://localhost:8081/*` e `/api/faturamento/*` → `http://localhost:8082/*` (PathRewrite remove o prefixo `/api/...`).

### Login no Frontend

Acesse **http://localhost:4200/login** e entre com as credenciais do admin:

- Email: `admin@korp.local`
- Senha: `korp26`
