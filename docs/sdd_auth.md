# sdd_auth.md — Autenticação e acesso

auth
├── PostgreSQL próprio
├── Gin
├── GORM
├── JWT RS256
├── bcrypt
├── middleware
└── API REST


Angular
   │
   │ POST /auth/login
   ▼
 Auth Service
   │
   ├── busca usuário
   ├── valida senha com bcrypt
   └── assina JWT com chave privada
             │
             ▼
          Angular
             │
             │ Bearer JWT
       ┌─────┴─────┐
       ▼           ▼
  Faturamento   Estoque
       │           │
       └── valida com chave pública


postgres-auth
     │
     └── auth_db
           │
           └── usuarios


bibliotecas utilizadas:

go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get golang.org/x/crypto
go get github.com/golang-jwt/jwt/v5

Implementado configuração e docker-compose para auth:
postgres-auth:
  image: postgres:16-alpine
  container_name: korp-postgres-auth
  environment:
    POSTGRES_USER: auth
    POSTGRES_PASSWORD: auth
    POSTGRES_DB: auth_db
  ports:
    - "15434:5432"
  volumes:
    - pg_auth_data:/var/lib/postgresql/data
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U auth -d auth_db"]
    interval: 5s
    timeout: 5s
    retries: 10



