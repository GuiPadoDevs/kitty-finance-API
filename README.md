# 🎀 Finanças da Sah — API Backend (Go / Golang)

API RESTful em Go com PostgreSQL para gerenciamento de finanças pessoais com ciclos financeiros customizados, categorização com metas e autenticação JWT.

---

## 🌸 Tecnologias

- **Go (Golang 1.23+)**
- **go-chi/chi/v5** (Roteador HTTP rápido e idiomático)
- **jackc/pgx/v5** (Driver de alta performance com pool de conexões PostgreSQL)
- **golang-jwt/jwt/v5** (Emissão e validação de tokens JWT)
- **golang.org/x/crypto/bcrypt** (Hashing de senhas)
- **PostgreSQL 16** (Banco relacional)

---

## 🚀 Como Executar Localmente

1. **Subir o banco PostgreSQL:**
   ```bash
   docker compose up -d
   ```

2. **Configurar variáveis de ambiente:**
   ```bash
   cp .env.example .env
   ```

3. **Executar a API Go:**
   ```bash
   go run cmd/api/main.go
   ```
   A API iniciará em: `http://localhost:8080`

---

## ☁️ Deploy na Nuvem (Render / Railway / Fly.io)

### Opção 1: Render.com (Recomendado & 100% Grátis)
1. Crie um **Web Service** no [Render](https://render.com) e conecte este repositório (`kitty-finance-API`).
2. Configurações de Build & Start:
   - **Runtime:** Go (ou Docker)
   - **Build Command:** `go build -o api cmd/api/main.go`
   - **Start Command:** `./api`
3. Adicione as variáveis de ambiente:
   - `PORT`: `8080` (ou a porta padrão do Render)
   - `DATABASE_URL`: `postgres://user:pass@ep-host.neon.tech/financas_sah?sslmode=require` (seu banco Neon)
   - `JWT_SECRET`: `sua-chave-secreta-forte-com-mais-de-32-caracteres`
4. Clique em **Create Web Service**! As migrações SQL são executadas automaticamente na inicialização.

---

## 📋 Endpoints Principais

- `GET /api/v1/health` - Status da API e conexão com banco
- `POST /api/v1/auth/register` - Cadastro de usuária
- `POST /api/v1/auth/login` - Login e emissão de JWT
- `GET /api/v1/auth/me` - Dados da conta e preferências
- `GET /api/v1/cycles/current` - Ciclo financeiro aberto
- `POST /api/v1/cycles/close` - Fechamento manual do mês com snapshot
- `GET /api/v1/categories` - Listagem de tópicos/categorias
- `POST /api/v1/transactions` - Registro de receita/despesa
- `GET /api/v1/dashboard/summary` - Resumo financeiro consolidado
